package repository

import (
	"SureMFService/database/cloudsql"
	"SureMFService/database/cloudsql/entity"
	"fmt"
)

func GetSureUserByUID(uid string) (*entity.SureUser, error) {
	var user entity.SureUser
	err := cloudsql.DB.Where("uuid = ?", uid).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("user not found in sure_user.users for uid %s: %w", uid, err)
	}
	return &user, nil
}
