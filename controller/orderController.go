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
	userIP := getUserIP(c)
	order, err := service.PlacePurchaseOrder(uid, fpData, req, userIP)
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
	userIP := getUserIP(c)
	order, err := service.PlaceSIPOrder(uid, fpData, req, userIP)
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
	userIP := getUserIP(c)
	order, err := service.PlaceRedemptionOrder(uid, fpData, req, userIP)
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

func UpdateConsent(c *gin.Context) {
	uid := getUID(c)
	orderID := c.Param("id")
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	resp, err := service.UpdatePurchaseConsent(uid, orderID, fpData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func CreatePayment(c *gin.Context) {
	uid := getUID(c)
	orderID := c.Param("id")
	var req structs.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	resp, err := service.CreatePayment(uid, fpData, orderID, req.Method)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func ConfirmPurchaseState(c *gin.Context) {
	uid := getUID(c)
	orderID := c.Param("id")
	resp, err := service.ConfirmPurchaseState(uid, orderID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func GetPurchaseStatus(c *gin.Context) {
	uid := getUID(c)
	orderID := c.Param("id")
	resp, err := service.GetPurchaseStatus(uid, orderID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func GetHoldings(c *gin.Context) {
	folio := c.Query("folio")
	if folio == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "folio query parameter is required"})
		return
	}
	holdings, err := service.GetHoldings(folio)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, holdings.Holdings, nil, "MF")
}

// ---- SIP Lifecycle ----

func ConfirmSIP(c *gin.Context) {
	uid := getUID(c)
	sipID := c.Param("id")
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	resp, err := service.ConfirmSIP(uid, sipID, fpData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func GetSIPDetail(c *gin.Context) {
	uid := getUID(c)
	sipID := c.Param("id")
	resp, err := service.GetSIPDetail(uid, sipID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func ListSIPs(c *gin.Context) {
	uid := getUID(c)
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	resp, err := service.ListSIPs(fpData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func GetSIPInstallments(c *gin.Context) {
	sipID := c.Param("id")
	resp, err := service.GetSIPInstallments(sipID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func CancelSIP(c *gin.Context) {
	uid := getUID(c)
	sipID := c.Param("id")
	var req structs.CancelSIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	resp, err := service.CancelSIP(uid, sipID, req.CancellationCode)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

// ---- Redemption Lifecycle ----

func ConfirmRedemption(c *gin.Context) {
	uid := getUID(c)
	redemptionID := c.Param("id")
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	resp, err := service.ConfirmRedemption(uid, redemptionID, fpData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func GetRedemptionDetail(c *gin.Context) {
	uid := getUID(c)
	redemptionID := c.Param("id")
	resp, err := service.GetRedemptionDetail(uid, redemptionID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func ListRedemptions(c *gin.Context) {
	uid := getUID(c)
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	resp, err := service.ListRedemptions(fpData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

// ---- Portfolio / Folios ----

func GetPortfolio(c *gin.Context) {
	uid := getUID(c)
	fpData, err := service.GetUserFPData(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	resp, err := service.GetPortfolio(fpData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}

func GetFolioDetail(c *gin.Context) {
	folioID := c.Param("id")
	resp, err := service.GetFolioDetail(folioID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, resp, nil, "MF")
}
