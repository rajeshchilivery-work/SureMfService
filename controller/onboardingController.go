package controller

import (
	"SureCommonService/utils"
	"SureMFService/service"
	"SureMFService/structs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getUID(c *gin.Context) string {
	uid, _ := c.Get("uid")
	return uid.(string)
}

func getUserIP(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "::1" || ip == "" || ip == "127.0.0.1" {
		return "10.0.128.12"
	}
	return ip
}

func GetPreVerificationStatus(c *gin.Context) {
	fpID := c.Param("fp_id")
	pv, fpPV, err := service.GetPreVerificationStatus(fpID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{
		"fp_pre_verification_id": pv.FpPreVerificationID,
		"verification_type":      pv.VerificationType,
		"status":                 pv.Status,
		"fp_status":              fpPV.Status,
		"pan":                    fpPV.PAN,
		"readiness":              fpPV.Readiness,
		"bank_accounts":          fpPV.BankAccounts,
	}, nil, "MF")
}

func GetOnboardingStatus(c *gin.Context) {
	uid := getUID(c)
	data, err := service.GetOnboardingStatus(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, data, nil, "MF")
}

func KYCCheck(c *gin.Context) {
	uid := getUID(c)
	var req structs.KYCCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	pv, err := service.KYCCheck(uid, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{
		"fp_pre_verification_id": pv.FpPreVerificationID,
		"status":                 pv.Status,
	}, nil, "MF")
}

func CreateInvestorProfile(c *gin.Context) {
	uid := getUID(c)
	var req structs.InvestorProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	investorID, err := service.CreateInvestorProfile(uid, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{"fp_investor_id": investorID}, nil, "MF")
}

func AddPhone(c *gin.Context) {
	uid := getUID(c)
	fpData, err := service.GetUserFPData(uid)
	if err != nil || fpData.FpInvestorID == "" {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "investor profile not found"})
		return
	}
	phoneID, err := service.AddPhone(uid, fpData.FpInvestorID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{"fp_phone_id": phoneID}, nil, "MF")
}

func AddEmail(c *gin.Context) {
	uid := getUID(c)
	fpData, err := service.GetUserFPData(uid)
	if err != nil || fpData.FpInvestorID == "" {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "investor profile not found"})
		return
	}
	emailID, err := service.AddEmail(uid, fpData.FpInvestorID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{"fp_email_id": emailID}, nil, "MF")
}

func AddAddress(c *gin.Context) {
	uid := getUID(c)
	var req structs.AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	fpData, err := service.GetUserFPData(uid)
	if err != nil || fpData.FpInvestorID == "" {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "investor profile not found"})
		return
	}
	addressID, err := service.AddAddress(uid, fpData.FpInvestorID, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{"fp_address_id": addressID}, nil, "MF")
}

func AddBankAccount(c *gin.Context) {
	uid := getUID(c)
	var req structs.BankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	fpData, err := service.GetUserFPData(uid)
	if err != nil || fpData.FpInvestorID == "" {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "investor profile not found"})
		return
	}
	fpPreVerifID, fpBankAccountID, err := service.AddBankAccount(uid, fpData.FpInvestorID, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":                 500,
			"msg":                    err.Error(),
			"fp_pre_verification_id": fpPreVerifID,
		})
		return
	}
	utils.HandleResponse(c, gin.H{
		"fp_bank_account_id":     fpBankAccountID,
		"fp_pre_verification_id": fpPreVerifID,
		"verification_status":    "completed",
	}, nil, "MF")
}

func AddNominee(c *gin.Context) {
	uid := getUID(c)
	var req structs.NomineeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	fpData, err := service.GetUserFPData(uid)
	if err != nil || fpData.FpInvestorID == "" {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "investor profile not found"})
		return
	}
	nomineeID, err := service.AddNominee(uid, fpData.FpInvestorID, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{"fp_nominee_id": nomineeID}, nil, "MF")
}

func PaymentCallback(c *gin.Context) {
	log.Printf("[PAYMENT CALLBACK] method=%s query=%s", c.Request.Method, c.Request.URL.RawQuery)
	if c.Request.Method == http.MethodPost {
		var payload map[string]interface{}
		if err := c.ShouldBindJSON(&payload); err == nil {
			log.Printf("[PAYMENT CALLBACK] body: %+v", payload)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"status":     "received",
		"order_id":   c.Query("order_id"),
		"payment_id": c.Query("payment_id"),
	})
}

func MandateCallback(c *gin.Context) {
	log.Printf("[MANDATE CALLBACK] method=%s query=%s", c.Request.Method, c.Request.URL.RawQuery)
	if c.Request.Method == http.MethodPost {
		var payload map[string]interface{}
		if err := c.ShouldBindJSON(&payload); err == nil {
			log.Printf("[MANDATE CALLBACK] body: %+v", payload)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"status":     "received",
		"mandate_id": c.Query("mandate_id"),
	})
}

func ActivateAccount(c *gin.Context) {
	uid := getUID(c)
	var req structs.ActivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	fpData, err := service.GetUserFPData(uid)
	if err != nil || fpData.FpInvestorID == "" || fpData.FpBankAccountID == "" {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "complete profile and bank setup before activating"})
		return
	}
	accountID, err := service.ActivateAccount(uid, fpData, req.Nominee1IdentityProofType)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{"fp_investment_account_id": accountID, "is_activated": true}, nil, "MF")
}
