package stores

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func newTestApp(repo Repository) *fiber.App {
	h := NewHandler(NewUseCase(repo))

	app := fiber.New()
	app.Post("/stores", h.Create)
	app.Get("/stores", h.List)
	app.Post("/stores/paginate", h.Paginate)
	app.Patch("/stores/:uuid", h.Update)
	app.Delete("/stores/:uuid", h.Delete)

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
	app := newTestApp(repo)

	resp := doRequest(t, app, http.MethodPost, "/stores", `{"name":"Main Store"}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(repo.stores) != 1 || repo.stores[0].Name != "Main Store" {
		t.Fatalf("expected store persisted, got %+v", repo.stores)
	}
}

func TestHandlerCreate_EmptyName(t *testing.T) {
	repo := &fakeRepo{}
	app := newTestApp(repo)

	resp := doRequest(t, app, http.MethodPost, "/stores", `{"name":"   "}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlerList_Success(t *testing.T) {
	repo := &fakeRepo{stores: []Store{{ID: 1, Name: "Main Store"}}}
	app := newTestApp(repo)

	resp := doRequest(t, app, http.MethodGet, "/stores", "")

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeBody(t, resp)
	data, ok := body["data"].(map[string]interface{})
	if !ok || data["stores"] == nil {
		t.Fatalf("expected stores in response, got %v", body)
	}
}

func TestHandlerPaginate_Success(t *testing.T) {
	repo := &fakeRepo{stores: []Store{{ID: 1, Name: "Main Store"}}, total: 1}
	app := newTestApp(repo)

	resp := doRequest(t, app, http.MethodPost, "/stores/paginate", `{}`)

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
	app := newTestApp(repo)

	resp := doRequest(t, app, http.MethodPost, "/stores/paginate", `{"limit":999}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlerUpdate_Success(t *testing.T) {
	repo := &fakeRepo{stores: []Store{{ID: 1, Name: "Old Name"}}}
	app := newTestApp(repo)

	resp := doRequest(t, app, http.MethodPatch, "/stores/1", `{"name":"New Name"}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandlerUpdate_NotFound(t *testing.T) {
	repo := &fakeRepo{}
	app := newTestApp(repo)

	resp := doRequest(t, app, http.MethodPatch, "/stores/999", `{"name":"New Name"}`)

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandlerDelete_Success(t *testing.T) {
	repo := &fakeRepo{deletedID: 999}
	app := newTestApp(repo)

	resp := doRequest(t, app, http.MethodDelete, "/stores/1", "")

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandlerDelete_NotFound(t *testing.T) {
	repo := &fakeRepo{deletedID: 1}
	app := newTestApp(repo)

	resp := doRequest(t, app, http.MethodDelete, "/stores/1", "")

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
