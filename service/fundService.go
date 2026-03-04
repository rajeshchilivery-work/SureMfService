package service

import (
	"SureMFService/structs"
	"fmt"
)

func ListFunds(investmentOption, planType, amcID string, page, size int) (*structs.FPFundSchemeListResponse, error) {
	params := map[string]string{
		"investment_option": investmentOption,
		"plan_type":         planType,
		"amc_id":            amcID,
	}
	if page >= 0 {
		params["page"] = fmt.Sprintf("%d", page)
	}
	if size > 0 {
		params["size"] = fmt.Sprintf("%d", size)
	}
	return FPListFundSchemes(params)
}
