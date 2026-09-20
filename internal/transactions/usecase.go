package transactions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/BernardBerenes/SupplyHub-API/internal/stores"
)

var (
	ErrNotFound              = errors.New("transaction not found")
	ErrStoreNotFound         = errors.New("store not found")
	ErrDuplicatePending      = errors.New("store already has a pending transaction on this date")
	ErrStoreUpdateNotPending = errors.New("store can only be changed while the transaction is pending")
)

type StoreLookup interface {
	FindByID(ctx context.Context, id int64) (*stores.Store, error)
	FindByIDIncludingDeleted(ctx context.Context, id int64) (*stores.Store, error)
}

type UseCase struct {
	repo  Repository
	store StoreLookup
}

func NewUseCase(repo Repository, store StoreLookup) *UseCase {
	return &UseCase{
		repo:  repo,
		store: store,
	}
}

func (u *UseCase) Create(ctx context.Context, input CreateInput) error {
	store, err := u.store.FindByID(ctx, input.StoreID)
	if err != nil {
		return err
	}
	if store == nil {
		return ErrStoreNotFound
	}

	conflict, err := u.repo.ExistsPendingByStoreAndDate(ctx, input.StoreID, input.Date)
	if err != nil {
		return err
	}
	if conflict {
		return ErrDuplicatePending
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	transaction := &Transaction{
		ID:             id.String(),
		Store:          StoreSnapshot{ID: store.ID, Name: store.Name},
		PaymentStatus:  PAYMENT_STATUS_UNPAID,
		DeliveryStatus: DELIVERY_STATUS_PENDING,
		Date:           input.Date,
	}

	return u.repo.Create(ctx, transaction)
}

func (u *UseCase) Paginate(ctx context.Context, req PaginateRequest) ([]Transaction, map[string]int64, int64, error) {
	offset := (req.Page - 1) * req.Limit

	filter := PaginateFilter{
		PaymentStatus:  req.PaymentStatus,
		DeliveryStatus: req.DeliveryStatus,
		DateFrom:       req.DateFrom,
		DateTo:         req.DateTo,
	}

	transactions, total, err := u.repo.FindPaginated(ctx, filter, req.Limit, offset)
	if err != nil {
		return nil, nil, 0, err
	}

	ids := make([]string, len(transactions))
	for i, t := range transactions {
		ids[i] = t.ID
	}

	totalPrices, err := u.repo.SumTotalPriceByTransactionIDs(ctx, ids)
	if err != nil {
		return nil, nil, 0, err
	}

	return transactions, totalPrices, total, nil
}

func (u *UseCase) Update(ctx context.Context, id string, input UpdateInput) error {
	existing, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}

	updates := map[string]interface{}{}

	if input.StoreID != nil {
		if existing.DeliveryStatus != DELIVERY_STATUS_PENDING {
			return ErrStoreUpdateNotPending
		}

		store, err := u.store.FindByID(ctx, *input.StoreID)
		if err != nil {
			return err
		}
		if store == nil {
			return ErrStoreNotFound
		}

		updates["store"] = StoreSnapshot{ID: store.ID, Name: store.Name}
	}

	if input.PaymentStatus != nil {
		updates["payment_status"] = *input.PaymentStatus
	}
	if input.DeliveryStatus != nil {
		updates["delivery_status"] = *input.DeliveryStatus
	}
	if input.Date != nil {
		updates["date"] = *input.Date
	}

	if len(updates) == 0 {
		return nil
	}

	return u.repo.Update(ctx, id, updates)
}

// Exists and FindAllPendingIDs let other modules (e.g. transaction details)
// depend on this module's Usecase instead of its Repository.
func (u *UseCase) Exists(ctx context.Context, id string) (bool, error) {
	transaction, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return false, err
	}
	return transaction != nil, nil
}

func (u *UseCase) FindAllPendingIDs(ctx context.Context) ([]string, error) {
	pending, err := u.repo.FindAllPending(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(pending))
	for i, transaction := range pending {
		ids[i] = transaction.ID
	}

	return ids, nil
}

func (u *UseCase) Delete(ctx context.Context, id string) error {
	rows, err := u.repo.SoftDelete(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (u *UseCase) Revenue(ctx context.Context, req RevenueRequest) (RevenueResponse, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	var from, to time.Time
	to = today

	if req.DateFrom != "" || req.DateTo != "" {
		if req.DateFrom != "" {
			from, _ = time.Parse(DateFormat, req.DateFrom)
		}
		if req.DateTo != "" {
			to, _ = time.Parse(DateFormat, req.DateTo)
		}
		if req.DateFrom == "" {
			earliest, err := u.repo.FindEarliestDate(ctx)
			if err != nil {
				return RevenueResponse{}, err
			}
			if earliest != nil {
				from = *earliest
			} else {
				from = to
			}
		}
	} else {
		switch req.Period {
		case PERIOD_1D:
			from = today
		case PERIOD_1M:
			from = today.AddDate(0, -1, 0)
		case PERIOD_3M:
			from = today.AddDate(0, -3, 0)
		case PERIOD_6M:
			from = today.AddDate(0, -6, 0)
		case PERIOD_1Y:
			from = today.AddDate(-1, 0, 0)
		case PERIOD_ALL:
			earliest, err := u.repo.FindEarliestDate(ctx)
			if err != nil {
				return RevenueResponse{}, err
			}
			if earliest != nil {
				from = *earliest
			} else {
				from = today
			}
		default:
			from = today.AddDate(0, -1, 0)
		}
	}

	groupBy := req.GroupBy
	if groupBy == "" {
		groupBy = autoGroupBy(from, to)
	}

	filter := RevenueFilter{
		DateFrom: from.Format(DateFormat),
		DateTo:   to.Format(DateFormat),
		GroupBy:  groupBy,
	}

	buckets, err := u.repo.SumRevenueByBucket(ctx, filter)
	if err != nil {
		return RevenueResponse{}, err
	}

	revenueByPeriod := make(map[string]int64, len(buckets))
	var total int64
	for _, b := range buckets {
		revenueByPeriod[b.Period] = b.Revenue
		total += b.Revenue
	}

	points := buildPoints(from, to, groupBy, revenueByPeriod)

	counts, err := u.repo.CountStatusesInRange(ctx, filter.DateFrom, filter.DateTo)
	if err != nil {
		return RevenueResponse{}, err
	}

	return RevenueResponse{
		Period:            req.Period,
		GroupBy:           groupBy,
		DateFrom:          filter.DateFrom,
		DateTo:            filter.DateTo,
		TotalRevenue:      total,
		Points:            points,
		TransactionCount:  counts.PaidCount + counts.UnpaidCount,
		PaidCount:         counts.PaidCount,
		UnpaidCount:       counts.UnpaidCount,
		PendingDeliveries: counts.PendingDeliveries,
		OnDelivery:        counts.OnDelivery,
		DeliveredCount:    counts.DeliveredCount,
	}, nil
}

func autoGroupBy(from, to time.Time) string {
	days := int(to.Sub(from).Hours() / 24)
	switch {
	case days <= 1:
		return GROUP_BY_TOTAL
	case days <= 62:
		return GROUP_BY_DAY
	default:
		return GROUP_BY_MONTH
	}
}

func buildPoints(from, to time.Time, groupBy string, revenueByPeriod map[string]int64) []RevenuePoint {
	points := []RevenuePoint{}

	switch groupBy {
	case GROUP_BY_TOTAL:
		key := "total"
		points = append(points, RevenuePoint{Period: key, Revenue: revenueByPeriod[key]})
	case GROUP_BY_DAY:
		for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
			key := d.Format(DateFormat)
			points = append(points, RevenuePoint{Period: key, Revenue: revenueByPeriod[key]})
		}
	case GROUP_BY_WEEK:
		seen := map[string]bool{}
		for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
			year, week := d.ISOWeek()
			key := fmt.Sprintf("%04d-W%02d", year, week)
			if seen[key] {
				continue
			}
			seen[key] = true
			points = append(points, RevenuePoint{Period: key, Revenue: revenueByPeriod[key]})
		}
	case GROUP_BY_MONTH:
		start := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, time.UTC)
		for d := start; !d.After(end); d = d.AddDate(0, 1, 0) {
			key := d.Format("2006-01")
			points = append(points, RevenuePoint{Period: key, Revenue: revenueByPeriod[key]})
		}
	}

	return points
}

func (u *UseCase) SyncStoreNames(ctx context.Context) error {
	pending, err := u.repo.FindAllPending(ctx)
	if err != nil {
		return err
	}

	for _, transaction := range pending {
		store, err := u.store.FindByIDIncludingDeleted(ctx, transaction.Store.ID)
		if err != nil {
			return err
		}
		if store == nil || store.Name == transaction.Store.Name {
			continue
		}

		updates := map[string]interface{}{
			"store": StoreSnapshot{ID: store.ID, Name: store.Name},
		}

		if err := u.repo.Update(ctx, transaction.ID, updates); err != nil {
			return err
		}
	}

	return nil
}
