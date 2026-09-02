package main

import (
	"gorm.io/gorm"

	"github.com/BernardBerenes/SupplyHub-API/internal/config"
	"github.com/BernardBerenes/SupplyHub-API/internal/users"
)

type Container struct {
	UserHandler *users.Handler
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	userRepo := users.NewRepository(db)
	userUseCase := users.NewUseCase(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
	userHandler := users.NewHandler(userUseCase)

	return &Container{
		UserHandler: userHandler,
	}
}
