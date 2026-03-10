package entity

// InterestRateV2 maps to sure_credit_report.interest_rates_v2 — read-only
type InterestRateV2 struct {
	ID            int64   `gorm:"primaryKey"`
	MinScore      int64   `gorm:"column:min_score"`
	MaxScore      int64   `gorm:"column:max_score"`
	MarketRate    float64 `gorm:"column:market_rate"`
	AccountTypeID int64   `gorm:"column:account_type_id"`
	IsActive      bool    `gorm:"column:is_active"`
}

func (InterestRateV2) TableName() string {
	return "sure_credit_report.interest_rates_v2"
}
