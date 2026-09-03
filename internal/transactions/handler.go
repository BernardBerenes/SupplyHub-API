package transactions

import (
	"errors"
	"strings"
	"time"

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

	if req.StoreID <= 0 {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
			{Field: "store_id", Message: "store_id is required"},
		})
	}

	date, err := time.Parse(DateFormat, req.Date)
	if err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
			{Field: "date", Message: "date must be in YYYY-MM-DD format"},
		})
	}

	err = h.useCase.Create(ctx.Context(), CreateInput{StoreID: req.StoreID, Date: date})
	if err != nil {
		switch {
		case errors.Is(err, ErrStoreNotFound):
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Store not found", nil)
		case errors.Is(err, ErrDuplicatePending):
			return presenter.ErrorResponse(ctx, fiber.StatusConflict, "Store already has a pending transaction on this date", nil)
		default:
			return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
		}
	}

	return presenter.OKWithoutData(ctx, "Transaction created successfully")
}

func (h *Handler) Paginate(ctx *fiber.Ctx) error {
	var req PaginateRequest

	if ok, err := presenter.BindPaginate(ctx, &req, &req.Page, &req.Limit, 10); !ok {
		return err
	}

	req.PaymentStatus = strings.ToUpper(strings.TrimSpace(req.PaymentStatus))
	req.DeliveryStatus = strings.ToUpper(strings.TrimSpace(req.DeliveryStatus))

	if req.DateFrom != "" && req.DateTo != "" && req.DateTo < req.DateFrom {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
			{Field: "date_to", Message: "date_to must not be before date_from"},
		})
	}

	transactions, total, err := h.useCase.Paginate(ctx.Context(), req)
	if err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	mapped, metadata := presenter.MapToResponseListPaginate(transactions, total, req.Page, req.Limit, ToResponse)

	return presenter.OK(ctx, "Transactions retrieved successfully", PaginateResponse{
		Page:         metadata.Page,
		Size:         metadata.Size,
		TotalItem:    metadata.Total,
		TotalPage:    metadata.TotalPage,
		Transactions: mapped,
	})
}

func (h *Handler) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("uuid")

	var req UpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", nil)
	}

	var input UpdateInput

	if req.StoreID != nil {
		if *req.StoreID <= 0 {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "store_id", Message: "store_id must be valid"},
			})
		}
		input.StoreID = req.StoreID
	}

	if req.PaymentStatus != nil {
		status := strings.ToUpper(strings.TrimSpace(*req.PaymentStatus))
		if status != PAYMENT_STATUS_PAID && status != PAYMENT_STATUS_UNPAID {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "payment_status", Message: "payment_status must be PAID or UNPAID"},
			})
		}
		input.PaymentStatus = &status
	}

	if req.DeliveryStatus != nil {
		status := strings.ToUpper(strings.TrimSpace(*req.DeliveryStatus))
		if status != DELIVERY_STATUS_PENDING && status != DELIVERY_STATUS_ON_DELIVERY && status != DELIVERY_STATUS_DELIVERED {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "delivery_status", Message: "delivery_status must be PENDING, ON_DELIVERY, or DELIVERED"},
			})
		}
		input.DeliveryStatus = &status
	}

	if req.Date != nil {
		date, err := time.Parse(DateFormat, *req.Date)
		if err != nil {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "date", Message: "date must be in YYYY-MM-DD format"},
			})
		}
		input.Date = &date
	}

	err := h.useCase.Update(ctx.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Transaction not found", nil)
		case errors.Is(err, ErrStoreNotFound):
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Store not found", nil)
		case errors.Is(err, ErrStoreUpdateNotPending):
			return presenter.ErrorResponse(ctx, fiber.StatusConflict, "Store can only be changed while the transaction is pending", nil)
		default:
			return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
		}
	}

	return presenter.OKWithoutData(ctx, "Transaction updated successfully")
}

func (h *Handler) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("uuid")

	if err := h.useCase.Delete(ctx.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Transaction not found", nil)
		}
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OKWithoutData(ctx, "Transaction deleted successfully")
}

func (h *Handler) Sync(ctx *fiber.Ctx) error {
	if err := h.useCase.SyncStoreNames(ctx.Context()); err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OKWithoutData(ctx, "Transactions synced successfully")
}
