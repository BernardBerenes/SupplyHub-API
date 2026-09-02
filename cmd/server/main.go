package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/BernardBerenes/SupplyHub-API/internal/config"
	"github.com/BernardBerenes/SupplyHub-API/internal/database"
	"github.com/BernardBerenes/SupplyHub-API/internal/middleware"
	"github.com/BernardBerenes/SupplyHub-API/internal/users"
)

func main() {
	cfg := config.Load()

	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = database.Migrate(db, &users.User{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	app := fiber.New()
	app.Use(middleware.Recovery)

	container := NewContainer(db, cfg)
	RegisterRoutes(app, container)

	log.Fatal(app.Listen(":" + cfg.Port))
}
