package dtos

import "time"

type CreateLinkDto struct {
	LONG_URL  string     `json:"long_url"`
	ExpiresAt *time.Time `json:"expires_at"`
}
