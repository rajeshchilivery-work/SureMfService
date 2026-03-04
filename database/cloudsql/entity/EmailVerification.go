package entity

import "time"

type EmailVerification struct {
	ID           int64      `json:"id" gorm:"primaryKey"`
	UUID         string     `json:"uuid" gorm:"not null"`
	Email        string     `json:"email" gorm:"not null"`
	Method       string     `json:"method" gorm:"default:'otp'"`
	TokenHash    *string    `json:"token_hash"`
	Status       string     `json:"status" gorm:"default:'pending'"`
	AttemptCount int16      `json:"attempt_count" gorm:"default:0"`
	MaxAttempts  int16      `json:"max_attempts" gorm:"default:3"`
	InitiatedAt  time.Time  `json:"initiated_at"`
	ExpiresAt    time.Time  `json:"expires_at" gorm:"not null"`
	VerifiedAt   *time.Time `json:"verified_at"`
}

func (EmailVerification) TableName() string {
	return "sure_mf.email_verification"
}
