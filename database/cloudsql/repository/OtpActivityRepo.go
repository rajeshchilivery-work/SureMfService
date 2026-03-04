package repository

import (
	"SureMFService/database/cloudsql"
	"SureMFService/database/cloudsql/entity"
)

func CreateOtpActivity(otp *entity.OtpActivity) error {
	return cloudsql.DB.Create(otp).Error
}

func GetOtpActivityByFpOrderID(fpOrderID string) (*entity.OtpActivity, error) {
	var otp entity.OtpActivity
	err := cloudsql.DB.Where("fp_order_id = ?", fpOrderID).Order("initiated_at DESC").First(&otp).Error
	return &otp, err
}

func UpdateOtpActivity(otp *entity.OtpActivity) error {
	return cloudsql.DB.Save(otp).Error
}
