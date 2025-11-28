package main

import (
	"linkfast/write-api/configs"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	configs.ConnectDB()

	app.Listen(":8888")
}
