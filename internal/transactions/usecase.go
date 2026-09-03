package transactions

import (
	"context"
	"errors"

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

func (u *UseCase) Paginate(ctx context.Context, req PaginateRequest) ([]Transaction, int64, error) {
	offset := (req.Page - 1) * req.Limit

	filter := PaginateFilter{
		PaymentStatus:  req.PaymentStatus,
		DeliveryStatus: req.DeliveryStatus,
		DateFrom:       req.DateFrom,
		DateTo:         req.DateTo,
	}

	return u.repo.FindPaginated(ctx, filter, req.Limit, offset)
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
