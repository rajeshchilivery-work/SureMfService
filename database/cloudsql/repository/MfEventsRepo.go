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

func HasTerminalEvent(fpEntityID, eventType string) bool {
	var count int64
	cloudsql.DB.Model(&entity.MfEvent{}).
		Where("fp_entity_id = ? AND event_type = ?", fpEntityID, eventType).
		Count(&count)
	return count > 0
}
