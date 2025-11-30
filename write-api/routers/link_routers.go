package routers

import (
	"linkfast/write-api/handlers"

	"github.com/gofiber/fiber/v2"
)

func LinkRoute(app *fiber.App, linkHandler handlers.LinkHandler) {
	router := app.Group("/api/v1/links")

	router.Get("/:id", linkHandler.GetByID)
	router.Get("/:code/code", linkHandler.GetByShotCode)
	router.Post("", linkHandler.Create)
	router.Delete("/:id", linkHandler.Delete)
}
