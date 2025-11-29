package configs

import (
	"context"
	"fmt"
	"linkfast/write-api/models"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		panic("Err the load file .env")
	}

	PG_URL := os.Getenv("POSTGRES_URL")

	if PG_URL == "" {

		user := os.Getenv("PG_USER")
		password := os.Getenv("PG_PASSWORD")
		host := os.Getenv("PG_HOST")
		port := os.Getenv("PG_PORT")
		dbname := os.Getenv("PG_DBNAME")

		if user == "" || password == "" || host == "" || port == "" || dbname == "" {
			log.Fatal("POSTGRES_URL or PG_USER/PG_PASSWORD/PG_HOST/PG_PORT/PG_DBNAME not defined in .env")
		}

		PG_URL = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo",
			host, user, password, dbname, port)
	}

	db, err := gorm.Open(postgres.Open(PG_URL), &gorm.Config{
		// Ex: Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Falha ao conectar ao PostgreSQL usando GORM: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Falha ao obter o SQL DB: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	fmt.Println("PostgreSQL connected successfully using GORM!")
	return DB
}

// Função de Migração mantida, agora usando o *gorm.DB
func Migrate(db *gorm.DB) error {
	fmt.Println("Running migrations...")
	return db.AutoMigrate(&models.Links{})
}

// CloseDB foi removida/comentada, pois GORM gerencia o ciclo de vida do pool.
