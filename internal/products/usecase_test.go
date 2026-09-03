package products

import (
	"context"
	"errors"
	"io"
	"testing"
)

type fakeRepo struct {
	products  []Product
	total     int64
	createErr error
	deletedID string
}

func (r *fakeRepo) Create(ctx context.Context, product *Product) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.products = append(r.products, *product)
	return nil
}

func (r *fakeRepo) FindActive(ctx context.Context, name string) ([]Product, error) {
	return r.products, nil
}

func (r *fakeRepo) FindByID(ctx context.Context, id string) (*Product, error) {
	for _, p := range r.products {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, nil
}

func (r *fakeRepo) FindByIDIncludingDeleted(ctx context.Context, id string) (*Product, error) {
	return r.FindByID(ctx, id)
}

func (r *fakeRepo) FindPaginated(ctx context.Context, name string, limit, offset int) ([]Product, int64, error) {
	return r.products, r.total, nil
}

func (r *fakeRepo) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	return nil
}

func (r *fakeRepo) SoftDelete(ctx context.Context, id string) (int64, error) {
	if id == r.deletedID {
		return 0, nil
	}
	return 1, nil
}

type fakeStorage struct {
	uploadedKey string
	deletedKey  string
}

func (s *fakeStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	s.uploadedKey = objectKey
	return nil
}

func (s *fakeStorage) Delete(ctx context.Context, objectKey string) error {
	s.deletedKey = objectKey
	return nil
}

func TestCreate_UploadsPhotoAndPersists(t *testing.T) {
	repo := &fakeRepo{}
	storage := &fakeStorage{}
	uc := NewUseCase(repo, storage)

	err := uc.Create(context.Background(), CreateInput{
		Name:  "Coffee",
		Price: 25000,
		Photo: &PhotoUpload{Data: []byte("fake-bytes"), Extension: ".jpg", ContentType: "image/jpeg"},
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(repo.products) != 1 {
		t.Fatalf("expected 1 product persisted, got %d", len(repo.products))
	}
	if repo.products[0].Photo == nil || *repo.products[0].Photo != storage.uploadedKey {
		t.Fatalf("expected persisted photo key to match uploaded key %q, got %v", storage.uploadedKey, repo.products[0].Photo)
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := &fakeRepo{deletedID: "missing-id"}
	uc := NewUseCase(repo, &fakeStorage{})

	err := uc.Delete(context.Background(), "missing-id")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDetail_NotFound(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewUseCase(repo, &fakeStorage{})

	_, err := uc.Detail(context.Background(), "missing-id")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
