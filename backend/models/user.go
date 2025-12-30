package models

import "time"

type User struct {
	ID                   int64      `json:"id"`
	Username             string     `json:"username"`
	Email                *string    `json:"email"`
	PasswordHash         string     `json:"-"` // Never serialize password hash
	RecoveryUUID         *string    `json:"-"`
	RecoverySecretHash   *string    `json:"-"`
	PasswordResetToken   *string    `json:"-"`
	PasswordResetExpires *time.Time `json:"-"`
	WeeklyReportEnabled  *bool      `json:"weekly_report_enabled"`
	TOTPSecret           *string    `json:"-"`
	TOTPEnabled          *bool      `json:"totp_enabled"`
	TOTPBackupCodes      *string    `json:"-"`
	CreatedAt            time.Time  `json:"created_at"`
}

