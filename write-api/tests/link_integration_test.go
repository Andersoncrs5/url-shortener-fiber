package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"linkfast/write-api/dtos"
	"linkfast/write-api/handlers"
	"linkfast/write-api/models"
	"linkfast/write-api/repositories"
	"linkfast/write-api/services"
)

func setupApp() (*fiber.App, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Falha ao conectar ao banco de dados de teste: %v", err)
	}

	if err := db.AutoMigrate(&models.Links{}); err != nil {
		log.Fatalf("Falha ao executar a migração de teste: %v", err)
	}

	linkRepository := repositories.NewLinkRepository(db)
	linkService := services.NewLinkService(linkRepository)
	linkHandler := handlers.NewLinkHandler(linkService)

	app := fiber.New()

	app.Use(requestid.New(requestid.Config{
		ContextKey: "trace_id",
	}))

	v1 := app.Group("/v1")
	v1.Post("/links", linkHandler.Create)
	v1.Get("/links/:id", linkHandler.GetByID)
	v1.Delete("/links/:id", linkHandler.Delete)
	v1.Get("/codes/:code", linkHandler.GetByShotCode)

	return app, db
}

func TestLinkHandler_Create_Integration(t *testing.T) {
	app, db := setupApp()

	tests := []struct {
		description  string
		payload      dtos.CreateLinkDto
		expectedCode int
		shouldExist  bool
	}{
		{
			description: "Sucesso: Criação de link sem data de expiração (opcional)",
			payload: dtos.CreateLinkDto{
				LONG_URL: "https://www.google.com/test-url-1",
			},
			expectedCode: http.StatusCreated,
			shouldExist:  true,
		},
		{
			description: "Sucesso: Criação de link com data de expiração futura",
			payload: dtos.CreateLinkDto{
				LONG_URL:  "https://www.google.com/test-url-2",
				ExpiresAt: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
			},
			expectedCode: http.StatusCreated,
			shouldExist:  true,
		},
		{
			description: "Falha: Data de expiração passada (Validação 'gt=now' do DTO)",
			payload: dtos.CreateLinkDto{
				LONG_URL:  "https://www.google.com/test-url-3",
				ExpiresAt: func() *time.Time { t := time.Now().Add(-1 * time.Hour); return &t }(),
			},
			expectedCode: http.StatusBadRequest,
			shouldExist:  false,
		},
		{
			description: "Falha: URL ausente (Validação 'required' do DTO)",
			payload: dtos.CreateLinkDto{
				LONG_URL:  "",
				ExpiresAt: nil,
			},
			expectedCode: http.StatusBadRequest,
			shouldExist:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			payloadBytes, _ := json.Marshal(test.payload)

			req := httptest.NewRequest(http.MethodPost, "/v1/links", bytes.NewReader(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, 3000)
			if err != nil {
				t.Fatalf("Erro ao executar a requisição: %v", err)
			}

			bodyBytes, readErr := io.ReadAll(resp.Body)
			resp.Body.Close() // Fecha o corpo da resposta após a leitura
			if readErr != nil {
				t.Fatalf("Falha ao ler o corpo da resposta: %v", readErr)
			}

			if resp.StatusCode != test.expectedCode {
				t.Errorf("Status code esperado: %d, obtido: %d. Corpo da resposta: %s",
					test.expectedCode, resp.StatusCode, bodyBytes)
			}

			if test.shouldExist {
				var response struct {
					Payload dtos.LinkDto `json:"payload"`
				}

				if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&response); err != nil {
					t.Fatalf("Falha ao decodificar a resposta JSON: %v. Body: %s", err, bodyBytes)
				}
				shortCode := response.Payload.SHORT_CODE

				var link models.Links
				result := db.Where("short_code = ?", shortCode).First(&link)

				if result.Error != nil {
					t.Errorf("Falha na verificação de integração: link com ShortCode %s não foi encontrado no DB. Erro: %v", shortCode, result.Error)
				}

				if link.LONG_URL != test.payload.LONG_URL {
					t.Errorf("Verificação de dados falhou: URL esperada %s, obtida %s.", test.payload.LONG_URL, link.LONG_URL)
				}
			}
		})
	}
}

func TestLinkHandler_GetByID_Integration(t *testing.T) {
	app, db := setupApp()
	repo := repositories.NewLinkRepository(db)

	validLinkToFind := models.Links{
		LONG_URL: "https://www.example.com/find-me",
	}
	createdLink, err := repo.Create(validLinkToFind)
	if err != nil {
		t.Fatalf("Setup falhou: não foi possível criar o link no DB: %v", err)
	}

	tests := []struct {
		description  string
		id           string
		expectedCode int
		expectedID   int64
	}{
		{
			description:  "Sucesso: Busca por ID válido",
			id:           fmt.Sprintf("%d", createdLink.ID),
			expectedCode: http.StatusOK,
			expectedID:   createdLink.ID,
		},
		{
			description:  "Falha: ID não encontrado (404)",
			id:           "999999",
			expectedCode: http.StatusNotFound,
			expectedID:   0,
		},
		{
			description:  "Falha: ID inválido (não numérico)",
			id:           "abc",
			expectedCode: http.StatusBadRequest,
			expectedID:   0,
		},
		{
			description:  "Falha: ID zero ou negativo",
			id:           "0",
			expectedCode: http.StatusBadRequest,
			expectedID:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/links/"+test.id, nil)

			resp, err := app.Test(req, 3000)
			if err != nil {
				t.Fatalf("Erro ao executar a requisição: %v", err)
			}

			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != test.expectedCode {
				t.Errorf("Status code esperado: %d, obtido: %d. Corpo da resposta: %s",
					test.expectedCode, resp.StatusCode, bodyBytes)
				return
			}

			if test.expectedCode == http.StatusOK {
				var response struct {
					Payload dtos.LinkDto `json:"payload"`
				}

				if err := json.Unmarshal(bodyBytes, &response); err != nil {
					t.Fatalf("Falha ao decodificar a resposta JSON: %v. Body: %s", err, bodyBytes)
				}

				if response.Payload.ID != test.expectedID {
					t.Errorf("ID esperado: %d, obtido: %d", test.expectedID, response.Payload.ID)
				}
			}
		})
	}
}

func TestLinkHandler_GetByShotCode_Integration(t *testing.T) {
	app, db := setupApp()
	repo := repositories.NewLinkRepository(db)

	validLinkToFind := models.Links{
		LONG_URL: "https://www.example.com/find-by-code",
	}
	createdLink, err := repo.Create(validLinkToFind)
	if err != nil {
		t.Fatalf("Setup falhou: não foi possível criar o link no DB: %v", err)
	}

	tests := []struct {
		description  string
		code         string
		expectedCode int
		expectedURL  string
	}{
		{
			description:  "Sucesso: Busca por Short Code válido",
			code:         createdLink.SHORT_CODE,
			expectedCode: http.StatusOK,
			expectedURL:  createdLink.LONG_URL,
		},
		{
			description:  "Falha: Short Code não encontrado (404)",
			code:         "INVALIDCODE",
			expectedCode: http.StatusNotFound,
			expectedURL:  "",
		},
		{
			description:  "Falha: Short Code vazio",
			code:         "",
			expectedCode: http.StatusNotFound,
			expectedURL:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/codes/"+test.code, nil)

			resp, err := app.Test(req, 3000)
			if err != nil {
				t.Fatalf("Erro ao executar a requisição: %v", err)
			}

			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != test.expectedCode {
				t.Errorf("Status code esperado: %d, obtido: %d. Corpo da resposta: %s",
					test.expectedCode, resp.StatusCode, bodyBytes)
				return
			}

			if test.expectedCode == http.StatusOK {
				var response struct {
					Payload dtos.LinkDto `json:"payload"`
				}

				if err := json.Unmarshal(bodyBytes, &response); err != nil {
					t.Fatalf("Falha ao decodificar a resposta JSON: %v. Body: %s", err, bodyBytes)
				}

				if response.Payload.LONG_URL != test.expectedURL {
					t.Errorf("URL esperada: %s, obtida: %s", test.expectedURL, response.Payload.LONG_URL)
				}
			}
		})
	}
}

func TestLinkHandler_Delete_Integration(t *testing.T) {
	app, db := setupApp()
	repo := repositories.NewLinkRepository(db)

	linkToDelete := models.Links{
		LONG_URL: "https://www.example.com/to-be-deleted",
	}
	createdLink, err := repo.Create(linkToDelete)
	if err != nil {
		t.Fatalf("Setup falhou: não foi possível criar o link para exclusão: %v", err)
	}

	tests := []struct {
		description  string
		id           string
		expectedCode int
	}{
		{
			description:  "Sucesso: Exclusão de link por ID válido",
			id:           fmt.Sprintf("%d", createdLink.ID),
			expectedCode: http.StatusOK,
		},
		{
			description:  "Falha: Tentativa de excluir ID inexistente (404)",
			id:           "999999",
			expectedCode: http.StatusNotFound,
		},
		{
			description:  "Falha: ID inválido (não numérico)",
			id:           "invalid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/v1/links/"+test.id, nil)

			resp, err := app.Test(req, 3000)
			if err != nil {
				t.Fatalf("Erro ao executar a requisição: %v", err)
			}

			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != test.expectedCode {
				t.Errorf("Status code esperado: %d, obtido: %d. Corpo da resposta: %s",
					test.expectedCode, resp.StatusCode, bodyBytes)
				return
			}

			if test.description == "Sucesso: Exclusão de link por ID válido" {
				var checkLink models.Links
				// Tenta buscar o link excluído diretamente no DB
				result := db.First(&checkLink, createdLink.ID)

				// Esperamos um erro (ErrRecordNotFound) se a exclusão funcionou
				if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
					t.Errorf("O link deveria ter sido excluído, mas ainda foi encontrado no DB.")
				}
			}
		})
	}
}
