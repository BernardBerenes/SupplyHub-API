package transactiondetails

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const (
	UNIT_PIECES = "PIECES"
	UNIT_DOZENS = "DOZENS"
	UNIT_BOX    = "BOX"
	UNIT_CARTON = "CARTON"
)

type ProductSnapshot struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (p ProductSnapshot) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *ProductSnapshot) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("unsupported type for ProductSnapshot: %T", value)
		}
		bytes = []byte(str)
	}

	return json.Unmarshal(bytes, p)
}

type TransactionDetail struct {
	ID            string          `json:"id" gorm:"type:uuid;primaryKey"`
	TransactionID string          `json:"transaction_id" gorm:"type:uuid;not null;index"`
	Product       ProductSnapshot `json:"product" gorm:"column:product;type:jsonb;not null"`
	Quantity      int64           `json:"quantity" gorm:"not null"`
	Unit          string          `json:"unit" gorm:"size:10;not null"`
	Price         int64           `json:"price" gorm:"not null"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     *time.Time      `json:"-" gorm:"index"`
}

type CreateInput struct {
	TransactionID string
	ProductID     string
	Quantity      int64
	Unit          string
	Price         int64
}

type UpdateInput struct {
	ProductID *string
	Quantity  *int64
	Unit      *string
	Price     *int64
}

type CreateRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int64  `json:"quantity"`
	Unit      string `json:"unit"`
	Price     int64  `json:"price"`
}

type UpdateRequest struct {
	ProductID *string `json:"product_id"`
	Quantity  *int64  `json:"quantity"`
	Unit      *string `json:"unit"`
	Price     *int64  `json:"price"`
}

type PaginateRequest struct {
	Page  int `json:"page" validate:"min=1"`
	Limit int `json:"limit" validate:"oneof=10 25 50 100"`
}

type TransactionDetailResponse struct {
	ID            string          `json:"id"`
	TransactionID string          `json:"transaction_id"`
	Product       ProductSnapshot `json:"product"`
	Quantity      int64           `json:"quantity"`
	Unit          string          `json:"unit"`
	Price         int64           `json:"price"`
}

type PaginateResponse struct {
	Page               int                         `json:"page"`
	Size               int                         `json:"size"`
	TotalItem          int64                       `json:"total_item"`
	TotalPage          int                         `json:"total_page"`
	TransactionDetails []TransactionDetailResponse `json:"transaction_details"`
}

func ToResponse(d TransactionDetail) TransactionDetailResponse {
	return TransactionDetailResponse{
		ID:            d.ID,
		TransactionID: d.TransactionID,
		Product:       d.Product,
		Quantity:      d.Quantity,
		Unit:          d.Unit,
		Price:         d.Price,
	}
}

func IsValidUnit(unit string) bool {
	switch unit {
	case UNIT_PIECES, UNIT_DOZENS, UNIT_BOX, UNIT_CARTON:
		return true
	default:
		return false
	}
}
