package service

import (
	"SureMFService/config"
	"SureMFService/structs"
	"fmt"
	"strconv"
)

func CreateMandate(uid string, fpData *structs.UserFPData, req structs.CreateMandateRequest) (*structs.FPMandateResponse, error) {
	if fpData.FpInvestmentAccountID == "" {
		return nil, fmt.Errorf("user has no active investment account")
	}
	if fpData.FpBankAccountID == "" {
		return nil, fmt.Errorf("user has no bank account")
	}

	bankAccount, err := FPGetBankAccount(fpData.FpBankAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bank account: %w", err)
	}

	mandateType := req.MandateType
	if mandateType == "" {
		mandateType = "E_MANDATE"
	}

	fpResp, err := FPCreateMandate(structs.FPCreateMandateRequest{
		BankAccountID: bankAccount.OldID,
		MandateType:   mandateType,
		MandateLimit:  req.MandateLimit,
		ProviderName:  "CYBRILLAPOA",
	})
	if err != nil {
		return nil, err
	}

	logMfEvent(uid, "mandate_created", strconv.Itoa(fpResp.ID), "", req.MandateLimit, 0, map[string]interface{}{
		"mandate_type": mandateType,
	})
	return fpResp, nil
}

func AuthorizeMandate(uid string, mandateID int) (*structs.FPMandateAuthResponse, error) {
	postbackURL := fmt.Sprintf("%s?mandate_id=%d&uid=%s", config.AppConfig.MandatePostbackURL, mandateID, uid)
	fpResp, err := FPAuthorizeMandate(structs.FPMandateAuthRequest{
		MandateID:          mandateID,
		PaymentPostbackURL: postbackURL,
	})
	if err != nil {
		return nil, err
	}

	return fpResp, nil
}

func ListMandates(fpData *structs.UserFPData) (*structs.FPMandateListResponse, error) {
	if fpData.FpBankAccountID == "" {
		return &structs.FPMandateListResponse{Mandates: []structs.FPMandateResponse{}}, nil
	}
	bankAccount, err := FPGetBankAccount(fpData.FpBankAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bank account: %w", err)
	}
	return FPListMandates(bankAccount.OldID)
}

func GetMandateStatus(uid, mandateID string) (*structs.FPMandateResponse, error) {
	fpResp, err := FPGetMandate(mandateID)
	if err != nil {
		return nil, err
	}
	switch fpResp.Status {
	case "approved":
		logTerminalEvent(uid, "mandate_approved", mandateID, "", fpResp.MandateLimit, 0, nil)
	case "failed", "rejected":
		logTerminalEvent(uid, "mandate_failed", mandateID, "", fpResp.MandateLimit, 0, map[string]interface{}{"status": fpResp.Status})
	}
	return fpResp, nil
}

func CancelMandate(uid, mandateID string) error {
	if err := FPCancelMandate(mandateID); err != nil {
		return err
	}
	logMfEvent(uid, "mandate_cancelled", mandateID, "", 0, 0, nil)
	return nil
}
