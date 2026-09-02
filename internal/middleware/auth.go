package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/BernardBerenes/SupplyHub-API/presenter"
)

func JWTAuth(secret string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		header := ctx.Get(fiber.HeaderAuthorization)
		if !strings.HasPrefix(header, "Bearer ") {
			return presenter.ErrorResponse(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")

		claims := jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(
			tokenString,
			&claims,
			func(t *jwt.Token) (interface{}, error) { return []byte(secret), nil },
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		)
		if err != nil || !token.Valid {
			return presenter.ErrorResponse(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
		}

		ctx.Locals("userID", claims.Subject)

		return ctx.Next()
	}
}
