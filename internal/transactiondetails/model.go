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
)

type ProductSnapshot struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
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
	PricePerUnit  int64           `json:"price_per_unit" gorm:"column:price_per_unit;not null;default:0"`
	TotalPrice    int64           `json:"total_price" gorm:"column:total_price;not null;default:0"`
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
	PricePerUnit  int64           `json:"price_per_unit"`
	TotalPrice    int64           `json:"total_price"`
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
		PricePerUnit:  d.PricePerUnit,
		TotalPrice:    d.TotalPrice,
	}
}

func IsValidUnit(unit string) bool {
	switch unit {
	case UNIT_PIECES, UNIT_DOZENS:
		return true
	default:
		return false
	}
}

// unitMultiplier is how many pieces a single unit of quantity represents. A
// DOZENS quantity counts dozens rather than pieces, so it expands by 12.
func unitMultiplier(unit string) int64 {
	if unit == UNIT_DOZENS {
		return 12
	}
	return 1
}

// CalculatePricing derives price_per_unit and total_price from the raw,
// always-per-piece price the client sends.
func CalculatePricing(price, quantity int64, unit string) (pricePerUnit, totalPrice int64) {
	pricePerUnit = price * unitMultiplier(unit)
	totalPrice = pricePerUnit * quantity
	return pricePerUnit, totalPrice
}

// RawPrice reverses CalculatePricing's price_per_unit back to the raw,
// per-piece price, so an update that only touches quantity or unit can
// recalculate pricing without the client resending price.
func RawPrice(pricePerUnit int64, unit string) int64 {
	return pricePerUnit / unitMultiplier(unit)
}
