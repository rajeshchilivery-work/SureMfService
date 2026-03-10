package controller

import (
	"SureCommonService/utils"
	"SureMFService/service"

	"github.com/gin-gonic/gin"
)

func GetEMIROIDelta(c *gin.Context) {
	uid := getUID(c)
	result, err := service.GetEMIROIDelta(uid)
	utils.HandleResponse(c, result, err, "MF")
}
