package products

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("product not found")

const photoObjectPrefix = "products/"

type Storage interface {
	Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, objectKey string) error
}

type UseCase struct {
	repo    Repository
	storage Storage
}

func NewUseCase(repo Repository, storage Storage) *UseCase {
	return &UseCase{
		repo:    repo,
		storage: storage,
	}
}

func (u *UseCase) Create(ctx context.Context, input CreateInput) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	product := &Product{
		ID:    id.String(),
		Name:  input.Name,
		Price: input.Price,
	}

	if input.Photo != nil {
		objectKey, err := u.uploadPhoto(ctx, product.ID, input.Photo)
		if err != nil {
			return err
		}
		product.Photo = &objectKey
	}

	return u.repo.Create(ctx, product)
}

func (u *UseCase) List(ctx context.Context, name string) ([]Product, error) {
	return u.repo.FindActive(ctx, name)
}

func (u *UseCase) Detail(ctx context.Context, id string) (*Product, error) {
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrNotFound
	}

	return product, nil
}

func (u *UseCase) Paginate(ctx context.Context, req PaginateRequest) ([]Product, int64, error) {
	offset := (req.Page - 1) * req.Limit

	return u.repo.FindPaginated(ctx, req.Name, req.Limit, offset)
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

	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Price != nil {
		updates["price"] = *input.Price
	}

	if input.Photo != nil {
		objectKey, err := u.uploadPhoto(ctx, id, input.Photo)
		if err != nil {
			return err
		}
		updates["photo"] = objectKey
	}

	if len(updates) == 0 {
		return nil
	}

	if err := u.repo.Update(ctx, id, updates); err != nil {
		return err
	}

	if input.Photo != nil && existing.Photo != nil {
		_ = u.storage.Delete(ctx, *existing.Photo)
	}

	return nil
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

func (u *UseCase) uploadPhoto(ctx context.Context, productID string, photo *PhotoUpload) (string, error) {
	objectKey := photoObjectPrefix + productID + photo.Extension

	if err := u.storage.Upload(ctx, objectKey, bytes.NewReader(photo.Data), int64(len(photo.Data)), photo.ContentType); err != nil {
		return "", err
	}

	return objectKey, nil
}
