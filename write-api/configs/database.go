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
	Name   string            `json:"name"`
	Config map[string]string `json:"config"`
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

func ConfiguredCDC(db *gorm.DB) error {
	fmt.Println("Configuring CDC users and permissions...")

	err := db.Exec("CREATE USER debezium;").Error
	if err != nil && err.Error() != "ERROR: role \"debezium\" already exists (SQLSTATE 42710)" {
		log.Printf("Falha ao criar o usuário debezium: %v", err)
	}

	err = db.Exec("ALTER USER debezium WITH PASSWORD '12345678' REPLICATION LOGIN;").Error
	if err != nil {
		log.Printf("Falha ao configurar a senha e login de replicação para debezium: %v", err)
	}

	err = db.Exec("CREATE ROLE replication_group;").Error
	if err != nil && err.Error() != "ERROR: role \"replication_group\" already exists (SQLSTATE 42710)" {
		log.Printf("Falha ao criar o role replication_group: %v", err)
	}

	db.Exec("GRANT replication_group TO postgres;")
	err = db.Exec("GRANT replication_group TO debezium;").Error
	if err != nil {
		log.Printf("Falha ao conceder role 'replication_group' ao 'debezium': %v", err)
	}

	err = db.Exec("GRANT USAGE ON SCHEMA link_fast_sc TO replication_group;").Error
	if err != nil {
		log.Printf("Falha ao conceder USAGE em link_fast_sc para replication_group: %v", err)
	}

	err = db.Exec("GRANT SELECT ON ALL TABLES IN SCHEMA link_fast_sc TO replication_group;").Error
	if err != nil {
		log.Printf("Aviso: Falha ao conceder SELECT em tabelas existentes (pode ser ignorado se for a primeira migração): %v", err)
	}
	err = db.Exec("ALTER DEFAULT PRIVILEGES IN SCHEMA link_fast_sc GRANT SELECT ON TABLES TO replication_group;").Error
	if err != nil {
		log.Printf("Falha ao conceder SELECT em tabelas futuras: %v", err)
	}

	fmt.Println("CDC users and permissions configured.")
	return nil
}

func PostMigrationSetup(db *gorm.DB) error {
	fmt.Println("Running post-migration CDC setup...")

	err := db.Exec("CREATE PUBLICATION dbz_publication FOR TABLE link_fast_sc.links;").Error
	if err != nil && err.Error() != "ERROR: publication \"dbz_publication\" already exists (SQLSTATE 42710)" {
		log.Printf("Falha ao criar PUBLICATION para links: %v", err)
	}

	fmt.Println("Setting table owner...")
	if err := db.Exec("ALTER TABLE link_fast_sc.links OWNER TO replication_group;").Error; err != nil {
		log.Printf("Falha ao alterar OWNER da tabela links: %v", err)
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

	connectorConfig := DebeziumConnectorConfig{
		Name: "postgres-cdc",
		Config: map[string]string{
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

			"topic.prefix": "pgserver1",

			"tombstones.on.delete":   "false",
			"include.schema.changes": "false",

			"publication.name": "dbz_publication",
		},
	}

	body, err := json.Marshal(connectorConfig)
	if err != nil {
		log.Printf("erro ao serializar config do conector: %v", err)
	}

	req, err := http.NewRequest("POST", API_URL_CONNECT, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("erro ao criar requisição HTTP: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("erro ao enviar requisição para Kafka Connect API. Verifique se o Kafka Connect está rodando e acessível em %s: %v", API_URL_CONNECT, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)

		if msg, ok := errorBody["message"].(string); ok && resp.StatusCode == 409 {
			fmt.Printf("Aviso: Conector Debezium já registrado. Status: %s\n", msg)
			return nil
		}

		log.Printf("falha ao registrar conector Debezium. Status: %s. Erro: %v", resp.Status, errorBody)
	}

	fmt.Println("Conector Debezium registrado com sucesso!")
	return nil
}

func Migrate(db *gorm.DB) error {
	fmt.Println("Ensuring schema exists...")
	err := db.Exec("CREATE SCHEMA IF NOT EXISTS link_fast_sc;").Error
	if err != nil {
		log.Printf("failed to create schema 'link_fast_sc': %v", err)
	}

	if err := ConfiguredCDC(db); err != nil {
		return err
	}

	fmt.Println("Running migrations...")
	if err := db.AutoMigrate(&models.Links{}); err != nil {
		log.Printf("Error the auto migrate: %v", err)
	}

	if err := PostMigrationSetup(db); err != nil {
		log.Printf("Error the PostMigrationSetup: %v", err)
	}

	if err := RegisterDebeziumConnector(); err != nil {
		log.Printf("erro ao registrar conector CDC: %v", err)
	}

	return nil
}
