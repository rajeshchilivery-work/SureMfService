package entity

import "time"

// SureUser maps to sure_user.users — read-only, owned by SureCommonService
type SureUser struct {
	ID          int64     `gorm:"primaryKey"`
	UUID        string    `gorm:"column:uuid"`
	Name        string    `gorm:"column:name"`
	PhoneNumber string    `gorm:"column:phone_number"`
	PAN         string    `gorm:"column:pan"`
	DOB         time.Time `gorm:"column:dob"`
	Email       string    `gorm:"column:email"`
	Gender         string    `gorm:"column:gender"`
	HasGmailAccess bool      `gorm:"column:has_gmail_access"`
}

func (SureUser) TableName() string {
	return "sure_user.users"
}
