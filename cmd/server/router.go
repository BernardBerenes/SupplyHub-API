package main

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(app *fiber.App, c *Container) {
	api := app.Group("/api/v1")
	api.Post("/login", c.UserHandler.Login)
}
