package main

import (
	"linkfast/write-api/configs"
	"linkfast/write-api/handlers"
	"linkfast/write-api/repositories"
	"linkfast/write-api/routers"
	"linkfast/write-api/services"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	configs.ConnectDB()
	db := configs.DB

	linkRepository := repositories.NewLinkRepository(db)
	linkService := services.NewUserService(linkRepository)
	linkHandler := handlers.NewTaskHandler(linkService)

	routers.LinkRoute(app, linkHandler)

	app.Listen(":8888")
}
