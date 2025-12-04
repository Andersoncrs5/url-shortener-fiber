package main

import (
	"context"
	"linkfast/read-api/configs"
	"linkfast/read-api/handlers"
	"linkfast/read-api/repositories"
	"linkfast/read-api/routers"
	"linkfast/read-api/services"
	"linkfast/read-api/utils/envs"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func Main() {
	log.Print("Waiting the another container startup.........")
	time.Sleep(20 * time.Second)

	app := fiber.New()

	log.Printf("Configuring vars mongoURI and mongoDBName....")
	mongoURI := envs.GetEnvWithFallback("MONGO_URI", "")
	mongoDBName := envs.GetEnvWithFallback("MONGO_DB_NAME", "")

	required := map[string]string{
		"MONGO_URI":     mongoURI,
		"MONGO_DB_NAME": mongoDBName,
	}

	for key, value := range required {
		if value == "" {
			log.Fatalf("Environment variable %s not defined!", key)
		}
	}

	log.Printf("mongoURI and mongoDBName configured")

	mongoCfg := configs.MongoConfig{
		URI:    mongoURI,
		DBName: mongoDBName,
	}
	mongoClient, err := configs.InitMongoDBConnection(mongoCfg)
	if err != nil {
		log.Fatalf("Critical failure connecting to MongoDB.: %v", err)
	}

	defer func() {
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	app.Use(requestid.New(requestid.Config{
		ContextKey: "trace_id",
	}))

	mongoDB := mongoClient.Database(mongoDBName)
	linkRepo := repositories.NewLinkRepository(mongoDB)
	linkService := services.NewLinkService(linkRepo)
	linkHandler := handlers.NewLinkHandler(linkService)

	routers.LinkRoute(app, linkHandler)

	app.Listen(":8888")
}
