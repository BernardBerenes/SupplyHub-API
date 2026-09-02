package presenter

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

type Success[T any] struct {
	Message          string            `json:"message"`
	Data             T                 `json:"data,omitempty"`
	PaginateMetadata *PaginateMetadata `json:"metadata,omitempty"`
}

type PaginateMetadata struct {
	Page      int   `json:"page"`
	Size      int   `json:"size"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
}

type Error struct {
	Message string      `json:"message,omitempty"`
	Errors  []ErrorItem `json:"errors,omitempty"`
}

type ErrorItem struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func OK(ctx *fiber.Ctx, message string, data any) error {
	return ctx.Status(fiber.StatusOK).JSON(Success[any]{
		Message: message,
		Data:    data,
	})
}

func OKWithoutData(ctx *fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusOK).JSON(Success[any]{
		Message: message,
	})
}

func ErrorResponse(ctx *fiber.Ctx, status int, message string, errors []ErrorItem) error {
	return ctx.Status(status).JSON(Error{
		Message: message,
		Errors:  errors,
	})
}

func FormatValidationError(err error) []ErrorItem {
	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil
	}

	items := make([]ErrorItem, 0, len(verrs))
	for _, fe := range verrs {
		field := strings.ToLower(fe.Field())
		items = append(items, ErrorItem{
			Field:   field,
			Message: field + " is " + strings.ToLower(fe.Tag()),
		})
	}

	return items
}

func BindPaginate[T any](ctx *fiber.Ctx, req *T, page *int, limit *int, defaultLimit int) (ok bool, err error) {
	if err := ctx.BodyParser(req); err != nil {
		return false, ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", nil)
	}

	if *page == 0 {
		*page = 1
	}
	if *limit == 0 {
		*limit = defaultLimit
	}

	if err := validate.Struct(req); err != nil {
		return false, ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid request", FormatValidationError(err))
	}

	return true, nil
}

func MapToResponseList[T any, R any](items []T, mapper func(T) R) []R {
	result := make([]R, 0, len(items))
	for _, item := range items {
		result = append(result, mapper(item))
	}

	return result
}

func MapToResponseListPaginate[T any, R any](items []T, total int64, page int, limit int, mapper func(T) R) ([]R, *PaginateMetadata) {
	totalPage := 0
	if limit > 0 {
		totalPage = int((total + int64(limit) - 1) / int64(limit))
	}

	return MapToResponseList(items, mapper), &PaginateMetadata{
		Page:      page,
		Size:      len(items),
		Total:     total,
		TotalPage: totalPage,
	}
}
