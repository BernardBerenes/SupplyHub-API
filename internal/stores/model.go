package stores

import "time"

type Store struct {
	ID        int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string     `json:"name" gorm:"size:100;not null"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`
}

type CreateInput struct {
	Name string
}

type UpdateInput struct {
	Name *string
}

type CreateRequest struct {
	Name string `json:"name"`
}

type UpdateRequest struct {
	Name *string `json:"name"`
}

type PaginateRequest struct {
	Page  int    `json:"page" validate:"min=1"`
	Limit int    `json:"limit" validate:"oneof=10 25 50 100"`
	Name  string `json:"name"`
}

type StoreResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ListResponse struct {
	Stores []StoreResponse `json:"stores"`
}

type PaginateResponse struct {
	Page      int             `json:"page"`
	Size      int             `json:"size"`
	TotalItem int64           `json:"total_item"`
	TotalPage int             `json:"total_page"`
	Stores    []StoreResponse `json:"stores"`
}

func ToResponse(s Store) StoreResponse {
	return StoreResponse{
		ID:   s.ID,
		Name: s.Name,
	}
}
