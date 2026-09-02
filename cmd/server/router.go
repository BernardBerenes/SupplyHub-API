package main

import (
	"github.com/gofiber/fiber/v2"

	"github.com/BernardBerenes/SupplyHub-API/internal/middleware"
)

func RegisterRoutes(app *fiber.App, c *Container) {
	api := app.Group("/api/v1")

	api.Post("/login", c.UserHandler.Login)

	auth := middleware.JWTAuth(c.JWTSecret)

	products := api.Group("/products", auth)
	products.Post("", c.ProductHandler.Create)
	products.Get("", c.ProductHandler.List)
	products.Get("/:uuid", c.ProductHandler.Detail)
	products.Post("/paginate", c.ProductHandler.Paginate)
	products.Patch("/:uuid", c.ProductHandler.Update)
	products.Delete("/:uuid", c.ProductHandler.Delete)
}
