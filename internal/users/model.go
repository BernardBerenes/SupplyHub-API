package users

import "time"

type User struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Password  string     `json:"-" gorm:"size:100;not null"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}
