package main

import (
	"context"
	configs "linkfast/url-projector/config"
	"linkfast/url-projector/consumer"
	"linkfast/url-projector/repositories"
	"linkfast/url-projector/services"
	"linkfast/url-projector/utils/envs"
	"log"
	"time"
)

func main() {
	log.Println("Iniciando microserviço....")

	log.Println("Waiting the anothers container startup")
	time.Sleep(15 * time.Second)

	kafkaBrokers := envs.GetEnvWithFallback("KAFKA_BROKERS", "")
	kafkaTopic := envs.GetEnvWithFallback("KAFKA_TOPIC", "")
	mongoURI := envs.GetEnvWithFallback("MONGO_URI", "")
	mongoDBName := envs.GetEnvWithFallback("MONGO_DB_NAME", "")

	required := map[string]string{
		"KAFKA_BROKERS": kafkaBrokers,
		"KAFKA_TOPIC":   kafkaTopic,
		"MONGO_URI":     mongoURI,
		"MONGO_DB_NAME": mongoDBName,
	}

	for key, value := range required {
		if value == "" {
			log.Fatalf("Environment variable %s not defined!", key)
		}
	}

	mongoCfg := configs.MongoConfig{
		URI:    mongoURI,
		DBName: mongoDBName,
	}
	mongoClient, err := configs.InitMongoDBConnection(mongoCfg)
	if err != nil {
		log.Fatalf("Falha crítica ao conectar ao MongoDB: %v", err)
	}

	defer func() {
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	mongoDB := mongoClient.Database(mongoDBName)

	linkRepo := repositories.NewLinkRepository(mongoDB)

	linkService := services.NewLinkService(linkRepo)

	log.Println("Reading topics!")
	consumer.LinkConsumer(kafkaBrokers, kafkaTopic, linkService)
}
