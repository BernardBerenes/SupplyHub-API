package users

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type UseCase struct {
	repo      Repository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewUseCase(repo Repository, jwtSecret string, jwtExpiry time.Duration) *UseCase {
	return &UseCase{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (u *UseCase) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := u.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(user.ID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(u.jwtExpiry)),
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(u.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken: signed,
	}, nil
}
