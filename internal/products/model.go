package products

import "time"

type Product struct {
	ID        string     `json:"id" gorm:"type:uuid;primaryKey"`
	Name      string     `json:"name" gorm:"size:100;not null"`
	Price     int64      `json:"price" gorm:"not null"`
	Photo     *string    `json:"photo" gorm:"type:text"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`
}

type PhotoUpload struct {
	Data        []byte
	Extension   string
	ContentType string
}

type CreateInput struct {
	Name  string
	Price int64
	Photo *PhotoUpload
}

type UpdateInput struct {
	Name  *string
	Price *int64
	Photo *PhotoUpload
}

type PaginateRequest struct {
	Page  int    `json:"page" validate:"min=1"`
	Limit int    `json:"limit" validate:"oneof=10 25 50 100"`
	Name  string `json:"name"`
}

type ProductResponse struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price int64   `json:"price"`
	Photo *string `json:"photo"`
}

type ListResponse struct {
	Products []ProductResponse `json:"products"`
}

type DetailResponse struct {
	Product ProductResponse `json:"product"`
}

type PaginateResponse struct {
	Page      int               `json:"page"`
	Size      int               `json:"size"`
	TotalItem int64             `json:"total_item"`
	TotalPage int               `json:"total_page"`
	Products  []ProductResponse `json:"products"`
}

func ToResponse(p Product) ProductResponse {
	return ProductResponse{
		ID:    p.ID,
		Name:  p.Name,
		Price: p.Price,
		Photo: p.Photo,
	}
}
