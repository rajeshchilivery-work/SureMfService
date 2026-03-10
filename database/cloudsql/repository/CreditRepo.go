package repository

import (
	"SureMFService/database/cloudsql"
	"SureMFService/database/cloudsql/entity"
	"fmt"
)

func GetCreditDetailsByUserID(userID int64) (*entity.CreditDetails, error) {
	var credit entity.CreditDetails
	err := cloudsql.DB.Where("user_id = ?", userID).First(&credit).Error
	if err != nil {
		return nil, fmt.Errorf("credit details not found for user_id %d: %w", userID, err)
	}
	return &credit, nil
}

func GetMarketRate(accountTypeID int, score int64) (float64, error) {
	var rate entity.InterestRateV2
	err := cloudsql.DB.
		Where("account_type_id = ? AND min_score <= ? AND max_score >= ? AND is_active = ?", accountTypeID, score, score, true).
		First(&rate).Error
	if err != nil {
		return 0, fmt.Errorf("market rate not found for account_type_id %d, score %d: %w", accountTypeID, score, err)
	}
	return rate.MarketRate, nil
}
