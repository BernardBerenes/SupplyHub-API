package users

import (
	"errors"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/BernardBerenes/SupplyHub-API/presenter"
)

var validate = validator.New()

type Handler struct {
	useCase *UseCase
}

func NewHandler(useCase *UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

func (h *Handler) Login(ctx *fiber.Ctx) error {
	var req LoginRequest

	if err := ctx.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", nil)
	}

	if err := validate.Struct(req); err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", presenter.FormatValidationError(err))
	}

	res, err := h.useCase.Login(ctx.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return presenter.ErrorResponse(ctx, fiber.StatusUnauthorized, "Invalid username or password", nil)
		}
		log.Printf("login failed: %v", err)
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OK(ctx, "Login successful", res)
}
