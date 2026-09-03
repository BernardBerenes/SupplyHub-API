package transactions

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const (
	PAYMENT_STATUS_PAID   = "PAID"
	PAYMENT_STATUS_UNPAID = "UNPAID"

	DELIVERY_STATUS_PENDING     = "PENDING"
	DELIVERY_STATUS_ON_DELIVERY = "ON_DELIVERY"
	DELIVERY_STATUS_DELIVERED   = "DELIVERED"
)

const DateFormat = "2006-01-02"

type StoreSnapshot struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (s StoreSnapshot) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StoreSnapshot) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("unsupported type for StoreSnapshot: %T", value)
		}
		bytes = []byte(str)
	}

	return json.Unmarshal(bytes, s)
}

type Transaction struct {
	ID             string        `json:"id" gorm:"type:uuid;primaryKey"`
	Store          StoreSnapshot `json:"store" gorm:"column:store;type:jsonb;not null"`
	PaymentStatus  string        `json:"payment_status" gorm:"size:10;not null"`
	DeliveryStatus string        `json:"delivery_status" gorm:"size:15;not null"`
	Date           time.Time     `json:"date" gorm:"type:date;not null"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	DeletedAt      *time.Time    `json:"-" gorm:"index"`
}

type CreateInput struct {
	StoreID int64
	Date    time.Time
}

type UpdateInput struct {
	StoreID        *int64
	PaymentStatus  *string
	DeliveryStatus *string
	Date           *time.Time
}

type CreateRequest struct {
	StoreID int64  `json:"store_id"`
	Date    string `json:"date"`
}

type UpdateRequest struct {
	StoreID        *int64  `json:"store_id"`
	PaymentStatus  *string `json:"payment_status"`
	DeliveryStatus *string `json:"delivery_status"`
	Date           *string `json:"date"`
}

type PaginateRequest struct {
	Page           int    `json:"page" validate:"min=1"`
	Limit          int    `json:"limit" validate:"oneof=10 25 50 100"`
	PaymentStatus  string `json:"payment_status" validate:"omitempty,oneof=PAID UNPAID"`
	DeliveryStatus string `json:"delivery_status" validate:"omitempty,oneof=PENDING ON_DELIVERY DELIVERED"`
	DateFrom       string `json:"date_from" validate:"omitempty,datetime=2006-01-02"`
	DateTo         string `json:"date_to" validate:"omitempty,datetime=2006-01-02"`
}

type TransactionResponse struct {
	ID             string        `json:"id"`
	Store          StoreSnapshot `json:"store"`
	PaymentStatus  string        `json:"payment_status"`
	DeliveryStatus string        `json:"delivery_status"`
	Date           string        `json:"date"`
}

type PaginateResponse struct {
	Page         int                   `json:"page"`
	Size         int                   `json:"size"`
	TotalItem    int64                 `json:"total_item"`
	TotalPage    int                   `json:"total_page"`
	Transactions []TransactionResponse `json:"transactions"`
}

func ToResponse(t Transaction) TransactionResponse {
	return TransactionResponse{
		ID:             t.ID,
		Store:          t.Store,
		PaymentStatus:  t.PaymentStatus,
		DeliveryStatus: t.DeliveryStatus,
		Date:           t.Date.Format(DateFormat),
	}
}
