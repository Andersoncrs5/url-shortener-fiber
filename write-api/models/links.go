package models

import "time"

type Links struct {
	ID         int64      `json:"id" gorm:"primaryKey;type:bigint;not null"`
	SHORT_CODE string     `json:"short_code" gorm:"type:varchar(10);uniqueIndex;not null"`
	LONG_URL   string     `json:"long_url" gorm:"type:text;not null"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
}
