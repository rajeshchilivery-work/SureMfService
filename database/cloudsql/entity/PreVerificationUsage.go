package entity

import "time"

type PreVerificationUsage struct {
	ID                   int64      `json:"id" gorm:"primaryKey"`
	UUID                 string     `json:"uuid" gorm:"not null"`
	VerificationType     string     `json:"verification_type" gorm:"not null"`
	FpPreVerificationID  *string    `json:"fp_pre_verification_id"`
	Pan                  string     `json:"pan" gorm:"not null"`
	Status               string     `json:"status" gorm:"default:'pending'"`
	Result               *string    `json:"result"`
	BankIFSC             *string    `json:"bank_ifsc"`
	BankAccountNumber    *string    `json:"bank_account_number"`
	PollCount            int16      `json:"poll_count" gorm:"default:0"`
	CreatedAt            time.Time  `json:"created_at"`
	TriggeredBy          *string    `json:"triggered_by"`
}

func (PreVerificationUsage) TableName() string {
	return "sure_mf.pre_verification_usage"
}
