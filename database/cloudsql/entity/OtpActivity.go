package entity

import "time"

type OtpActivity struct {
	ID                  int64      `json:"id" gorm:"primaryKey"`
	UUID                string     `json:"uuid" gorm:"not null"`
	OrderType           string     `json:"order_type" gorm:"not null"`
	FpOrderID           string     `json:"fp_order_id" gorm:"not null"`
	Status              string     `json:"status" gorm:"default:'initiated'"`
	InitiatedAt         time.Time  `json:"initiated_at"`
	ConfirmedAt         *time.Time `json:"confirmed_at"`
	ResultingOrderState *string    `json:"resulting_order_state"`
}

func (OtpActivity) TableName() string {
	return "sure_mf.otp_activity"
}
