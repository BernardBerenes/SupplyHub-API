package transactions

import (
	"context"
	"errors"
	"strconv"
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
		Where("store->>'id' = ?", strconv.FormatInt(storeID, 10)).
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
