package cdc

import (
	"encoding/json"
	"fmt"
	models "linkfast/url-projector/model"
	"log"
	"time"
)

const timeFormat = "2006-01-02T15:04:05Z"

type Envelope struct {
	Schema  struct{} `json:"schema"`
	Payload struct {
		Before   map[string]interface{} `json:"before"`
		After    map[string]interface{} `json:"after"`
		Source   map[string]interface{} `json:"source"`
		Op       string                 `json:"op"`
		TsMs     int64                  `json:"ts_ms"`
		Sequence string                 `json:"sequence"`
	} `json:"payload"`
}

func ParseToEnvelope(value []byte, envelope *Envelope) error {
	if err := json.Unmarshal(value, &envelope); err != nil {
		log.Printf("ERROR: Failed to deserialize Debezium JSON: %v. Message: %s", err, string(value))
		return err
	}
	return nil
}

func parseTime(raw interface{}, fieldName string) (time.Time, error) {
	if raw == nil {
		return time.Time{}, fmt.Errorf("campo obrigatório '%s' está nulo", fieldName)
	}

	timeStr, ok := raw.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("campo '%s' não é uma string, é %T", fieldName, raw)
	}

	t, err := time.Parse(timeFormat, timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("erro ao parsear data e hora de '%s': %w", fieldName, err)
	}
	return t, nil
}

func parseOptionalTime(raw interface{}, fieldName string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}

	t, err := parseTime(raw, fieldName)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func GetLinkFromAfter(envelope Envelope) (models.Link, error) {
	after := envelope.Payload.After
	link := models.Link{}

	if rawID, ok := after["id"].(float64); ok {
		link.ID = int64(rawID)
	} else if rawID, ok := after["id"].(int64); ok {
		link.ID = rawID
	} else {
		return models.Link{}, fmt.Errorf("ID inválido: esperado float64 ou int64, obteve %T", after["id"])
	}

	var ok bool
	if link.SHORT_CODE, ok = after["short_code"].(string); !ok {
		return models.Link{}, fmt.Errorf("short_code inválido: esperado string, obteve %T", after["short_code"])
	}

	if link.LONG_URL, ok = after["long_url"].(string); !ok {
		return models.Link{}, fmt.Errorf("long_url inválido: esperado string, obteve %T", after["long_url"])
	}

	createdAt, err := parseTime(after["created_at"], "created_at")
	if err != nil {
		return models.Link{}, err
	}
	link.CreatedAt = createdAt

	expiresAt, err := parseOptionalTime(after["expires_at"], "expires_at")
	if err != nil {
		return models.Link{}, err
	}
	link.ExpiresAt = expiresAt

	return link, nil
}
