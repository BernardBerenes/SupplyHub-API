package stores

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/BernardBerenes/SupplyHub-API/presenter"
)

type Handler struct {
	useCase *UseCase
}

func NewHandler(useCase *UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

func (h *Handler) Create(ctx *fiber.Ctx) error {
	var req CreateRequest

	if err := ctx.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", nil)
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
			{Field: "name", Message: "name is required"},
		})
	}

	if err := h.useCase.Create(ctx.Context(), CreateInput{Name: name}); err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OKWithoutData(ctx, "Store created successfully")
}

func (h *Handler) List(ctx *fiber.Ctx) error {
	name := ctx.Query("name")

	stores, err := h.useCase.List(ctx.Context(), name)
	if err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OK(ctx, "Stores retrieved successfully", ListResponse{
		Stores: presenter.MapToResponseList(stores, ToResponse),
	})
}

func (h *Handler) Paginate(ctx *fiber.Ctx) error {
	var req PaginateRequest

	if ok, err := presenter.BindPaginate(ctx, &req, &req.Page, &req.Limit, 10); !ok {
		return err
	}

	stores, total, err := h.useCase.Paginate(ctx.Context(), req)
	if err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	mapped, metadata := presenter.MapToResponseListPaginate(stores, total, req.Page, req.Limit, ToResponse)

	return presenter.OK(ctx, "Stores retrieved successfully", PaginateResponse{
		Page:      metadata.Page,
		Size:      metadata.Size,
		TotalItem: metadata.Total,
		TotalPage: metadata.TotalPage,
		Stores:    mapped,
	})
}

func (h *Handler) Update(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("uuid")
	if err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Store not found", nil)
	}

	var req UpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", nil)
	}

	var input UpdateInput

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "name", Message: "name must not be empty"},
			})
		}
		input.Name = &name
	}

	if err := h.useCase.Update(ctx.Context(), int64(id), input); err != nil {
		if errors.Is(err, ErrNotFound) {
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Store not found", nil)
		}
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OKWithoutData(ctx, "Store updated successfully")
}

func (h *Handler) Delete(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("uuid")
	if err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Store not found", nil)
	}

	if err := h.useCase.Delete(ctx.Context(), int64(id)); err != nil {
		if errors.Is(err, ErrNotFound) {
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Store not found", nil)
		}
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OKWithoutData(ctx, "Store deleted successfully")
}
