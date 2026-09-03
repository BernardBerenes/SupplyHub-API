package transactiondetails

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, detail *TransactionDetail) error
	FindByID(ctx context.Context, transactionID, id string) (*TransactionDetail, error)
	FindPaginatedByTransactionID(ctx context.Context, transactionID string, limit, offset int) ([]TransactionDetail, int64, error)
	FindActiveByTransactionID(ctx context.Context, transactionID string) ([]TransactionDetail, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	SoftDelete(ctx context.Context, transactionID, id string) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, detail *TransactionDetail) error {
	return r.db.
		WithContext(ctx).
		Table("transaction_details").
		Create(detail).
		Error
}

func (r *repository) FindByID(ctx context.Context, transactionID, id string) (*TransactionDetail, error) {
	var detail TransactionDetail

	err := r.db.
		WithContext(ctx).
		Table("transaction_details").
		Where("id = ?", id).
		Where("transaction_id = ?", transactionID).
		Where("deleted_at IS NULL").
		First(&detail).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &detail, nil
}

func (r *repository) FindPaginatedByTransactionID(ctx context.Context, transactionID string, limit, offset int) ([]TransactionDetail, int64, error) {
	var details []TransactionDetail
	var total int64

	base := r.db.
		WithContext(ctx).
		Table("transaction_details").
		Where("transaction_id = ?", transactionID).
		Where("deleted_at IS NULL")

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := base.
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&details).
		Error

	return details, total, err
}

func (r *repository) FindActiveByTransactionID(ctx context.Context, transactionID string) ([]TransactionDetail, error) {
	var details []TransactionDetail

	err := r.db.
		WithContext(ctx).
		Table("transaction_details").
		Where("transaction_id = ?", transactionID).
		Where("deleted_at IS NULL").
		Find(&details).
		Error

	return details, err
}

func (r *repository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	return r.db.
		WithContext(ctx).
		Table("transaction_details").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Updates(updates).
		Error
}

func (r *repository) SoftDelete(ctx context.Context, transactionID, id string) (int64, error) {
	result := r.db.
		WithContext(ctx).
		Table("transaction_details").
		Where("id = ?", id).
		Where("transaction_id = ?", transactionID).
		Where("deleted_at IS NULL").
		Update("deleted_at", time.Now())

	return result.RowsAffected, result.Error
}
