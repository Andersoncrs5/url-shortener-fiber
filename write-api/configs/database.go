package configs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"linkfast/write-api/models"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func getEnvWithFallback(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

type DebeziumConnectorConfig struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

func ConnectDB() *gorm.DB {
	PG_URL := getEnvWithFallback("POSTGRES_URL", "")

	if PG_URL == "" {

		user := getEnvWithFallback("PG_USER", "")
		password := getEnvWithFallback("PG_PASSWORD", "")
		host := getEnvWithFallback("PG_HOST", "")
		port := getEnvWithFallback("PG_PORT", "")
		dbname := getEnvWithFallback("PG_DBNAME", "")

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
		log.Printf("Failed to connect to PostgreSQL using GORM: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to retrieve the SQL DB: %v", err)
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

func ConfiguredCDC(db *gorm.DB) error {
	fmt.Println("Configuring CDC users and permissions...")

	err := db.Exec("CREATE USER debezium;").Error
	if err != nil && err.Error() != "ERROR: role \"debezium\" already exists (SQLSTATE 42710)" {
		log.Printf("Failed to create user debezium.: %v", err)
	}

	err = db.Exec("ALTER USER debezium WITH PASSWORD '12345678' REPLICATION LOGIN;").Error
	if err != nil {
		log.Printf("Failed to configure the replication password and login for debezium.: %v", err)
	}

	err = db.Exec("CREATE ROLE replication_group;").Error
	if err != nil && err.Error() != "ERROR: role \"replication_group\" already exists (SQLSTATE 42710)" {
		log.Printf("Failed to create the replication_group role.: %v", err)
	}

	db.Exec("GRANT replication_group TO postgres;")
	err = db.Exec("GRANT replication_group TO debezium;").Error
	if err != nil {
		log.Printf("Failed to grant role 'replication_group' to 'debezium': %v", err)
	}

	err = db.Exec("GRANT USAGE ON SCHEMA link_fast_sc TO replication_group;").Error
	if err != nil {
		log.Printf("Failed to grant USAGE on link_fast_sc for replication_group: %v", err)
	}

	err = db.Exec("GRANT SELECT ON ALL TABLES IN SCHEMA link_fast_sc TO replication_group;").Error
	if err != nil {
		log.Printf("Warning: Failed to grant SELECT on existing tables (can be ignored if this is the first migration): %v", err)
	}
	err = db.Exec("ALTER DEFAULT PRIVILEGES IN SCHEMA link_fast_sc GRANT SELECT ON TABLES TO replication_group;").Error
	if err != nil {
		log.Printf("Failed to grant SELECT on future tables.: %v", err)
	}

	fmt.Println("CDC users and permissions configured.")
	return nil
}

func PostMigrationSetup(db *gorm.DB) error {
	fmt.Println("Running post-migration CDC setup...")

	err := db.Exec("CREATE PUBLICATION dbz_publication FOR TABLE link_fast_sc.links;").Error
	if err != nil && err.Error() != "ERROR: publication \"dbz_publication\" already exists (SQLSTATE 42710)" {
		log.Printf("Failed to create PUBLICATION for links.: %v", err)
	}

	fmt.Println("Setting table owner...")
	if err := db.Exec("ALTER TABLE link_fast_sc.links OWNER TO replication_group;").Error; err != nil {
		log.Printf("Failed to change OWNER in the links table.: %v", err)
	}

	fmt.Println("Post-migration setup complete.")
	return nil
}
func RegisterDebeziumConnector() error {
	DATABASE_HOSTNAME := getEnvWithFallback("DATABASE_HOSTNAME", "")
	USER_CDC := getEnvWithFallback("USER_CDC", "")
	USER_PASSWORD_CDC := getEnvWithFallback("USER_PASSWORD_CDC", "")
	LINK_DB := getEnvWithFallback("LINK_DB", "")
	PG_PORT := getEnvWithFallback("PG_PORT", "")
	API_URL_CONNECT := getEnvWithFallback("API_URL_CONNECT", "")
	TOPIC_PREFIX := getEnvWithFallback("TOPIC_PREFIX", "")
	PARTITIONS_STR := getEnvWithFallback("PARTITIONS_STR", "")
	BATCH_SIZE := getEnvWithFallback("BATCH_SIZE", "")
	QUEUE_SIZE := getEnvWithFallback("QUEUE_SIZE", "")

	PARTITIONS_INT, err_parse := strconv.Atoi(PARTITIONS_STR)

	required := map[string]string{
		"DATABASE_HOSTNAME": DATABASE_HOSTNAME,
		"USER_CDC":          USER_CDC,
		"USER_PASSWORD_CDC": USER_PASSWORD_CDC,
		"LINK_DB":           LINK_DB,
		"PG_PORT":           PG_PORT,
		"API_URL_CONNECT":   API_URL_CONNECT,
		"TOPIC_PREFIX":      TOPIC_PREFIX,
		"PARTITIONS_STR":    PARTITIONS_STR,
		"BATCH_SIZE":        BATCH_SIZE,
		"QUEUE_SIZE":        QUEUE_SIZE,
	}

	for key, value := range required {
		if value == "" {
			log.Fatalf("Environment variable %s not defined!", key)
		}
	}

	if err_parse != nil {
		log.Fatalf("Error parse PARTITIONS_STR to PARTITIONS_INT")
	}

	log.Println("All required environment variables to cdc loaded successfully!")

	connectorConfig := DebeziumConnectorConfig{
		Name: "postgres-cdc",
		Config: map[string]interface{}{
			"connector.class":   "io.debezium.connector.postgresql.PostgresConnector",
			"database.hostname": DATABASE_HOSTNAME,
			"database.port":     PG_PORT,
			"database.user":     USER_CDC,
			"database.password": USER_PASSWORD_CDC,
			"database.dbname":   LINK_DB,

			"plugin.name": "pgoutput",
			"slot.name":   "debezium_slot",

			"schema.include.list": "link_fast_sc",
			"table.include.list":  "link_fast_sc.links",

			"topic.prefix": TOPIC_PREFIX,

			"tombstones.on.delete":   "false",
			"include.schema.changes": "false",

			"publication.name": "dbz_publication",

			"decimal.handling.mode":             "string",
			"hstore.handling.mode":              "json",
			"topic.creation.default.partitions": PARTITIONS_INT,

			"topic.creation.enable": true,
			"max.batch.size":        BATCH_SIZE,
			"max.queue.size":        QUEUE_SIZE,
		},
	}

	body, err_parse := json.Marshal(connectorConfig)
	if err_parse != nil {
		log.Printf("Error serializing connector configuration.: %v", err_parse)
	}

	req, err_parse := http.NewRequest("POST", API_URL_CONNECT, bytes.NewBuffer(body))
	if err_parse != nil {
		log.Printf("Error creating HTTP request.: %v", err_parse)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err_parse := client.Do(req)
	if err_parse != nil {
		log.Printf("Error sending request to Kafka Connect API. Verify that Kafka Connect is running and accessible at %s: %v", API_URL_CONNECT, err_parse)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)

		if msg, ok := errorBody["message"].(string); ok && resp.StatusCode == 409 {
			fmt.Printf("Warning: Debezium connector already registered. Status: %s\n", msg)
			return nil
		}

		log.Printf("Failed to register Debezium connector. Status: %s. Error: %v", resp.Status, errorBody)
	}

	fmt.Println("Debezium connector successfully registered!")
	return nil
}

func Migrate(db *gorm.DB) error {
	fmt.Println("Ensuring schema exists...")
	err := db.Exec("CREATE SCHEMA IF NOT EXISTS link_fast_sc;").Error
	if err != nil {
		log.Printf("failed to create schema 'link_fast_sc': %v", err)
	}

	ConfiguredCDC(db)

	log.Println("Running migrations...")
	db.AutoMigrate(&models.Links{})
	PostMigrationSetup(db)

	RegisterDebeziumConnector()
	return nil
}
