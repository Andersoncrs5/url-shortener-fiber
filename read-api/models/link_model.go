package models

import "time"

type Link struct {
	ID         int64      `json:"id" bson:"_id"`
	SHORT_CODE string     `json:"short_code" bson:"short_code"`
	LONG_URL   string     `json:"long_url" bson:"long_url"`
	CreatedAt  time.Time  `json:"created_at" bson:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at" bson:"expires_at"`
}
