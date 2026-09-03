package transactiondetails

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/BernardBerenes/SupplyHub-API/internal/products"
)

var (
	ErrNotFound            = errors.New("transaction detail not found")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrProductNotFound     = errors.New("product not found")
)

// ProductLookup is the contract this domain needs from the Product module,
// satisfied by *products.UseCase.
type ProductLookup interface {
	FindByID(ctx context.Context, id string) (*products.Product, error)
	FindByIDIncludingDeleted(ctx context.Context, id string) (*products.Product, error)
}

// TransactionLookup is the contract this domain needs from the Transaction
// module, satisfied by *transactions.UseCase. It deliberately exposes no
// Transaction-specific type, so this package does not need to import the
// transactions package at all.
type TransactionLookup interface {
	Exists(ctx context.Context, id string) (bool, error)
	FindAllPendingIDs(ctx context.Context) ([]string, error)
}

type UseCase struct {
	repo        Repository
	product     ProductLookup
	transaction TransactionLookup
}

func NewUseCase(repo Repository, product ProductLookup, transaction TransactionLookup) *UseCase {
	return &UseCase{
		repo:        repo,
		product:     product,
		transaction: transaction,
	}
}

func (u *UseCase) Create(ctx context.Context, input CreateInput) error {
	exists, err := u.transaction.Exists(ctx, input.TransactionID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTransactionNotFound
	}

	product, err := u.product.FindByID(ctx, input.ProductID)
	if err != nil {
		return err
	}
	if product == nil {
		return ErrProductNotFound
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	detail := &TransactionDetail{
		ID:            id.String(),
		TransactionID: input.TransactionID,
		Product:       ProductSnapshot{ID: product.ID, Name: product.Name},
		Quantity:      input.Quantity,
		Unit:          input.Unit,
		Price:         input.Price,
	}

	return u.repo.Create(ctx, detail)
}

func (u *UseCase) Paginate(ctx context.Context, transactionID string, req PaginateRequest) ([]TransactionDetail, int64, error) {
	exists, err := u.transaction.Exists(ctx, transactionID)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, ErrTransactionNotFound
	}

	offset := (req.Page - 1) * req.Limit

	return u.repo.FindPaginatedByTransactionID(ctx, transactionID, req.Limit, offset)
}

func (u *UseCase) Update(ctx context.Context, transactionID, id string, input UpdateInput) error {
	existing, err := u.repo.FindByID(ctx, transactionID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}

	updates := map[string]interface{}{}

	if input.ProductID != nil {
		product, err := u.product.FindByID(ctx, *input.ProductID)
		if err != nil {
			return err
		}
		if product == nil {
			return ErrProductNotFound
		}
		updates["product"] = ProductSnapshot{ID: product.ID, Name: product.Name}
	}

	if input.Quantity != nil {
		updates["quantity"] = *input.Quantity
	}
	if input.Unit != nil {
		updates["unit"] = *input.Unit
	}
	if input.Price != nil {
		updates["price"] = *input.Price
	}

	if len(updates) == 0 {
		return nil
	}

	return u.repo.Update(ctx, id, updates)
}

func (u *UseCase) Delete(ctx context.Context, transactionID, id string) error {
	rows, err := u.repo.SoftDelete(ctx, transactionID, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// SyncProductNames keeps the product snapshot of every transaction detail
// belonging to a PENDING transaction in sync with the current product name.
func (u *UseCase) SyncProductNames(ctx context.Context) error {
	transactionIDs, err := u.transaction.FindAllPendingIDs(ctx)
	if err != nil {
		return err
	}

	for _, transactionID := range transactionIDs {
		details, err := u.repo.FindActiveByTransactionID(ctx, transactionID)
		if err != nil {
			return err
		}

		for _, detail := range details {
			product, err := u.product.FindByIDIncludingDeleted(ctx, detail.Product.ID)
			if err != nil {
				return err
			}
			if product == nil || product.Name == detail.Product.Name {
				continue
			}

			updates := map[string]interface{}{
				"product": ProductSnapshot{ID: product.ID, Name: product.Name},
			}

			if err := u.repo.Update(ctx, detail.ID, updates); err != nil {
				return err
			}
		}
	}

	return nil
}
