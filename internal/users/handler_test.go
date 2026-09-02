package users

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func newTestApp(uc *UseCase) *fiber.App {
	app := fiber.New()
	app.Post("/login", NewHandler(uc).Login)
	return app
}

func doLogin(t *testing.T, app *fiber.App, body string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
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

func TestHandlerLogin_Success(t *testing.T) {
	repo := &fakeRepo{user: &User{ID: 42, Username: "bernard", Password: mustHash(t, "secret")}}
	app := newTestApp(NewUseCase(repo, "test-secret", 30*time.Minute))

	resp := doLogin(t, app, `{"username":"bernard","password":"secret"}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeBody(t, resp)
	data, ok := body["data"].(map[string]interface{})
	if !ok || data["access_token"] == nil || data["access_token"] == "" {
		t.Fatalf("expected non-empty access_token in response, got %v", body)
	}
}

func TestHandlerLogin_WrongPassword(t *testing.T) {
	repo := &fakeRepo{user: &User{ID: 42, Username: "bernard", Password: mustHash(t, "secret")}}
	app := newTestApp(NewUseCase(repo, "test-secret", 30*time.Minute))

	resp := doLogin(t, app, `{"username":"bernard","password":"wrong"}`)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestHandlerLogin_UnknownUsername(t *testing.T) {
	repo := &fakeRepo{user: nil}
	app := newTestApp(NewUseCase(repo, "test-secret", 30*time.Minute))

	resp := doLogin(t, app, `{"username":"ghost","password":"secret"}`)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestHandlerLogin_MissingFields(t *testing.T) {
	repo := &fakeRepo{}
	app := newTestApp(NewUseCase(repo, "test-secret", 30*time.Minute))

	resp := doLogin(t, app, `{"username":""}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlerLogin_MalformedBody(t *testing.T) {
	repo := &fakeRepo{}
	app := newTestApp(NewUseCase(repo, "test-secret", 30*time.Minute))

	resp := doLogin(t, app, `not-json`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
