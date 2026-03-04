package repository

import (
	"SureMFService/database/cloudsql"
	"SureMFService/database/cloudsql/entity"
)

func CreateMfEvent(event *entity.MfEvent) error {
	return cloudsql.DB.Create(event).Error
}

func GetMfEventsByUserID(userID string) ([]entity.MfEvent, error) {
	var events []entity.MfEvent
	err := cloudsql.DB.Where("user_id = ?", userID).Order("event_at DESC").Find(&events).Error
	return events, err
}
