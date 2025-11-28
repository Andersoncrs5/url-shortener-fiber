package configs

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var DB *pgx.Conn

func ConnectDB() *pgx.Conn {
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

		PG_URL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, PG_URL)
	if err != nil {
		log.Fatalf("Falha ao conectar ao PostgreSQL: %v", err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		conn.Close(context.Background())
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	DB = conn
	fmt.Println("PostgreSQL connected successfully!")
	return DB
}

func CloseDB() {
	if DB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := DB.Close(ctx)
		if err != nil {
			log.Printf("Error closing the connection to PostgreSQL.: %v", err)
		} else {
			fmt.Println("Connection to PostgreSQL closed.")
		}
	}
}
