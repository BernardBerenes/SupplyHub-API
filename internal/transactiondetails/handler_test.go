package transactiondetails

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/BernardBerenes/SupplyHub-API/internal/products"
)

func newTestApp(repo Repository, productLookup ProductLookup, transactionLookup TransactionLookup) *fiber.App {
	h := NewHandler(NewUseCase(repo, productLookup, transactionLookup))

	app := fiber.New()
	app.Post("/transactions/:transaction_id/details", h.Create)
	app.Post("/transactions/:transaction_id/details/paginate", h.Paginate)
	app.Patch("/transactions/:transaction_id/details/:uuid", h.Update)
	app.Delete("/transactions/:transaction_id/details/:uuid", h.Delete)

	return app
}

func doRequest(t *testing.T, app *fiber.App, method, path, body string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return resp
}

func TestHandlerCreate_Success(t *testing.T) {
	repo := &fakeRepo{}
	productLookup := &fakeProductLookup{products: map[string]products.Product{"p1": {ID: "p1", Name: "Coffee"}}}
	transactionLookup := &fakeTransactionLookup{existingIDs: map[string]bool{"t1": true}}
	app := newTestApp(repo, productLookup, transactionLookup)

	resp := doRequest(t, app, http.MethodPost, "/transactions/t1/details", `{"product_id":"p1","quantity":12,"unit":"dozens","price":25000}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(repo.details) != 1 {
		t.Fatalf("expected detail persisted, got %+v", repo.details)
	}
}

func TestHandlerCreate_InvalidUnit(t *testing.T) {
	repo := &fakeRepo{}
	productLookup := &fakeProductLookup{products: map[string]products.Product{"p1": {ID: "p1", Name: "Coffee"}}}
	transactionLookup := &fakeTransactionLookup{existingIDs: map[string]bool{"t1": true}}
	app := newTestApp(repo, productLookup, transactionLookup)

	resp := doRequest(t, app, http.MethodPost, "/transactions/t1/details", `{"product_id":"p1","quantity":12,"unit":"SACKS","price":25000}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlerCreate_NonPositiveQuantity(t *testing.T) {
	repo := &fakeRepo{}
	productLookup := &fakeProductLookup{products: map[string]products.Product{"p1": {ID: "p1", Name: "Coffee"}}}
	transactionLookup := &fakeTransactionLookup{existingIDs: map[string]bool{"t1": true}}
	app := newTestApp(repo, productLookup, transactionLookup)

	resp := doRequest(t, app, http.MethodPost, "/transactions/t1/details", `{"product_id":"p1","quantity":0,"unit":"PIECES","price":25000}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlerCreate_NonPositivePrice(t *testing.T) {
	repo := &fakeRepo{}
	productLookup := &fakeProductLookup{products: map[string]products.Product{"p1": {ID: "p1", Name: "Coffee"}}}
	transactionLookup := &fakeTransactionLookup{existingIDs: map[string]bool{"t1": true}}
	app := newTestApp(repo, productLookup, transactionLookup)

	resp := doRequest(t, app, http.MethodPost, "/transactions/t1/details", `{"product_id":"p1","quantity":1,"unit":"PIECES","price":0}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlerCreate_TransactionNotFound(t *testing.T) {
	repo := &fakeRepo{}
	productLookup := &fakeProductLookup{products: map[string]products.Product{"p1": {ID: "p1", Name: "Coffee"}}}
	transactionLookup := &fakeTransactionLookup{existingIDs: map[string]bool{}}
	app := newTestApp(repo, productLookup, transactionLookup)

	resp := doRequest(t, app, http.MethodPost, "/transactions/missing/details", `{"product_id":"p1","quantity":1,"unit":"PIECES","price":1000}`)

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandlerCreate_ProductNotFound(t *testing.T) {
	repo := &fakeRepo{}
	productLookup := &fakeProductLookup{products: map[string]products.Product{}}
	transactionLookup := &fakeTransactionLookup{existingIDs: map[string]bool{"t1": true}}
	app := newTestApp(repo, productLookup, transactionLookup)

	resp := doRequest(t, app, http.MethodPost, "/transactions/t1/details", `{"product_id":"missing","quantity":1,"unit":"PIECES","price":1000}`)

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandlerPaginate_Success(t *testing.T) {
	repo := &fakeRepo{details: []TransactionDetail{{ID: "d1", TransactionID: "t1"}}, total: 1}
	transactionLookup := &fakeTransactionLookup{existingIDs: map[string]bool{"t1": true}}
	app := newTestApp(repo, &fakeProductLookup{}, transactionLookup)

	resp := doRequest(t, app, http.MethodPost, "/transactions/t1/details/paginate", `{}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandlerPaginate_TransactionNotFound(t *testing.T) {
	repo := &fakeRepo{}
	transactionLookup := &fakeTransactionLookup{existingIDs: map[string]bool{}}
	app := newTestApp(repo, &fakeProductLookup{}, transactionLookup)

	resp := doRequest(t, app, http.MethodPost, "/transactions/missing/details/paginate", `{}`)

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandlerUpdate_Success(t *testing.T) {
	repo := &fakeRepo{details: []TransactionDetail{{ID: "d1", TransactionID: "t1"}}}
	app := newTestApp(repo, &fakeProductLookup{}, &fakeTransactionLookup{})

	resp := doRequest(t, app, http.MethodPatch, "/transactions/t1/details/d1", `{"quantity":5}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandlerUpdate_NotFound(t *testing.T) {
	repo := &fakeRepo{}
	app := newTestApp(repo, &fakeProductLookup{}, &fakeTransactionLookup{})

	resp := doRequest(t, app, http.MethodPatch, "/transactions/t1/details/missing", `{"quantity":5}`)

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandlerUpdate_MismatchedTransactionNotFound(t *testing.T) {
	repo := &fakeRepo{details: []TransactionDetail{{ID: "d1", TransactionID: "t1"}}}
	app := newTestApp(repo, &fakeProductLookup{}, &fakeTransactionLookup{})

	resp := doRequest(t, app, http.MethodPatch, "/transactions/other-transaction/details/d1", `{"quantity":5}`)

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandlerDelete_Success(t *testing.T) {
	repo := &fakeRepo{deletedID: "missing"}
	app := newTestApp(repo, &fakeProductLookup{}, &fakeTransactionLookup{})

	resp := doRequest(t, app, http.MethodDelete, "/transactions/t1/details/d1", "")

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandlerDelete_NotFound(t *testing.T) {
	repo := &fakeRepo{deletedID: "d1"}
	app := newTestApp(repo, &fakeProductLookup{}, &fakeTransactionLookup{})

	resp := doRequest(t, app, http.MethodDelete, "/transactions/t1/details/d1", "")

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
