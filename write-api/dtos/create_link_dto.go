package dtos

import "time"

type CreateLinkDto struct {
	LONG_URL  string     `json:"long_url" validate:"required,min=8,max=2500"`
	ExpiresAt *time.Time `json:"expires_at" validate:"omitempty,gt=now"`
}
