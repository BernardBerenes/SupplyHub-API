package transactions

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type PaginateFilter struct {
	PaymentStatus  string
	DeliveryStatus string
	DateFrom       string
	DateTo         string
}

type Repository interface {
	Create(ctx context.Context, transaction *Transaction) error
	FindByID(ctx context.Context, id string) (*Transaction, error)
	FindPaginated(ctx context.Context, filter PaginateFilter, limit, offset int) ([]Transaction, int64, error)
	ExistsPendingByStoreAndDate(ctx context.Context, storeID int64, date time.Time) (bool, error)
	FindAllPending(ctx context.Context) ([]Transaction, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	SoftDelete(ctx context.Context, id string) (int64, error)
	SumRevenueByBucket(ctx context.Context, filter RevenueFilter) ([]RevenueBucket, error)
	FindEarliestDate(ctx context.Context) (*time.Time, error)
	SumTotalPriceByTransactionIDs(ctx context.Context, transactionIDs []string) (map[string]int64, error)
	CountStatusesInRange(ctx context.Context, dateFrom, dateTo string) (StatusCounts, error)
	SumRevenueByStore(ctx context.Context, dateFrom, dateTo string) ([]StoreRevenue, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, transaction *Transaction) error {
	return r.db.
		WithContext(ctx).
		Table("transactions").
		Create(transaction).
		Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*Transaction, error) {
	var transaction Transaction

	err := r.db.
		WithContext(ctx).
		Table("transactions").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		First(&transaction).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &transaction, nil
}

func (r *repository) FindPaginated(ctx context.Context, filter PaginateFilter, limit, offset int) ([]Transaction, int64, error) {
	var transactions []Transaction
	var total int64

	if err := r.filtered(ctx, filter).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.filtered(ctx, filter).
		Order("date DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).
		Error

	return transactions, total, err
}

func (r *repository) ExistsPendingByStoreAndDate(ctx context.Context, storeID int64, date time.Time) (bool, error) {
	var count int64

	err := r.db.
		WithContext(ctx).
		Table("transactions").
		Where("deleted_at IS NULL").
		Where("delivery_status = ?", DELIVERY_STATUS_PENDING).
		Where("(store->>'id')::bigint = ?", storeID).
		Where("date = ?", date.Format(DateFormat)).
		Count(&count).
		Error

	return count > 0, err
}

func (r *repository) FindAllPending(ctx context.Context) ([]Transaction, error) {
	var transactions []Transaction

	err := r.db.
		WithContext(ctx).
		Table("transactions").
		Where("deleted_at IS NULL").
		Where("delivery_status = ?", DELIVERY_STATUS_PENDING).
		Find(&transactions).
		Error

	return transactions, err
}

func (r *repository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	return r.db.
		WithContext(ctx).
		Table("transactions").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Updates(updates).
		Error
}

func (r *repository) SoftDelete(ctx context.Context, id string) (int64, error) {
	result := r.db.
		WithContext(ctx).
		Table("transactions").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Update("deleted_at", time.Now())

	return result.RowsAffected, result.Error
}

func (r *repository) SumRevenueByBucket(ctx context.Context, filter RevenueFilter) ([]RevenueBucket, error) {
	var buckets []RevenueBucket

	var bucketExpr string
	switch filter.GroupBy {
	case GROUP_BY_DAY:
		bucketExpr = `to_char(t.date, 'YYYY-MM-DD')`
	case GROUP_BY_WEEK:
		bucketExpr = `to_char(t.date, 'IYYY-"W"IW')`
	case GROUP_BY_MONTH:
		bucketExpr = `to_char(t.date, 'YYYY-MM')`
	default:
		bucketExpr = `'total'`
	}

	err := r.db.
		WithContext(ctx).
		Table("transactions AS t").
		Select(bucketExpr+" AS period, COALESCE(SUM(d.total_price), 0) AS revenue").
		Joins("JOIN transaction_details AS d ON d.transaction_id = t.id AND d.deleted_at IS NULL").
		Where("t.deleted_at IS NULL").
		Where("t.payment_status = ?", PAYMENT_STATUS_PAID).
		Where("t.date >= ?", filter.DateFrom).
		Where("t.date <= ?", filter.DateTo).
		Group("period").
		Order("period ASC").
		Scan(&buckets).
		Error

	return buckets, err
}

func (r *repository) SumTotalPriceByTransactionIDs(ctx context.Context, transactionIDs []string) (map[string]int64, error) {
	totals := make(map[string]int64, len(transactionIDs))
	if len(transactionIDs) == 0 {
		return totals, nil
	}

	var rows []struct {
		TransactionID string
		TotalPrice    int64
	}

	err := r.db.
		WithContext(ctx).
		Table("transaction_details").
		Select("transaction_id, COALESCE(SUM(total_price), 0) AS total_price").
		Where("transaction_id IN ?", transactionIDs).
		Where("deleted_at IS NULL").
		Group("transaction_id").
		Scan(&rows).
		Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		totals[row.TransactionID] = row.TotalPrice
	}

	return totals, nil
}

func (r *repository) CountStatusesInRange(ctx context.Context, dateFrom, dateTo string) (StatusCounts, error) {
	var counts StatusCounts

	selectExpr := `
		COUNT(*) FILTER (WHERE payment_status = ?) AS paid_count,
		COUNT(*) FILTER (WHERE payment_status = ?) AS unpaid_count,
		COUNT(*) FILTER (WHERE delivery_status = ?) AS pending_deliveries,
		COUNT(*) FILTER (WHERE delivery_status = ?) AS on_delivery,
		COUNT(*) FILTER (WHERE delivery_status = ?) AS delivered_count
	`

	err := r.db.
		WithContext(ctx).
		Table("transactions").
		Select(selectExpr, PAYMENT_STATUS_PAID, PAYMENT_STATUS_UNPAID, DELIVERY_STATUS_PENDING, DELIVERY_STATUS_ON_DELIVERY, DELIVERY_STATUS_DELIVERED).
		Where("deleted_at IS NULL").
		Where("date >= ?", dateFrom).
		Where("date <= ?", dateTo).
		Scan(&counts).
		Error

	return counts, err
}

func (r *repository) SumRevenueByStore(ctx context.Context, dateFrom, dateTo string) ([]StoreRevenue, error) {
	var rows []StoreRevenue

	err := r.db.
		WithContext(ctx).
		Table("stores AS s").
		Select("s.id AS store_id, s.name AS store_name, COALESCE(SUM(d.total_price), 0) AS revenue").
		Joins(`LEFT JOIN transactions AS t ON (t.store->>'id')::bigint = s.id
			AND t.deleted_at IS NULL
			AND t.payment_status = ?
			AND t.date >= ?
			AND t.date <= ?`, PAYMENT_STATUS_PAID, dateFrom, dateTo).
		Joins("LEFT JOIN transaction_details AS d ON d.transaction_id = t.id AND d.deleted_at IS NULL").
		Where("s.deleted_at IS NULL").
		Group("s.id, s.name").
		Order("revenue DESC, s.name ASC").
		Scan(&rows).
		Error

	return rows, err
}

func (r *repository) FindEarliestDate(ctx context.Context) (*time.Time, error) {
	var earliest *time.Time

	err := r.db.
		WithContext(ctx).
		Table("transactions").
		Where("deleted_at IS NULL").
		Select("MIN(date)").
		Scan(&earliest).
		Error

	return earliest, err
}

func (r *repository) filtered(ctx context.Context, filter PaginateFilter) *gorm.DB {
	query := r.db.
		WithContext(ctx).
		Table("transactions").
		Where("deleted_at IS NULL")

	if filter.PaymentStatus != "" {
		query = query.Where("payment_status = ?", filter.PaymentStatus)
	}
	if filter.DeliveryStatus != "" {
		query = query.Where("delivery_status = ?", filter.DeliveryStatus)
	}
	if filter.DateFrom != "" {
		query = query.Where("date >= ?", filter.DateFrom)
	}
	if filter.DateTo != "" {
		query = query.Where("date <= ?", filter.DateTo)
	}

	return query
}
