package transactiondetails

import (
	"context"
	"errors"
	"testing"

	"github.com/BernardBerenes/SupplyHub-API/internal/products"
)

type fakeRepo struct {
	details   []TransactionDetail
	total     int64
	deletedID string
}

func (r *fakeRepo) Create(ctx context.Context, detail *TransactionDetail) error {
	r.details = append(r.details, *detail)
	return nil
}

func (r *fakeRepo) FindByID(ctx context.Context, transactionID, id string) (*TransactionDetail, error) {
	for _, d := range r.details {
		if d.ID == id && d.TransactionID == transactionID {
			return &d, nil
		}
	}
	return nil, nil
}

func (r *fakeRepo) FindPaginatedByTransactionID(ctx context.Context, transactionID string, limit, offset int) ([]TransactionDetail, int64, error) {
	return r.details, r.total, nil
}

func (r *fakeRepo) FindActiveByTransactionID(ctx context.Context, transactionID string) ([]TransactionDetail, error) {
	var result []TransactionDetail
	for _, d := range r.details {
		if d.TransactionID == transactionID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (r *fakeRepo) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	for i, d := range r.details {
		if d.ID != id {
			continue
		}
		if product, ok := updates["product"].(ProductSnapshot); ok {
			r.details[i].Product = product
		}
		if quantity, ok := updates["quantity"].(int64); ok {
			r.details[i].Quantity = quantity
		}
		if unit, ok := updates["unit"].(string); ok {
			r.details[i].Unit = unit
		}
		if pricePerUnit, ok := updates["price_per_unit"].(int64); ok {
			r.details[i].PricePerUnit = pricePerUnit
		}
		if totalPrice, ok := updates["total_price"].(int64); ok {
			r.details[i].TotalPrice = totalPrice
		}
	}
	return nil
}

func (r *fakeRepo) SoftDelete(ctx context.Context, transactionID, id string) (int64, error) {
	if id == r.deletedID {
		return 0, nil
	}
	return 1, nil
}

type fakeProductLookup struct {
	products map[string]products.Product
}

func (r *fakeProductLookup) FindByID(ctx context.Context, id string) (*products.Product, error) {
	if p, ok := r.products[id]; ok {
		return &p, nil
	}
	return nil, nil
}

func (r *fakeProductLookup) FindByIDIncludingDeleted(ctx context.Context, id string) (*products.Product, error) {
	return r.FindByID(ctx, id)
}

type fakeTransactionLookup struct {
	existingIDs map[string]bool
	pendingIDs  []string
}

func (r *fakeTransactionLookup) Exists(ctx context.Context, id string) (bool, error) {
	return r.existingIDs[id], nil
}

func (r *fakeTransactionLookup) FindAllPendingIDs(ctx context.Context) ([]string, error) {
	return r.pendingIDs, nil
}

func TestCreate_Success(t *testing.T) {
	repo := &fakeRepo{}
	productLookup := &fakeProductLookup{products: map[string]products.Product{"p1": {ID: "p1", Name: "Coffee"}}}
	transactionLookup := &fakeTransactionLookup{existingIDs: map[string]bool{"t1": true}}
	uc := NewUseCase(repo, productLookup, transactionLookup)

	err := uc.Create(context.Background(), CreateInput{
		TransactionID: "t1",
		ProductID:     "p1",
		Quantity:      12,
		Unit:          UNIT_DOZENS,
		Price:         25000,
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(repo.details) != 1 {
		t.Fatalf("expected detail persisted, got %+v", repo.details)
	}
	if repo.details[0].Product != (ProductSnapshot{ID: "p1", Name: "Coffee", Price: 25000}) {
		t.Fatalf("expected product snapshot with price, got %+v", repo.details[0].Product)
	}
	if want := int64(25000 * 12); repo.details[0].PricePerUnit != want {
		t.Fatalf("expected price_per_unit = price * 12 for dozens, got %d want %d", repo.details[0].PricePerUnit, want)
	}
	if want := int64(25000 * 12 * 12); repo.details[0].TotalPrice != want {
		t.Fatalf("expected total_price = price * 12 * quantity for dozens, got %d want %d", repo.details[0].TotalPrice, want)
	}
}

func TestCreate_TransactionNotFound(t *testing.T) {
	repo := &fakeRepo{}
	productLookup := &fakeProductLookup{}
	transactionLookup := &fakeTransactionLookup{existingIDs: map[string]bool{}}
	uc := NewUseCase(repo, productLookup, transactionLookup)

	err := uc.Create(context.Background(), CreateInput{TransactionID: "missing", ProductID: "p1"})
	if !errors.Is(err, ErrTransactionNotFound) {
		t.Fatalf("expected ErrTransactionNotFound, got %v", err)
	}
}

func TestCreate_ProductNotFound(t *testing.T) {
	repo := &fakeRepo{}
	productLookup := &fakeProductLookup{products: map[string]products.Product{}}
	transactionLookup := &fakeTransactionLookup{existingIDs: map[string]bool{"t1": true}}
	uc := NewUseCase(repo, productLookup, transactionLookup)

	err := uc.Create(context.Background(), CreateInput{TransactionID: "t1", ProductID: "missing"})
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewUseCase(repo, &fakeProductLookup{}, &fakeTransactionLookup{})

	quantity := int64(5)
	err := uc.Update(context.Background(), "t1", "missing", UpdateInput{Quantity: &quantity})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdate_RegeneratesProductSnapshot(t *testing.T) {
	repo := &fakeRepo{details: []TransactionDetail{
		{ID: "d1", TransactionID: "t1", Product: ProductSnapshot{ID: "p1", Name: "Coffee"}},
	}}
	productLookup := &fakeProductLookup{products: map[string]products.Product{"p2": {ID: "p2", Name: "Tea"}}}
	uc := NewUseCase(repo, productLookup, &fakeTransactionLookup{})

	productID := "p2"
	if err := uc.Update(context.Background(), "t1", "d1", UpdateInput{ProductID: &productID}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if repo.details[0].Product != (ProductSnapshot{ID: "p2", Name: "Tea"}) {
		t.Fatalf("expected product snapshot updated, got %+v", repo.details[0].Product)
	}
}

func TestUpdate_RecalculatesTotalPrice(t *testing.T) {
	repo := &fakeRepo{details: []TransactionDetail{
		{ID: "d1", TransactionID: "t1", Quantity: 1, Unit: UNIT_PIECES, PricePerUnit: 1000, TotalPrice: 1000},
	}}
	uc := NewUseCase(repo, &fakeProductLookup{}, &fakeTransactionLookup{})

	quantity := int64(2)
	unit := UNIT_DOZENS
	if err := uc.Update(context.Background(), "t1", "d1", UpdateInput{Quantity: &quantity, Unit: &unit}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if want := int64(1000 * 2 * 12); repo.details[0].TotalPrice != want {
		t.Fatalf("expected recalculated total_price %d, got %d", want, repo.details[0].TotalPrice)
	}
}

func TestUpdate_PriceUpdatesProductSnapshotAndPricing(t *testing.T) {
	repo := &fakeRepo{details: []TransactionDetail{
		{
			ID: "d1", TransactionID: "t1", Quantity: 2, Unit: UNIT_PIECES,
			PricePerUnit: 1000, TotalPrice: 2000,
			Product: ProductSnapshot{ID: "p1", Name: "Coffee", Price: 1000},
		},
	}}
	uc := NewUseCase(repo, &fakeProductLookup{}, &fakeTransactionLookup{})

	price := int64(1500)
	if err := uc.Update(context.Background(), "t1", "d1", UpdateInput{Price: &price}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if repo.details[0].Product.Price != 1500 {
		t.Fatalf("expected product snapshot price updated, got %+v", repo.details[0].Product)
	}
	if repo.details[0].PricePerUnit != 1500 || repo.details[0].TotalPrice != 3000 {
		t.Fatalf("expected recalculated pricing, got %+v", repo.details[0])
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := &fakeRepo{deletedID: "missing"}
	uc := NewUseCase(repo, &fakeProductLookup{}, &fakeTransactionLookup{})

	err := uc.Delete(context.Background(), "t1", "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSyncProductNames_UpdatesOnlyPendingTransactionDetails(t *testing.T) {
	repo := &fakeRepo{details: []TransactionDetail{
		{ID: "d1", TransactionID: "t1", Product: ProductSnapshot{ID: "p1", Name: "Coffee", Price: 1000}},
		{ID: "d2", TransactionID: "t2", Product: ProductSnapshot{ID: "p1", Name: "Coffee", Price: 1000}},
	}}
	productLookup := &fakeProductLookup{products: map[string]products.Product{"p1": {ID: "p1", Name: "Premium Coffee"}}}
	transactionLookup := &fakeTransactionLookup{pendingIDs: []string{"t1"}}
	uc := NewUseCase(repo, productLookup, transactionLookup)

	if err := uc.SyncProductNames(context.Background()); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if repo.details[0].Product.Name != "Premium Coffee" {
		t.Fatalf("expected detail of pending transaction synced, got %+v", repo.details[0])
	}
	if repo.details[0].Product.Price != 1000 {
		t.Fatalf("expected snapshot price preserved on name sync, got %+v", repo.details[0])
	}
	if repo.details[1].Product.Name != "Coffee" {
		t.Fatalf("expected detail of non-pending transaction untouched, got %+v", repo.details[1])
	}
}
