package products

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/BernardBerenes/SupplyHub-API/presenter"
)

var validate = validator.New()

const MAX_PHOTO_SIZE_BYTES = 5 * 1024 * 1024

type photoFormat struct {
	extension   string
	contentType string
}

var allowedPhotoFormats = map[string]photoFormat{
	"jpeg": {extension: ".jpg", contentType: "image/jpeg"},
	"png":  {extension: ".png", contentType: "image/png"},
}

type Handler struct {
	useCase *UseCase
}

func NewHandler(useCase *UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

func (h *Handler) Create(ctx *fiber.Ctx) error {
	name := strings.TrimSpace(ctx.FormValue("name"))
	if name == "" || len(name) > 100 {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
			{Field: "name", Message: "name is required and must be at most 100 characters"},
		})
	}

	price, err := strconv.ParseInt(ctx.FormValue("price"), 10, 64)
	if err != nil || price < 0 {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
			{Field: "price", Message: "price is required and must not be negative"},
		})
	}

	input := CreateInput{
		Name:  name,
		Price: price,
	}

	if fileHeader, ferr := ctx.FormFile("photo"); ferr == nil {
		photo, perr := parsePhoto(fileHeader)
		if perr != nil {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "photo", Message: perr.Error()},
			})
		}
		input.Photo = photo
	}

	if err := h.useCase.Create(ctx.Context(), input); err != nil {
		log.Printf("create product failed: %v", err)
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OKWithoutData(ctx, "Product created successfully")
}

func (h *Handler) List(ctx *fiber.Ctx) error {
	name := ctx.Query("name")

	products, err := h.useCase.List(ctx.Context(), name)
	if err != nil {
		log.Printf("list products failed: %v", err)
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OK(ctx, "Products retrieved successfully", ListResponse{
		Products: presenter.MapToResponseList(products, ToResponse),
	})
}

func (h *Handler) Detail(ctx *fiber.Ctx) error {
	id := ctx.Params("uuid")

	product, err := h.useCase.Detail(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Product not found", nil)
		}
		log.Printf("get product failed: %v", err)
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OK(ctx, "Product retrieved successfully", DetailResponse{
		Product: ToResponse(*product),
	})
}

func (h *Handler) Paginate(ctx *fiber.Ctx) error {
	var req PaginateRequest

	if err := ctx.BodyParser(&req); err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", nil)
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	if err := validate.Struct(req); err != nil {
		return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", presenter.FormatValidationError(err))
	}

	products, total, err := h.useCase.Paginate(ctx.Context(), req)
	if err != nil {
		log.Printf("paginate products failed: %v", err)
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	mapped, metadata := presenter.MapToResponseListPaginate(products, total, req.Page, req.Limit, ToResponse)

	return presenter.OK(ctx, "Products retrieved successfully", PaginateResponse{
		Page:      metadata.Page,
		Size:      metadata.Size,
		TotalItem: metadata.Total,
		TotalPage: metadata.TotalPage,
		Products:  mapped,
	})
}

func (h *Handler) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("uuid")

	var input UpdateInput

	if name := strings.TrimSpace(ctx.FormValue("name")); name != "" {
		if len(name) > 100 {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "name", Message: "name must be at most 100 characters"},
			})
		}
		input.Name = &name
	}

	if priceStr := ctx.FormValue("price"); priceStr != "" {
		price, err := strconv.ParseInt(priceStr, 10, 64)
		if err != nil || price < 0 {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "price", Message: "price must not be negative"},
			})
		}
		input.Price = &price
	}

	if fileHeader, ferr := ctx.FormFile("photo"); ferr == nil {
		photo, perr := parsePhoto(fileHeader)
		if perr != nil {
			return presenter.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", []presenter.ErrorItem{
				{Field: "photo", Message: perr.Error()},
			})
		}
		input.Photo = photo
	}

	if err := h.useCase.Update(ctx.Context(), id, input); err != nil {
		if errors.Is(err, ErrNotFound) {
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Product not found", nil)
		}
		log.Printf("update product failed: %v", err)
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OKWithoutData(ctx, "Product updated successfully")
}

func (h *Handler) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("uuid")

	if err := h.useCase.Delete(ctx.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return presenter.ErrorResponse(ctx, fiber.StatusNotFound, "Product not found", nil)
		}
		log.Printf("delete product failed: %v", err)
		return presenter.ErrorResponse(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return presenter.OKWithoutData(ctx, "Product deleted successfully")
}

func parsePhoto(fileHeader *multipart.FileHeader) (*PhotoUpload, error) {
	if fileHeader.Size > MAX_PHOTO_SIZE_BYTES {
		return nil, fmt.Errorf("photo exceeds maximum size of %d bytes", MAX_PHOTO_SIZE_BYTES)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("photo is not a valid image")
	}

	info, ok := allowedPhotoFormats[format]
	if !ok {
		return nil, fmt.Errorf("unsupported photo format: %s", format)
	}

	return &PhotoUpload{
		Data:        data,
		Extension:   info.extension,
		ContentType: info.contentType,
	}, nil
}
