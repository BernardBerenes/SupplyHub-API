package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func newTestApp(secret string) *fiber.App {
	app := fiber.New()
	app.Get("/protected", JWTAuth(secret), func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})
	return app
}

func signToken(t *testing.T, secret string, expiry time.Duration) string {
	t.Helper()
	claims := jwt.RegisteredClaims{
		Subject:   "1",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func TestJWTAuth(t *testing.T) {
	cases := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{"missing header", "", fiber.StatusUnauthorized},
		{"malformed header", "not-a-bearer-token", fiber.StatusUnauthorized},
		{"expired token", "Bearer " + signToken(t, "secret", -time.Minute), fiber.StatusUnauthorized},
		{"wrong signature", "Bearer " + signToken(t, "other-secret", time.Minute), fiber.StatusUnauthorized},
		{"valid token", "Bearer " + signToken(t, "secret", time.Minute), fiber.StatusOK},
	}

	app := newTestApp("secret")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.authHeader != "" {
				req.Header.Set(fiber.HeaderAuthorization, tc.authHeader)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, resp.StatusCode)
			}
		})
	}
}
