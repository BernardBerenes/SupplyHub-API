package users

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var user User

	err := r.db.
		WithContext(ctx).
		Table("users").
		Select("id", "username", "password").
		Where("username = ?", username).
		Where("deleted_at IS NULL").
		First(&user).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
