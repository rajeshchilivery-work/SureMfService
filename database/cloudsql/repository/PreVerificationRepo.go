package repository

import (
	"SureMFService/database/cloudsql"
	"SureMFService/database/cloudsql/entity"
)

func CreatePreVerification(pv *entity.PreVerificationUsage) error {
	return cloudsql.DB.Create(pv).Error
}

func GetPreVerificationByFpID(fpID string) (*entity.PreVerificationUsage, error) {
	var pv entity.PreVerificationUsage
	err := cloudsql.DB.Where("fp_pre_verification_id = ?", fpID).First(&pv).Error
	return &pv, err
}

func UpdatePreVerification(pv *entity.PreVerificationUsage) error {
	return cloudsql.DB.Save(pv).Error
}

func GetLatestPreVerificationByUUIDAndType(uuid, verificationType string) (*entity.PreVerificationUsage, error) {
	var pv entity.PreVerificationUsage
	err := cloudsql.DB.
		Where("uuid = ? AND verification_type = ?", uuid, verificationType).
		Order("created_at DESC").
		First(&pv).Error
	return &pv, err
}
