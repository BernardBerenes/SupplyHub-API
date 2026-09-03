package stores

import (
	"context"
	"errors"
	"strings"
)

var ErrNotFound = errors.New("store not found")

type UseCase struct {
	repo Repository
}

func NewUseCase(repo Repository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

func (u *UseCase) Create(ctx context.Context, input CreateInput) error {
	store := &Store{
		Name: strings.TrimSpace(input.Name),
	}

	return u.repo.Create(ctx, store)
}

func (u *UseCase) List(ctx context.Context, name string) ([]Store, error) {
	return u.repo.FindActive(ctx, name)
}

func (u *UseCase) Paginate(ctx context.Context, req PaginateRequest) ([]Store, int64, error) {
	offset := (req.Page - 1) * req.Limit

	return u.repo.FindPaginated(ctx, req.Name, req.Limit, offset)
}

func (u *UseCase) Update(ctx context.Context, id int64, input UpdateInput) error {
	existing, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}

	updates := map[string]interface{}{}

	if input.Name != nil {
		updates["name"] = strings.TrimSpace(*input.Name)
	}

	if len(updates) == 0 {
		return nil
	}

	return u.repo.Update(ctx, id, updates)
}

func (u *UseCase) FindByID(ctx context.Context, id int64) (*Store, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *UseCase) FindByIDIncludingDeleted(ctx context.Context, id int64) (*Store, error) {
	return u.repo.FindByIDIncludingDeleted(ctx, id)
}

func (u *UseCase) Delete(ctx context.Context, id int64) error {
	rows, err := u.repo.SoftDelete(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
