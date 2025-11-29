package dtos

import "time"

type LinkDto struct {
	ID         int64      `json:"id"`
	SHORT_CODE string     `json:"short_code"`
	LONG_URL   string     `json:"long_url"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
}
