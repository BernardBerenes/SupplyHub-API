package stores

import (
	"context"
	"errors"
	"testing"
)

type fakeRepo struct {
	stores    []Store
	total     int64
	deletedID int64
}

func (r *fakeRepo) Create(ctx context.Context, store *Store) error {
	r.stores = append(r.stores, *store)
	return nil
}

func (r *fakeRepo) FindActive(ctx context.Context, name string) ([]Store, error) {
	return r.stores, nil
}

func (r *fakeRepo) FindByID(ctx context.Context, id int64) (*Store, error) {
	for _, s := range r.stores {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, nil
}

func (r *fakeRepo) FindByIDIncludingDeleted(ctx context.Context, id int64) (*Store, error) {
	for _, s := range r.stores {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, nil
}

func (r *fakeRepo) FindPaginated(ctx context.Context, name string, limit, offset int) ([]Store, int64, error) {
	return r.stores, r.total, nil
}

func (r *fakeRepo) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return nil
}

func (r *fakeRepo) SoftDelete(ctx context.Context, id int64) (int64, error) {
	if id == r.deletedID {
		return 0, nil
	}
	return 1, nil
}

func TestCreate_TrimsNameAndPersists(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewUseCase(repo)

	if err := uc.Create(context.Background(), CreateInput{Name: "  Main Store  "}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(repo.stores) != 1 || repo.stores[0].Name != "Main Store" {
		t.Fatalf("expected trimmed name persisted, got %+v", repo.stores)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewUseCase(repo)

	name := "New Name"
	err := uc.Update(context.Background(), 999, UpdateInput{Name: &name})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := &fakeRepo{deletedID: 999}
	uc := NewUseCase(repo)

	err := uc.Delete(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
