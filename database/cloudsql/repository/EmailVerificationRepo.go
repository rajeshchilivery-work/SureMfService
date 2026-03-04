package repository

import (
	"SureMFService/database/cloudsql"
	"SureMFService/database/cloudsql/entity"
)

func CreateEmailVerification(ev *entity.EmailVerification) error {
	return cloudsql.DB.Create(ev).Error
}

func GetLatestPendingEmailVerification(uuid, email string) (*entity.EmailVerification, error) {
	var ev entity.EmailVerification
	err := cloudsql.DB.
		Where("uuid = ? AND email = ? AND status = ?", uuid, email, "pending").
		Order("initiated_at DESC").
		First(&ev).Error
	return &ev, err
}

func UpdateEmailVerification(ev *entity.EmailVerification) error {
	return cloudsql.DB.Save(ev).Error
}
