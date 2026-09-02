package stores

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, store *Store) error
	FindActive(ctx context.Context, name string) ([]Store, error)
	FindByID(ctx context.Context, id int64) (*Store, error)
	FindPaginated(ctx context.Context, name string, limit, offset int) ([]Store, int64, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	SoftDelete(ctx context.Context, id int64) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, store *Store) error {
	return r.db.
		WithContext(ctx).
		Table("stores").
		Create(store).
		Error
}

func (r *repository) FindActive(ctx context.Context, name string) ([]Store, error) {
	var stores []Store

	err := r.filterByName(ctx, name).
		Select("id", "name").
		Order("name ASC").
		Find(&stores).
		Error

	return stores, err
}

func (r *repository) FindByID(ctx context.Context, id int64) (*Store, error) {
	var store Store

	err := r.db.
		WithContext(ctx).
		Table("stores").
		Select("id", "name").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		First(&store).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &store, nil
}

func (r *repository) FindPaginated(ctx context.Context, name string, limit, offset int) ([]Store, int64, error) {
	var stores []Store
	var total int64

	if err := r.filterByName(ctx, name).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.filterByName(ctx, name).
		Select("id", "name").
		Order("name ASC").
		Limit(limit).
		Offset(offset).
		Find(&stores).
		Error

	return stores, total, err
}

func (r *repository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	return r.db.
		WithContext(ctx).
		Table("stores").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Updates(updates).
		Error
}

func (r *repository) SoftDelete(ctx context.Context, id int64) (int64, error) {
	result := r.db.
		WithContext(ctx).
		Table("stores").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Update("deleted_at", time.Now())

	return result.RowsAffected, result.Error
}

func (r *repository) filterByName(ctx context.Context, name string) *gorm.DB {
	query := r.db.
		WithContext(ctx).
		Table("stores").
		Where("deleted_at IS NULL")

	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	return query
}
