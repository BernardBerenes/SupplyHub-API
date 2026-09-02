package products

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, product *Product) error
	FindActive(ctx context.Context, name string) ([]Product, error)
	FindByID(ctx context.Context, id string) (*Product, error)
	FindPaginated(ctx context.Context, name string, limit, offset int) ([]Product, int64, error)
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

func (r *repository) Create(ctx context.Context, product *Product) error {
	return r.db.
		WithContext(ctx).
		Table("products").
		Create(product).
		Error
}

func (r *repository) FindActive(ctx context.Context, name string) ([]Product, error) {
	var products []Product

	err := r.filterByName(ctx, name).
		Select("id", "name", "price", "photo").
		Order("name ASC").
		Find(&products).
		Error

	return products, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*Product, error) {
	var product Product

	err := r.db.
		WithContext(ctx).
		Table("products").
		Select("id", "name", "price", "photo").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		First(&product).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

func (r *repository) FindPaginated(ctx context.Context, name string, limit, offset int) ([]Product, int64, error) {
	var products []Product
	var total int64

	if err := r.filterByName(ctx, name).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.filterByName(ctx, name).
		Select("id", "name", "price", "photo").
		Order("name ASC").
		Limit(limit).
		Offset(offset).
		Find(&products).
		Error

	return products, total, err
}

func (r *repository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	return r.db.
		WithContext(ctx).
		Table("products").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Updates(updates).
		Error
}

func (r *repository) SoftDelete(ctx context.Context, id string) (int64, error) {
	result := r.db.
		WithContext(ctx).
		Table("products").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Update("deleted_at", time.Now())

	return result.RowsAffected, result.Error
}

func (r *repository) filterByName(ctx context.Context, name string) *gorm.DB {
	query := r.db.
		WithContext(ctx).
		Table("products").
		Where("deleted_at IS NULL")

	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	return query
}
