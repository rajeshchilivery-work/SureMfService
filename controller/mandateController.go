package controller

import (
	"SureCommonService/utils"
	"SureMFService/service"
	"SureMFService/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateMandate(c *gin.Context) {
	uid := getUID(c)
	var req structs.CreateMandateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	resp, err := service.CreateMandate(uid, fpData, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func AuthorizeMandate(c *gin.Context) {
	uid := getUID(c)
	var req structs.AuthorizeMandateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	resp, err := service.AuthorizeMandate(uid, req.MandateID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func ListMandates(c *gin.Context) {
	uid := getUID(c)
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	resp, err := service.ListMandates(fpData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func GetMandateStatus(c *gin.Context) {
	uid := getUID(c)
	mandateID := c.Param("id")
	resp, err := service.GetMandateStatus(uid, mandateID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func CancelMandate(c *gin.Context) {
	uid := getUID(c)
	mandateID := c.Param("id")
	if err := service.CancelMandate(uid, mandateID); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{"message": "mandate cancelled"}, nil, "MF")
}
