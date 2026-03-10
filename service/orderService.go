package service

import (
	"SureMFService/config"
	"SureMFService/database/cloudsql/entity"
	"SureMFService/database/cloudsql/repository"
	"SureMFService/structs"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func PlacePurchaseOrder(uid string, fpData *structs.UserFPData, req structs.PurchaseOrderRequest, userIP string) (*structs.FPOrderResponse, error) {
	if fpData.FpInvestmentAccountID == "" {
		return nil, fmt.Errorf("user has no active investment account")
	}

	fpResp, err := FPCreatePurchaseOrder(structs.FPPurchaseOrderRequest{
		MFInvestmentAccount: fpData.FpInvestmentAccountID,
		SchemeID:            req.SchemeID,
		Amount:              req.Amount,
		UserIP:              userIP,
	})
	if err != nil {
		return nil, err
	}

	logMfEvent(uid, "purchase_order_created", fpResp.ID, req.SchemeID, req.Amount, 0, nil)
	return fpResp, nil
}

func PlaceSIPOrder(uid string, fpData *structs.UserFPData, req structs.SIPOrderRequest, userIP string) (*structs.FPOrderResponse, error) {
	if fpData.FpInvestmentAccountID == "" {
		return nil, fmt.Errorf("user has no active investment account")
	}

	fpReq := structs.FPSIPOrderRequest{
		MFInvestmentAccount:      fpData.FpInvestmentAccountID,
		SchemeID:                 req.SchemeID,
		Amount:                   req.Amount,
		Frequency:                req.Frequency,
		InstallmentDay:           req.SIPDate,
		Systematic:               true,
		UserIP:                   userIP,
		AutoGenerateInstallments: true,
		NumberOfInstallments:     req.NumberOfInstallments,
	}
	if req.MandateID != 0 {
		fpReq.PaymentMethod = "mandate"
		fpReq.PaymentSource = strconv.Itoa(req.MandateID)
	}

	fpResp, err := FPCreateSIPOrder(fpReq)
	if err != nil {
		return nil, err
	}

	logMfEvent(uid, "sip_order_created", fpResp.ID, req.SchemeID, req.Amount, 0, map[string]interface{}{
		"frequency":        req.Frequency,
		"installment_day":  req.SIPDate,
		"installments":     req.NumberOfInstallments,
		"mandate_id":       req.MandateID,
	})
	return fpResp, nil
}

func PlaceRedemptionOrder(uid string, fpData *structs.UserFPData, req structs.RedemptionOrderRequest, userIP string) (*structs.FPOrderResponse, error) {
	if fpData.FpInvestmentAccountID == "" {
		return nil, fmt.Errorf("user has no active investment account")
	}

	fpResp, err := FPCreateRedemptionOrder(structs.FPRedemptionOrderRequest{
		MFInvestmentAccount: fpData.FpInvestmentAccountID,
		FolioNumber:         req.FolioNumber,
		SchemeID:            req.SchemeID,
		Amount: req.Amount,
		Units:  req.Units,
		UserIP: userIP,
	})
	if err != nil {
		return nil, err
	}

	logMfEvent(uid, "redemption_order_created", fpResp.ID, req.SchemeID, req.Amount, req.Units, nil)
	return fpResp, nil
}

func GetUserOrders(fpData *structs.UserFPData) ([]json.RawMessage, error) {
	if fpData.FpInvestmentAccountID == "" {
		return []json.RawMessage{}, nil
	}
	b, err := FPListOrders(fpData.FpInvestmentAccountID)
	if err != nil {
		return nil, err
	}
	var orders []json.RawMessage
	if err := json.Unmarshal(b, &orders); err != nil {
		return nil, fmt.Errorf("failed to parse orders: %w", err)
	}
	return orders, nil
}

func UpdatePurchaseConsent(uid, purchaseID string, fpData *structs.UserFPData) (*structs.FPOrderResponse, error) {
	consent, err := getConsentData(fpData)
	if err != nil {
		return nil, fmt.Errorf("failed to get consent data: %w", err)
	}
	fpResp, err := FPUpdatePurchaseConsent(structs.FPConsentUpdateRequest{
		ID:      purchaseID,
		Consent: *consent,
	})
	if err != nil {
		return nil, err
	}

	return fpResp, nil
}

func CreatePayment(uid string, fpData *structs.UserFPData, purchaseID, method string) (*structs.FPCreatePaymentResponse, error) {
	// Fetch purchase order to get old_id
	purchase, err := FPGetPurchaseOrder(purchaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch purchase order: %w", err)
	}

	// Fetch bank account to get old_id
	bankAccount, err := FPGetBankAccount(fpData.FpBankAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bank account: %w", err)
	}

	payResp, err := FPCreatePayment(structs.FPCreatePaymentRequest{
		AMCOrderIDs:        []int{purchase.OldID},
		Method:             method,
		PaymentPostbackURL: config.AppConfig.PaymentPostbackURL,
		BankAccountID:      bankAccount.OldID,
	})
	if err != nil {
		return nil, err
	}

	return payResp, nil
}

func ConfirmPurchaseState(uid, purchaseID string) (*structs.FPOrderResponse, error) {
	fpResp, err := FPConfirmPurchaseState(structs.FPConfirmStateRequest{
		ID:    purchaseID,
		State: "confirmed",
	})
	if err != nil {
		return nil, err
	}

	logMfEvent(uid, "purchase_confirmed", purchaseID, fpResp.Scheme, fpResp.Amount, 0, map[string]interface{}{"state": fpResp.State})
	return fpResp, nil
}

func GetPurchaseStatus(uid, purchaseID string) (*structs.FPOrderResponse, error) {
	fpResp, err := FPGetPurchaseOrder(purchaseID)
	if err != nil {
		return nil, err
	}
	switch fpResp.State {
	case "successful":
		logTerminalEvent(uid, "purchase_successful", purchaseID, fpResp.Scheme, fpResp.Amount, 0, nil)
	case "failed":
		logTerminalEvent(uid, "purchase_failed", purchaseID, fpResp.Scheme, fpResp.Amount, 0, nil)
	}
	return fpResp, nil
}

func GetHoldings(fpData *structs.UserFPData) (json.RawMessage, error) {
	if fpData.FpInvestmentAccountID == "" {
		return json.RawMessage("{}"), nil
	}
	account, err := FPGetMFInvestmentAccount(fpData.FpInvestmentAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch investment account: %w", err)
	}
	b, err := FPGetHoldings(account.OldID)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

// ---- SIP Lifecycle ----

func ConfirmSIP(uid, sipID string, fpData *structs.UserFPData) (*structs.FPSIPDetailResponse, error) {
	consent, err := getConsentData(fpData)
	if err != nil {
		return nil, fmt.Errorf("failed to get consent data: %w", err)
	}
	fpResp, err := FPConfirmSIP(structs.FPSIPConfirmRequest{
		ID:    sipID,
		State: "confirmed",
		Consent: *consent,
	})
	if err != nil {
		return nil, err
	}
	logMfEvent(uid, "sip_confirmed", sipID, fpResp.Scheme, fpResp.Amount, 0, map[string]interface{}{"state": fpResp.State})
	return fpResp, nil
}

func GetSIPDetail(uid, sipID string) (*structs.FPSIPDetailResponse, error) {
	fpResp, err := FPGetSIPDetail(sipID)
	if err != nil {
		return nil, err
	}
	switch fpResp.State {
	case "active":
		logTerminalEvent(uid, "sip_active", sipID, fpResp.Scheme, fpResp.Amount, 0, nil)
	case "failed":
		logTerminalEvent(uid, "sip_failed", sipID, fpResp.Scheme, fpResp.Amount, 0, nil)
	}
	return fpResp, nil
}

func CancelSIP(uid, sipID, cancellationCode string) (*structs.FPSIPDetailResponse, error) {
	fpResp, err := FPCancelSIP(sipID, cancellationCode)
	if err != nil {
		return nil, err
	}
	logMfEvent(uid, "sip_cancelled", sipID, fpResp.Scheme, fpResp.Amount, 0, nil)
	return fpResp, nil
}

func ListSIPs(fpData *structs.UserFPData) ([]structs.FPSIPDetailResponse, error) {
	if fpData.FpInvestmentAccountID == "" {
		return []structs.FPSIPDetailResponse{}, nil
	}
	return FPListSIPs(fpData.FpInvestmentAccountID)
}

// ---- Redemption Lifecycle ----

func ConfirmRedemption(uid, redemptionID string, fpData *structs.UserFPData) (*structs.FPRedemptionDetailResponse, error) {
	consent, err := getConsentData(fpData)
	if err != nil {
		return nil, fmt.Errorf("failed to get consent data: %w", err)
	}
	fpResp, err := FPConfirmRedemption(structs.FPRedemptionConfirmRequest{
		ID:    redemptionID,
		State: "confirmed",
		Consent: *consent,
	})
	if err != nil {
		return nil, err
	}
	logMfEvent(uid, "redemption_confirmed", redemptionID, fpResp.Scheme, fpResp.Amount, fpResp.Units, map[string]interface{}{"state": fpResp.State})
	return fpResp, nil
}

func GetRedemptionDetail(uid, redemptionID string) (*structs.FPRedemptionDetailResponse, error) {
	fpResp, err := FPGetRedemption(redemptionID)
	if err != nil {
		return nil, err
	}
	switch fpResp.State {
	case "successful":
		logTerminalEvent(uid, "redemption_successful", redemptionID, fpResp.Scheme, fpResp.Amount, fpResp.Units, nil)
	case "failed":
		logTerminalEvent(uid, "redemption_failed", redemptionID, fpResp.Scheme, fpResp.Amount, fpResp.Units, nil)
	}
	return fpResp, nil
}

func ListRedemptions(fpData *structs.UserFPData) ([]structs.FPRedemptionDetailResponse, error) {
	if fpData.FpInvestmentAccountID == "" {
		return []structs.FPRedemptionDetailResponse{}, nil
	}
	return FPListRedemptions(fpData.FpInvestmentAccountID)
}

// ---- Portfolio / Folios ----

func GetPortfolio(fpData *structs.UserFPData) (*structs.FPFolioListResponse, error) {
	if fpData.FpInvestmentAccountID == "" {
		return &structs.FPFolioListResponse{Data: []structs.FPFolio{}}, nil
	}
	return FPGetFolios(fpData.FpInvestmentAccountID)
}

func GetSchemeWiseReturns(fpData *structs.UserFPData) (json.RawMessage, error) {
	if fpData.FpInvestmentAccountID == "" {
		return json.RawMessage("{}"), nil
	}
	b, err := FPGetSchemeWiseReturns(fpData.FpInvestmentAccountID)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

func GetInvestmentAccountReturns(fpData *structs.UserFPData) (json.RawMessage, error) {
	if fpData.FpInvestmentAccountID == "" {
		return json.RawMessage("{}"), nil
	}
	b, err := FPGetInvestmentAccountReturns(fpData.FpInvestmentAccountID)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

// ---- Internal helpers ----

func getConsentData(fpData *structs.UserFPData) (*structs.FPConsentDetail, error) {
	if fpData.FpPhoneID == "" || fpData.FpEmailID == "" {
		return nil, fmt.Errorf("user missing phone or email ID")
	}
	phone, err := FPGetPhone(fpData.FpPhoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch phone: %w", err)
	}
	email, err := FPGetEmail(fpData.FpEmailID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch email: %w", err)
	}
	return &structs.FPConsentDetail{
		Email:   email.Email,
		ISDCode: "91",
		Mobile:  phone.Number,
	}, nil
}

func logTerminalEvent(uid, eventType, fpEntityID, isin string, amount, units float64, payload map[string]interface{}) {
	if repository.HasTerminalEvent(fpEntityID, eventType) {
		return
	}
	logMfEvent(uid, eventType, fpEntityID, isin, amount, units, payload)
}

func logMfEvent(uid, eventType, fpEntityID, isin string, amount, units float64, payload map[string]interface{}) {
	event := &entity.MfEvent{
		UserID:    uid,
		EventType: eventType,
		EventAt:   time.Now(),
	}
	if fpEntityID != "" {
		event.FpEntityID = strPtr(fpEntityID)
	}
	if isin != "" {
		event.ISIN = strPtr(isin)
	}
	if amount != 0 {
		event.Amount = &amount
	}
	if units != 0 {
		event.Units = &units
	}
	if payload != nil {
		event.RawPayload = entity.JSONB(payload)
	}
	_ = repository.CreateMfEvent(event)
}

