package transactiondetails

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
	transactionID := ctx.Params("transaction_id")

	var req CreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", nil)
	}

	if strings.TrimSpace(req.ProductID) == "" {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
			{Field: "product_id", Message: "product_id is required"},
		})
	}

	if req.Quantity <= 0 {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
			{Field: "quantity", Message: "quantity must be greater than 0"},
		})
	}

	if req.Price <= 0 {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
			{Field: "price", Message: "price must be greater than 0"},
		})
	}

	unit := strings.ToUpper(strings.TrimSpace(req.Unit))
	if !IsValidUnit(unit) {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
			{Field: "unit", Message: "unit must be PIECES, DOZENS, BOX, or CARTON"},
		})
	}

	err := h.useCase.Create(ctx.Context(), CreateInput{
		TransactionID: transactionID,
		ProductID:     req.ProductID,
		Quantity:      req.Quantity,
		Unit:          unit,
		Price:         req.Price,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrTransactionNotFound):
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Transaction not found", nil)
		case errors.Is(err, ErrProductNotFound):
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Product not found", nil)
		default:
			return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
		}
	}

	return presenter.OKWithoutData(ctx, "Transaction detail created successfully")
}

func (h *Handler) Paginate(ctx *fiber.Ctx) error {
	transactionID := ctx.Params("transaction_id")

	var req PaginateRequest
	if ok, err := presenter.BindPaginate(ctx, &req, &req.Page, &req.Limit, 10); !ok {
		return err
	}

	details, total, err := h.useCase.Paginate(ctx.Context(), transactionID, req)
	if err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Transaction not found", nil)
		}
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	mapped, metadata := presenter.MapToResponseListPaginate(details, total, req.Page, req.Limit, ToResponse)

	return presenter.OK(ctx, "Transaction details retrieved successfully", PaginateResponse{
		Page:               metadata.Page,
		Size:               metadata.Size,
		TotalItem:          metadata.Total,
		TotalPage:          metadata.TotalPage,
		TransactionDetails: mapped,
	})
}

func (h *Handler) Update(ctx *fiber.Ctx) error {
	transactionID := ctx.Params("transaction_id")
	id := ctx.Params("uuid")

	var req UpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", nil)
	}

	var input UpdateInput

	if req.ProductID != nil {
		productID := strings.TrimSpace(*req.ProductID)
		if productID == "" {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "product_id", Message: "product_id must not be empty"},
			})
		}
		input.ProductID = &productID
	}

	if req.Quantity != nil {
		if *req.Quantity <= 0 {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "quantity", Message: "quantity must be greater than 0"},
			})
		}
		input.Quantity = req.Quantity
	}

	if req.Price != nil {
		if *req.Price <= 0 {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "price", Message: "price must be greater than 0"},
			})
		}
		input.Price = req.Price
	}

	if req.Unit != nil {
		unit := strings.ToUpper(strings.TrimSpace(*req.Unit))
		if !IsValidUnit(unit) {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "unit", Message: "unit must be PIECES, DOZENS, BOX, or CARTON"},
			})
		}
		input.Unit = &unit
	}

	err := h.useCase.Update(ctx.Context(), transactionID, id, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Transaction detail not found", nil)
		case errors.Is(err, ErrProductNotFound):
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Product not found", nil)
		default:
			return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
		}
	}

	return presenter.OKWithoutData(ctx, "Transaction detail updated successfully")
}

func (h *Handler) Delete(ctx *fiber.Ctx) error {
	transactionID := ctx.Params("transaction_id")
	id := ctx.Params("uuid")

	if err := h.useCase.Delete(ctx.Context(), transactionID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Transaction detail not found", nil)
		}
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OKWithoutData(ctx, "Transaction detail deleted successfully")
}
