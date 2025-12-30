package models

import "time"

type User struct {
	ID                int64     `json:"id"`
	Username          string    `json:"username"`
	PasswordHash      string    `json:"-"` // Never serialize password hash
	RecoveryUUID      *string   `json:"-"`
	RecoverySecretHash *string   `json:"-"`
	CreatedAt         time.Time `json:"created_at"`
}

