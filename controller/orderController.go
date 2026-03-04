package controller

import (
	"SureCommonService/utils"
	"SureMFService/service"
	"SureMFService/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PlacePurchaseOrder(c *gin.Context) {
	uid := getUID(c)
	var req structs.PurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	order, err := service.PlacePurchaseOrder(uid, fpData, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, order, nil, "MF")
}

func PlaceSIPOrder(c *gin.Context) {
	uid := getUID(c)
	var req structs.SIPOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	order, err := service.PlaceSIPOrder(uid, fpData, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, order, nil, "MF")
}

func PlaceRedemptionOrder(c *gin.Context) {
	uid := getUID(c)
	var req structs.RedemptionOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	order, err := service.PlaceRedemptionOrder(uid, fpData, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, order, nil, "MF")
}

func ConfirmOTP(c *gin.Context) {
	uid := getUID(c)
	orderID := c.Param("id")
	orderType := c.Query("type") // ?type=purchase|sip|redemption
	if orderType == "" {
		orderType = "purchase"
	}
	var req structs.ConfirmOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	if err := service.ConfirmOrderOTP(uid, orderID, orderType, req.OTP); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{"message": "OTP confirmed successfully"}, nil, "MF")
}

func GetOrders(c *gin.Context) {
	uid := getUID(c)
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	orders, err := service.GetUserOrders(fpData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, orders, nil, "MF")
}
