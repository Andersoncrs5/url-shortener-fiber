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
		return nil, fmt.Errorf("variável MONGO_URI não configurada")
	}

	clientOptions := options.Client().ApplyURI(cfg.URI)

	const maxRetries = 10
	const retryDelay = 5 * time.Second
	var _ *mongo.Client

	for i := 0; i < maxRetries; i++ {
		log.Printf("Tentando conectar ao MongoDB... Tentativa %d/%d", i+1, maxRetries)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		client, err := mongo.Connect(ctx, clientOptions)
		cancel()

		if err == nil {
			ctxPing, cancelPing := context.WithTimeout(context.Background(), 2*time.Second)
			err = client.Ping(ctxPing, nil)
			cancelPing()

			if err == nil {
				log.Println("MongoDB conectado com sucesso!")
				return client, nil
			}

			log.Printf("Ping do MongoDB falhou. Tentando novamente... Erro: %v", err)
		} else {
			log.Printf("Conexão inicial com MongoDB falhou. Tentando novamente... Erro: %v", err)
		}

		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("falha ao conectar e pingar o MongoDB após %d tentativas", maxRetries)
}

func GetCollection(client *mongo.Client, dbName, collectionName string) *mongo.Collection {
	if client == nil {
		log.Fatal("Cliente MongoDB não foi inicializado.")
	}

	if dbName == "" {
		log.Fatal("MONGO_DB_NAME não pode ser vazio.")
	}

	return client.Database(dbName).Collection(collectionName)
}
