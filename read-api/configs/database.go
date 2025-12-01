package configs

import (
	"context"
	"fmt"
	"log"

	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	URI    string
	DBName string
}

func InitMongoDBConnection(cfg MongoConfig) (*mongo.Client, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("envriment MONGO_URI not set up")
	}

	clientOptions := options.Client().ApplyURI(cfg.URI)

	const maxRetries = 10
	const retryDelay = 5 * time.Second
	var _ *mongo.Client

	for i := 0; i < maxRetries; i++ {
		log.Printf("Tring to connect the MongoDB... Attempts %d/%d", i+1, maxRetries)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		client, err := mongo.Connect(ctx, clientOptions)
		cancel()

		if err == nil {
			ctxPing, cancelPing := context.WithTimeout(context.Background(), 2*time.Second)
			err = client.Ping(ctxPing, nil)
			cancelPing()

			if err == nil {
				log.Println("MongoDB connected successfully!")
				return client, nil
			}

			log.Printf("Ping to MongoDB failed. Trying again... Error: %v", err)
		} else {
			log.Printf("Initial connection to MongoDB failed. Trying again... Error: %v", err)
		}

		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("Failed to connect to and ping MongoDB after %d attempts.", maxRetries)
}

func GetCollection(client *mongo.Client, dbName, collectionName string, cfg MongoConfig) *mongo.Collection {
	if client == nil {
		log.Fatal("The MongoDB client has not been initialized.")
	}

	if dbName == "" {
		log.Fatal("MONGO_DB_NAME cannot be empty.")
	}

	return client.Database(dbName).Collection(collectionName)
}
