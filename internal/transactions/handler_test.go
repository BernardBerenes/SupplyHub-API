package transactions

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/BernardBerenes/SupplyHub-API/internal/stores"
)

func newTestApp(repo Repository, storeRepo StoreLookup) *fiber.App {
	h := NewHandler(NewUseCase(repo, storeRepo))

	app := fiber.New()
	app.Post("/transactions", h.Create)
	app.Post("/transactions/paginate", h.Paginate)
	app.Post("/transactions/sync", h.Sync)
	app.Patch("/transactions/:uuid", h.Update)
	app.Delete("/transactions/:uuid", h.Delete)

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

func decodeBody(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	return body
}

func TestHandlerCreate_Success(t *testing.T) {
	repo := &fakeRepo{}
	storeRepo := &fakeStoreLookup{stores: map[int64]stores.Store{1: {ID: 1, Name: "Toko Surya"}}}
	app := newTestApp(repo, storeRepo)

	resp := doRequest(t, app, http.MethodPost, "/transactions", `{"store_id":1,"date":"2026-09-03"}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(repo.transactions) != 1 {
		t.Fatalf("expected transaction persisted, got %+v", repo.transactions)
	}
}

func TestHandlerCreate_InvalidDate(t *testing.T) {
	repo := &fakeRepo{}
	storeRepo := &fakeStoreLookup{stores: map[int64]stores.Store{1: {ID: 1, Name: "Toko Surya"}}}
	app := newTestApp(repo, storeRepo)

	resp := doRequest(t, app, http.MethodPost, "/transactions", `{"store_id":1,"date":"03-09-2026"}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlerCreate_StoreNotFound(t *testing.T) {
	repo := &fakeRepo{}
	storeRepo := &fakeStoreLookup{stores: map[int64]stores.Store{}}
	app := newTestApp(repo, storeRepo)

	resp := doRequest(t, app, http.MethodPost, "/transactions", `{"store_id":99,"date":"2026-09-03"}`)

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandlerCreate_DuplicatePending(t *testing.T) {
	repo := &fakeRepo{pending: true}
	storeRepo := &fakeStoreLookup{stores: map[int64]stores.Store{1: {ID: 1, Name: "Toko Surya"}}}
	app := newTestApp(repo, storeRepo)

	resp := doRequest(t, app, http.MethodPost, "/transactions", `{"store_id":1,"date":"2026-09-03"}`)

	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestHandlerPaginate_Success(t *testing.T) {
	repo := &fakeRepo{transactions: []Transaction{{ID: "t1"}}, total: 1}
	app := newTestApp(repo, &fakeStoreLookup{})

	resp := doRequest(t, app, http.MethodPost, "/transactions/paginate", `{}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeBody(t, resp)
	data, ok := body["data"].(map[string]interface{})
	if !ok || data["page"] != float64(1) {
		t.Fatalf("expected default page 1, got %v", body)
	}
}

func TestHandlerPaginate_InvalidLimit(t *testing.T) {
	repo := &fakeRepo{}
	app := newTestApp(repo, &fakeStoreLookup{})

	resp := doRequest(t, app, http.MethodPost, "/transactions/paginate", `{"limit":999}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlerPaginate_InvalidDeliveryStatus(t *testing.T) {
	repo := &fakeRepo{}
	app := newTestApp(repo, &fakeStoreLookup{})

	resp := doRequest(t, app, http.MethodPost, "/transactions/paginate", `{"delivery_status":"SHIPPED"}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlerPaginate_DateRange(t *testing.T) {
	repo := &fakeRepo{transactions: []Transaction{{ID: "t1"}}, total: 1}
	app := newTestApp(repo, &fakeStoreLookup{})

	resp := doRequest(t, app, http.MethodPost, "/transactions/paginate", `{"date_from":"2026-09-01","date_to":"2026-09-30"}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandlerPaginate_RejectsInvertedDateRange(t *testing.T) {
	repo := &fakeRepo{}
	app := newTestApp(repo, &fakeStoreLookup{})

	resp := doRequest(t, app, http.MethodPost, "/transactions/paginate", `{"date_from":"2026-09-30","date_to":"2026-09-01"}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlerUpdate_Success(t *testing.T) {
	repo := &fakeRepo{transactions: []Transaction{{ID: "t1", DeliveryStatus: DELIVERY_STATUS_PENDING}}}
	app := newTestApp(repo, &fakeStoreLookup{})

	resp := doRequest(t, app, http.MethodPatch, "/transactions/t1", `{"payment_status":"PAID"}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandlerUpdate_NotFound(t *testing.T) {
	repo := &fakeRepo{}
	app := newTestApp(repo, &fakeStoreLookup{})

	resp := doRequest(t, app, http.MethodPatch, "/transactions/missing", `{"payment_status":"PAID"}`)

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandlerUpdate_StoreChangeRejectedWhenNotPending(t *testing.T) {
	repo := &fakeRepo{transactions: []Transaction{{ID: "t1", DeliveryStatus: DELIVERY_STATUS_ON_DELIVERY}}}
	storeRepo := &fakeStoreLookup{stores: map[int64]stores.Store{2: {ID: 2, Name: "Toko Baru"}}}
	app := newTestApp(repo, storeRepo)

	resp := doRequest(t, app, http.MethodPatch, "/transactions/t1", `{"store_id":2}`)

	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestHandlerDelete_Success(t *testing.T) {
	repo := &fakeRepo{deletedID: "missing"}
	app := newTestApp(repo, &fakeStoreLookup{})

	resp := doRequest(t, app, http.MethodDelete, "/transactions/t1", "")

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandlerDelete_NotFound(t *testing.T) {
	repo := &fakeRepo{deletedID: "t1"}
	app := newTestApp(repo, &fakeStoreLookup{})

	resp := doRequest(t, app, http.MethodDelete, "/transactions/t1", "")

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandlerSync_Success(t *testing.T) {
	repo := &fakeRepo{transactions: []Transaction{
		{ID: "t1", DeliveryStatus: DELIVERY_STATUS_PENDING, Store: StoreSnapshot{ID: 1, Name: "Toko Surya"}},
	}}
	storeRepo := &fakeStoreLookup{stores: map[int64]stores.Store{1: {ID: 1, Name: "Toko Surya Baru"}}}
	app := newTestApp(repo, storeRepo)

	resp := doRequest(t, app, http.MethodPost, "/transactions/sync", "")

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if repo.transactions[0].Store.Name != "Toko Surya Baru" {
		t.Fatalf("expected store name synced, got %+v", repo.transactions[0])
	}
}
