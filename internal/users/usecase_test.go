package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type fakeRepo struct {
	user *User
	err  error
}

func (r *fakeRepo) FindByUsername(ctx context.Context, username string) (*User, error) {
	return r.user, r.err
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hash)
}

func TestLogin_Success(t *testing.T) {
	repo := &fakeRepo{user: &User{ID: 42, Username: "bernard", Password: mustHash(t, "secret")}}
	uc := NewUseCase(repo, "test-secret", 30*time.Minute)

	res, err := uc.Login(context.Background(), LoginRequest{Username: "bernard", Password: "secret"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	claims := jwt.RegisteredClaims{}
	_, err = jwt.ParseWithClaims(res.AccessToken, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("expected parseable token, got error: %v", err)
	}

	if claims.Subject != "42" {
		t.Fatalf("expected subject 42, got %s", claims.Subject)
	}

	if !claims.ExpiresAt.Time.After(time.Now()) {
		t.Fatalf("expected expiry in the future, got %v", claims.ExpiresAt.Time)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := &fakeRepo{user: &User{ID: 42, Username: "bernard", Password: mustHash(t, "secret")}}
	uc := NewUseCase(repo, "test-secret", 30*time.Minute)

	_, err := uc.Login(context.Background(), LoginRequest{Username: "bernard", Password: "wrong"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UnknownUsername(t *testing.T) {
	repo := &fakeRepo{user: nil}
	uc := NewUseCase(repo, "test-secret", 30*time.Minute)

	_, err := uc.Login(context.Background(), LoginRequest{Username: "ghost", Password: "secret"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
