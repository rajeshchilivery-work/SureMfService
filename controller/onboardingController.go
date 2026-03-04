package controller

import (
	"SureCommonService/utils"
	"SureMFService/service"
	"SureMFService/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getUID(c *gin.Context) string {
	uid, _ := c.Get("uid")
	return uid.(string)
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
	pv, err := service.KYCCheck(uid, req.PAN)
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
	var req structs.PhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	fpData, err := service.GetUserFPData(uid)
	if err != nil || fpData.FpInvestorID == "" {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "investor profile not found"})
		return
	}
	phoneID, err := service.AddPhone(uid, fpData.FpInvestorID, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{"fp_phone_id": phoneID}, nil, "MF")
}

func AddEmail(c *gin.Context) {
	uid := getUID(c)
	var req structs.EmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	fpData, err := service.GetUserFPData(uid)
	if err != nil || fpData.FpInvestorID == "" {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "investor profile not found"})
		return
	}
	emailID, err := service.AddEmail(uid, fpData.FpInvestorID, req)
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
	bankID, err := service.AddBankAccount(uid, fpData.FpInvestorID, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{"fp_bank_account_id": bankID}, nil, "MF")
}

func VerifyBankAccount(c *gin.Context) {
	uid := getUID(c)
	var req structs.BankVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	pv, err := service.VerifyBankAccount(uid, req.PAN, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{
		"fp_pre_verification_id": pv.FpPreVerificationID,
		"status":                 pv.Status,
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

func ActivateAccount(c *gin.Context) {
	uid := getUID(c)
	var req structs.ActivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": err.Error()})
		return
	}
	if !req.AgreedTnC {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "terms and conditions must be accepted"})
		return
	}
	fpData, err := service.GetUserFPData(uid)
	if err != nil || fpData.FpInvestorID == "" || fpData.FpBankAccountID == "" {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "complete profile and bank setup before activating"})
		return
	}
	accountID, err := service.ActivateAccount(uid, fpData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, gin.H{"fp_investment_account_id": accountID, "is_activated": true}, nil, "MF")
}
