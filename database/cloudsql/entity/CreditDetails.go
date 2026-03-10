package entity

// CreditDetails maps to sure_credit_report.credit_details — read-only
type CreditDetails struct {
	ID     int64 `gorm:"primaryKey"`
	UserID int   `gorm:"column:user_id"`
	Score  int64 `gorm:"column:score"`
}

func (CreditDetails) TableName() string {
	return "sure_credit_report.credit_details"
}
