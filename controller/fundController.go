package controller

import (
	"SureCommonService/utils"
	"SureMFService/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetFund(c *gin.Context) {
	isin := c.Param("isin")
	if isin == "" {
		c.JSON(http.StatusOK, gin.H{"status": 400, "msg": "ISIN is required"})
		return
	}
	fund, err := service.GetFundByISIN(isin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, fund, nil, "MF")
}

func ListFunds(c *gin.Context) {
	investmentOption := c.Query("investment_option") // GROWTH, DIV_REINVESTMENT, DIV_PAYOUT
	planType := c.Query("plan_type")                 // Regular, Direct
	amcID := c.Query("amc_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	funds, err := service.ListFunds(investmentOption, planType, amcID, page, size)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": err.Error()})
		return
	}
	utils.HandleResponse(c, funds, nil, "MF")
}
