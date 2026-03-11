package service

import (
	"SureMFService/config"
	"SureMFService/database/cloudsql/entity"
	"SureMFService/database/cloudsql/repository"
	"SureMFService/structs"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
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

	// Enrich orders with scheme_name by looking up ISINs
	schemeNames := map[string]string{}
	for _, raw := range orders {
		var m map[string]interface{}
		if json.Unmarshal(raw, &m) == nil {
			if isin, ok := m["scheme"].(string); ok && isin != "" {
				schemeNames[isin] = ""
			}
		}
	}
	for isin := range schemeNames {
		if scheme, err := FPGetFundScheme(isin); err == nil {
			schemeNames[isin] = scheme.Name
		}
	}
	enriched := make([]json.RawMessage, 0, len(orders))
	for _, raw := range orders {
		var m map[string]interface{}
		if json.Unmarshal(raw, &m) == nil {
			if isin, ok := m["scheme"].(string); ok && schemeNames[isin] != "" {
				m["scheme_name"] = schemeNames[isin]
			}
			if b, err := json.Marshal(m); err == nil {
				enriched = append(enriched, b)
				continue
			}
		}
		enriched = append(enriched, raw)
	}
	return enriched, nil
}

func UpdatePurchaseConsent(uid, purchaseID string, fpData *structs.UserFPData) (*structs.FPOrderResponse, error) {
	consent, err := getConsentData(fpData)
	if err != nil {
		return nil, fmt.Errorf("failed to get consent data: %w", err)
	}

	// Retry with delay — FP may take a moment to transition order from under_review to pending
	var fpResp *structs.FPOrderResponse
	for attempt := 0; attempt < 3; attempt++ {
		fpResp, err = FPUpdatePurchaseConsent(structs.FPConsentUpdateRequest{
			ID:      purchaseID,
			Consent: *consent,
		})
		if err == nil {
			return fpResp, nil
		}
		if strings.Contains(err.Error(), "not in pending state") && attempt < 2 {
			log.Printf("[INFO] Order %s not yet in pending state, retrying in 2s (attempt %d)", purchaseID, attempt+1)
			time.Sleep(2 * time.Second)
			continue
		}
		return nil, err
	}
	return fpResp, err
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
		Method:             strings.ToUpper(method),
		PaymentPostbackURL: config.AppConfig.PaymentPostbackURL + "?order_id=" + purchaseID + "&uid=" + uid,
		BankAccountID:      bankAccount.OldID,
		ProviderName:       "ONDC",
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
		// Handle "already confirmed" gracefully — FP may reject if order is already in confirmed state
		existingOrder, getErr := FPGetPurchaseOrder(purchaseID)
		if getErr == nil && existingOrder.State == "confirmed" {
			log.Printf("[INFO] Order %s already confirmed, proceeding", purchaseID)
			logMfEvent(uid, "purchase_confirmed", purchaseID, existingOrder.Scheme, existingOrder.Amount, 0,
				map[string]interface{}{"state": existingOrder.State, "note": "already_confirmed"})
			return existingOrder, nil
		}
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

	// Retry with delay — FP may take a moment to transition SIP plan to review_completed state
	var fpResp *structs.FPSIPDetailResponse
	for attempt := 0; attempt < 3; attempt++ {
		fpResp, err = FPConfirmSIP(structs.FPSIPConfirmRequest{
			ID:      sipID,
			State:   "confirmed",
			Consent: *consent,
		})
		if err == nil {
			logMfEvent(uid, "sip_confirmed", sipID, fpResp.Scheme, fpResp.Amount, 0, map[string]interface{}{"state": fpResp.State})
			return fpResp, nil
		}
		if strings.Contains(err.Error(), "review_completed") && attempt < 2 {
			log.Printf("[INFO] SIP %s not yet in review_completed state, retrying in 2s (attempt %d)", sipID, attempt+1)
			time.Sleep(2 * time.Second)
			continue
		}
		return nil, err
	}
	return fpResp, err
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
	sips, err := FPListSIPs(fpData.FpInvestmentAccountID)
	if err != nil {
		return nil, err
	}

	// Enrich with scheme_name
	schemeNames := map[string]string{}
	for _, s := range sips {
		if s.Scheme != "" {
			schemeNames[s.Scheme] = ""
		}
	}
	for isin := range schemeNames {
		if scheme, err := FPGetFundScheme(isin); err == nil {
			schemeNames[isin] = scheme.Name
		}
	}
	for i := range sips {
		if name := schemeNames[sips[i].Scheme]; name != "" {
			sips[i].SchemeName = name
		}
	}
	return sips, nil
}

// ---- Redemption Lifecycle ----

func ConfirmRedemption(uid, redemptionID string, fpData *structs.UserFPData) (*structs.FPRedemptionDetailResponse, error) {
	consent, err := getConsentData(fpData)
	if err != nil {
		return nil, fmt.Errorf("failed to get consent data: %w", err)
	}

	// Retry with delay — FP may take a moment to transition order from under_review to pending
	var fpResp *structs.FPRedemptionDetailResponse
	for attempt := 0; attempt < 3; attempt++ {
		fpResp, err = FPConfirmRedemption(structs.FPRedemptionConfirmRequest{
			ID:      redemptionID,
			State:   "confirmed",
			Consent: *consent,
		})
		if err == nil {
			logMfEvent(uid, "redemption_confirmed", redemptionID, fpResp.Scheme, fpResp.Amount, fpResp.Units, map[string]interface{}{"state": fpResp.State})
			return fpResp, nil
		}
		if strings.Contains(err.Error(), "not in pending state") && attempt < 2 {
			log.Printf("[INFO] Redemption %s not yet in pending state, retrying in 2s (attempt %d)", redemptionID, attempt+1)
			time.Sleep(2 * time.Second)
			continue
		}
		return nil, err
	}
	return fpResp, err
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
	redemptions, err := FPListRedemptions(fpData.FpInvestmentAccountID)
	if err != nil {
		return nil, err
	}

	// Enrich with scheme_name
	schemeNames := map[string]string{}
	for _, r := range redemptions {
		if r.Scheme != "" {
			schemeNames[r.Scheme] = ""
		}
	}
	for isin := range schemeNames {
		if scheme, err := FPGetFundScheme(isin); err == nil {
			schemeNames[isin] = scheme.Name
		}
	}
	for i := range redemptions {
		if name := schemeNames[redemptions[i].Scheme]; name != "" {
			redemptions[i].SchemeName = name
		}
	}
	return redemptions, nil
}

// ---- Purchase Orders (lumpsum only) ----

func ListPurchases(fpData *structs.UserFPData) ([]structs.FPOrderResponse, error) {
	if fpData.FpInvestmentAccountID == "" {
		return []structs.FPOrderResponse{}, nil
	}
	purchases, err := FPListPurchases(fpData.FpInvestmentAccountID)
	if err != nil {
		return nil, err
	}

	// Enrich with scheme_name
	schemeNames := map[string]string{}
	for _, p := range purchases {
		if p.Scheme != "" {
			schemeNames[p.Scheme] = ""
		}
	}
	for isin := range schemeNames {
		if scheme, err := FPGetFundScheme(isin); err == nil {
			schemeNames[isin] = scheme.Name
		}
	}
	for i := range purchases {
		if name := schemeNames[purchases[i].Scheme]; name != "" {
			purchases[i].SchemeName = name
		}
	}
	return purchases, nil
}

// ---- Portfolio / Folios ----

func GetPortfolio(fpData *structs.UserFPData) (*structs.FPFolioListResponse, error) {
	if fpData.FpInvestmentAccountID == "" {
		return &structs.FPFolioListResponse{Data: []structs.FPFolio{}}, nil
	}
	resp, err := FPGetFolios(fpData.FpInvestmentAccountID)
	if err != nil {
		return nil, err
	}

	// Enrich folios with scheme_name and fund_category from payout_details ISINs
	schemeCache := map[string]*structs.FPFundScheme{}
	for _, folio := range resp.Data {
		for _, pd := range folio.PayoutDetails {
			if pd.Scheme != "" {
				schemeCache[pd.Scheme] = nil
			}
		}
	}
	for isin := range schemeCache {
		if scheme, err := FPGetFundScheme(isin); err == nil {
			schemeCache[isin] = scheme
		}
	}
	for i := range resp.Data {
		if len(resp.Data[i].PayoutDetails) > 0 {
			isin := resp.Data[i].PayoutDetails[0].Scheme
			if scheme := schemeCache[isin]; scheme != nil {
				resp.Data[i].SchemeName = scheme.Name
				resp.Data[i].FundCategory = scheme.FundCategory
			}
		}
	}
	return resp, nil
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

