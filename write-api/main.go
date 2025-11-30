package main

import (
	"linkfast/write-api/configs"
	"linkfast/write-api/handlers"
	"linkfast/write-api/repositories"
	"linkfast/write-api/routers"
	"linkfast/write-api/services"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	app := fiber.New()

	configs.ConnectDB()
	db := configs.DB

	app.Use(requestid.New(requestid.Config{
		ContextKey: "trace_id",
	}))

	configs.Migrate(db)

	linkRepository := repositories.NewLinkRepository(db)
	linkService := services.NewLinkService(linkRepository)
	linkHandler := handlers.NewLinkHandler(linkService)

	routers.LinkRoute(app, linkHandler)

	app.Listen(":8888")
}
