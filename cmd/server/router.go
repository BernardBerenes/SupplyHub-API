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

	stores := api.Group("/stores", auth)
	stores.Post("", c.StoreHandler.Create)
	stores.Get("", c.StoreHandler.List)
	stores.Post("/paginate", c.StoreHandler.Paginate)
	stores.Patch("/:uuid", c.StoreHandler.Update)
	stores.Delete("/:uuid", c.StoreHandler.Delete)

	transactions := api.Group("/transactions", auth)
	transactions.Post("", c.TransactionHandler.Create)
	transactions.Post("/paginate", c.TransactionHandler.Paginate)
	transactions.Post("/sync", c.TransactionHandler.Sync)
	transactions.Patch("/:uuid", c.TransactionHandler.Update)
	transactions.Delete("/:uuid", c.TransactionHandler.Delete)

	transactionDetails := transactions.Group("/:transaction_id/details")
	transactionDetails.Post("", c.TransactionDetailHandler.Create)
	transactionDetails.Post("/paginate", c.TransactionDetailHandler.Paginate)
	transactionDetails.Patch("/:uuid", c.TransactionDetailHandler.Update)
	transactionDetails.Delete("/:uuid", c.TransactionDetailHandler.Delete)
}
