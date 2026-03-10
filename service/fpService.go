package service

import (
	"SureMFService/config"
	"SureMFService/structs"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	fpTokenMu     sync.Mutex
	fpAccessToken string
	fpTokenExpiry time.Time
)

func getFPToken() (string, error) {
	fpTokenMu.Lock()
	defer fpTokenMu.Unlock()

	if fpAccessToken != "" && time.Now().Before(fpTokenExpiry) {
		return fpAccessToken, nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", config.AppConfig.FPClientID)
	data.Set("client_secret", config.AppConfig.FPClientSecret)

	resp, err := http.PostForm(config.AppConfig.FPBaseURL+"/v2/auth/"+config.AppConfig.FPTenantID+"/token", data)
	if err != nil {
		return "", fmt.Errorf("fp auth request failed: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp structs.FPTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("fp auth decode failed: %w", err)
	}

	fpAccessToken = tokenResp.AccessToken
	fpTokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)
	return fpAccessToken, nil
}

func fpRequest(method, path string, body interface{}) ([]byte, int, error) {
	token, err := getFPToken()
	if err != nil {
		return nil, 0, err
	}

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal error: %w", err)
		}
		reqBody = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, config.AppConfig.FPBaseURL+path, reqBody)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("fp request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBytes, resp.StatusCode, nil
}

// ---- Investor Profile ----

func FPCreateInvestorProfile(req structs.FPInvestorProfileRequest) (*structs.FPInvestorProfileResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/v2/investor_profiles", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp create investor profile error %d: %s", status, string(b))
	}
	var resp structs.FPInvestorProfileResponse
	return &resp, json.Unmarshal(b, &resp)
}

// ---- Phone ----

func FPGetPhone(phoneID string) (*structs.FPPhoneResponse, error) {
	b, status, err := fpRequest(http.MethodGet, "/v2/phone_numbers/"+phoneID, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp get phone error %d: %s", status, string(b))
	}
	var resp structs.FPPhoneResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPGetEmail(emailID string) (*structs.FPEmailResponse, error) {
	b, status, err := fpRequest(http.MethodGet, "/v2/email_addresses/"+emailID, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp get email error %d: %s", status, string(b))
	}
	var resp structs.FPEmailResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPGetMFInvestmentAccount(accountID string) (*structs.FPMFInvestmentAccountResponse, error) {
	b, status, err := fpRequest(http.MethodGet, "/v2/mf_investment_accounts/"+accountID, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp get investment account error %d: %s", status, string(b))
	}
	var resp structs.FPMFInvestmentAccountResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPAddPhone(req structs.FPPhoneRequest) (*structs.FPPhoneResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/v2/phone_numbers", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp add phone error %d: %s", status, string(b))
	}
	var resp structs.FPPhoneResponse
	return &resp, json.Unmarshal(b, &resp)
}

// ---- Email ----

func FPAddEmail(req structs.FPEmailRequest) (*structs.FPEmailResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/v2/email_addresses", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp add email error %d: %s", status, string(b))
	}
	var resp structs.FPEmailResponse
	return &resp, json.Unmarshal(b, &resp)
}

// ---- Address ----

func FPAddAddress(req structs.FPAddressRequest) (*structs.FPAddressResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/v2/addresses", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp add address error %d: %s", status, string(b))
	}
	var resp structs.FPAddressResponse
	return &resp, json.Unmarshal(b, &resp)
}

// ---- Bank Account ----

func FPAddBankAccount(req structs.FPBankAccountRequest) (*structs.FPBankAccountResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/v2/bank_accounts", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp add bank error %d: %s", status, string(b))
	}
	var resp structs.FPBankAccountResponse
	return &resp, json.Unmarshal(b, &resp)
}

// ---- Nominee ----

func FPAddNominee(req structs.FPNomineeRequest) (*structs.FPNomineeResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/v2/related_parties", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp add nominee error %d: %s", status, string(b))
	}
	var resp structs.FPNomineeResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPPatchNominee(nomineeID string, data map[string]interface{}) error {
	b, status, err := fpRequest(http.MethodPatch, "/v2/related_parties/"+nomineeID, data)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("fp patch nominee error %d: %s", status, string(b))
	}
	return nil
}

// ---- MF Investment Account ----

func FPCreateMFInvestmentAccount(req structs.FPMFInvestmentAccountRequest) (*structs.FPMFInvestmentAccountResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/v2/mf_investment_accounts", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp create investment account error %d: %s", status, string(b))
	}
	var resp structs.FPMFInvestmentAccountResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPPatchMFInvestmentAccount(accountID string, req structs.FPMFInvestmentAccountPatchRequest) error {
	b, status, err := fpRequest(http.MethodPatch, "/v2/mf_investment_accounts", req)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("fp patch investment account error %d: %s", status, string(b))
	}
	return nil
}

// ---- Fund Schemes (OMS API — same Bearer token as all other FP calls) ----

func FPListFundSchemes(params map[string]string) (*structs.FPFundSchemeListResponse, error) {
	path := "/api/oms/fund_schemes"
	vals := url.Values{}
	for k, v := range params {
		if v != "" {
			vals.Set(k, v)
		}
	}
	if encoded := vals.Encode(); encoded != "" {
		path += "?" + encoded
	}
	b, status, err := fpRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp list fund_schemes error %d: %s", status, string(b))
	}
	var resp structs.FPFundSchemeListResponse
	return &resp, json.Unmarshal(b, &resp)
}

// ---- Orders ----

func FPCreatePurchaseOrder(req structs.FPPurchaseOrderRequest) (*structs.FPOrderResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/v2/mf_purchases", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp purchase order error %d: %s", status, string(b))
	}
	var resp structs.FPOrderResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPCreateSIPOrder(req structs.FPSIPOrderRequest) (*structs.FPOrderResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/v2/mf_purchase_plans", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp sip order error %d: %s", status, string(b))
	}
	var resp structs.FPOrderResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPCreateRedemptionOrder(req structs.FPRedemptionOrderRequest) (*structs.FPOrderResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/v2/mf_redemptions", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp redemption order error %d: %s", status, string(b))
	}
	var resp structs.FPOrderResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPUpdatePurchaseConsent(req structs.FPConsentUpdateRequest) (*structs.FPOrderResponse, error) {
	b, status, err := fpRequest(http.MethodPatch, "/v2/mf_purchases", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp update consent error %d: %s", status, string(b))
	}
	var resp structs.FPOrderResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPConfirmPurchaseState(req structs.FPConfirmStateRequest) (*structs.FPOrderResponse, error) {
	b, status, err := fpRequest(http.MethodPatch, "/v2/mf_purchases", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp confirm state error %d: %s", status, string(b))
	}
	var resp structs.FPOrderResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPCreatePayment(req structs.FPCreatePaymentRequest) (*structs.FPCreatePaymentResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/api/pg/payments/netbanking", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp create payment error %d: %s", status, string(b))
	}
	var resp structs.FPCreatePaymentResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPGetPurchaseOrder(purchaseID string) (*structs.FPOrderResponse, error) {
	b, status, err := fpRequest(http.MethodGet, "/v2/mf_purchases/"+purchaseID, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp get purchase error %d: %s", status, string(b))
	}
	var resp structs.FPOrderResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPGetBankAccount(bankAccountID string) (*structs.FPBankAccountResponse, error) {
	b, status, err := fpRequest(http.MethodGet, "/v2/bank_accounts/"+bankAccountID, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp get bank account error %d: %s", status, string(b))
	}
	var resp structs.FPBankAccountResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPGetHoldings(investmentAccountOldID int) ([]byte, error) {
	path := fmt.Sprintf("/api/oms/reports/holdings?investment_account_id=%d", investmentAccountOldID)
	b, status, err := fpRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp get holdings error %d: %s", status, string(b))
	}
	return b, nil
}

// ---- SIP Lifecycle ----

func FPConfirmSIP(req structs.FPSIPConfirmRequest) (*structs.FPSIPDetailResponse, error) {
	b, status, err := fpRequest(http.MethodPatch, "/v2/mf_purchase_plans", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp confirm sip error %d: %s", status, string(b))
	}
	var resp structs.FPSIPDetailResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPGetSIPDetail(sipID string) (*structs.FPSIPDetailResponse, error) {
	b, status, err := fpRequest(http.MethodGet, "/v2/mf_purchase_plans/"+sipID, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp get sip error %d: %s", status, string(b))
	}
	var resp structs.FPSIPDetailResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPListSIPs(investmentAccountID string) ([]structs.FPSIPDetailResponse, error) {
	path := "/v2/mf_purchase_plans?mf_investment_account=" + investmentAccountID
	b, status, err := fpRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp list sips error %d: %s", status, string(b))
	}
	var resp struct {
		Data []structs.FPSIPDetailResponse `json:"data"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func FPCancelSIP(sipID, cancellationCode string) (*structs.FPSIPDetailResponse, error) {
	body := map[string]string{
		"id":                sipID,
		"cancellation_code": cancellationCode,
	}
	b, status, err := fpRequest(http.MethodPost, "/v2/mf_purchase_plans/cancel", body)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp cancel sip error %d: %s", status, string(b))
	}
	var resp structs.FPSIPDetailResponse
	return &resp, json.Unmarshal(b, &resp)
}

// ---- Redemption Lifecycle ----

func FPConfirmRedemption(req structs.FPRedemptionConfirmRequest) (*structs.FPRedemptionDetailResponse, error) {
	b, status, err := fpRequest(http.MethodPatch, "/v2/mf_redemptions", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp confirm redemption error %d: %s", status, string(b))
	}
	var resp structs.FPRedemptionDetailResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPGetRedemption(redemptionID string) (*structs.FPRedemptionDetailResponse, error) {
	b, status, err := fpRequest(http.MethodGet, "/v2/mf_redemptions/"+redemptionID, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp get redemption error %d: %s", status, string(b))
	}
	var resp structs.FPRedemptionDetailResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPListRedemptions(investmentAccountID string) ([]structs.FPRedemptionDetailResponse, error) {
	path := "/v2/mf_redemptions?mf_investment_account=" + investmentAccountID
	b, status, err := fpRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp list redemptions error %d: %s", status, string(b))
	}
	var resp struct {
		Data []structs.FPRedemptionDetailResponse `json:"data"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// ---- Portfolio / Folios ----

func FPGetFolios(investmentAccountID string) (*structs.FPFolioListResponse, error) {
	path := "/v2/mf_folios?mf_investment_account=" + investmentAccountID
	b, status, err := fpRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp get folios error %d: %s", status, string(b))
	}
	var resp structs.FPFolioListResponse
	return &resp, json.Unmarshal(b, &resp)
}

// ---- Reports ----

func FPGetSchemeWiseReturns(mfiaID string) ([]byte, error) {
	body := structs.FPTransactionReportRequest{MFInvestmentAccount: mfiaID}
	b, status, err := fpRequest(http.MethodPost, "/v2/transactions/reports/scheme_wise_returns", body)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp scheme wise returns error %d: %s", status, string(b))
	}
	return b, nil
}

func FPGetInvestmentAccountReturns(mfiaID string) ([]byte, error) {
	body := structs.FPTransactionReportRequest{MFInvestmentAccount: mfiaID}
	b, status, err := fpRequest(http.MethodPost, "/v2/transactions/reports/investment_account_wise_returns", body)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp account wise returns error %d: %s", status, string(b))
	}
	return b, nil
}

// ---- Mandates ----

func FPCreateMandate(req structs.FPCreateMandateRequest) (*structs.FPMandateResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/api/pg/mandates", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp create mandate error %d: %s", status, string(b))
	}
	var resp structs.FPMandateResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPAuthorizeMandate(req structs.FPMandateAuthRequest) (*structs.FPMandateAuthResponse, error) {
	b, status, err := fpRequest(http.MethodPost, "/api/pg/payments/emandate/auth", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp authorize mandate error %d: %s", status, string(b))
	}
	var resp structs.FPMandateAuthResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPListMandates(bankAccountOldID int) (*structs.FPMandateListResponse, error) {
	path := fmt.Sprintf("/api/pg/mandates?bank_account_id=%d", bankAccountOldID)
	b, status, err := fpRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp list mandates error %d: %s", status, string(b))
	}
	var resp structs.FPMandateListResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPGetMandate(mandateID string) (*structs.FPMandateResponse, error) {
	b, status, err := fpRequest(http.MethodGet, "/api/pg/mandates/"+mandateID, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("fp get mandate error %d: %s", status, string(b))
	}
	var resp structs.FPMandateResponse
	return &resp, json.Unmarshal(b, &resp)
}

func FPCancelMandate(mandateID string) error {
	b, status, err := fpRequest(http.MethodPost, "/api/pg/mandates/"+mandateID+"/cancel", nil)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("fp cancel mandate error %d: %s", status, string(b))
	}
	return nil
}

// ---- All Orders (combined list) ----

func FPListOrders(investmentAccountID string) ([]byte, error) {
	params := url.Values{}
	params.Set("mf_investment_account", investmentAccountID)

	var allOrders []json.RawMessage

	for _, endpoint := range []string{"mf_purchases", "mf_purchase_plans", "mf_redemptions"} {
		b, status, err := fpRequest(http.MethodGet, "/v2/"+endpoint+"?"+params.Encode(), nil)
		if err != nil {
			continue
		}
		if status >= 400 {
			continue
		}
		var resp struct {
			Data []json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(b, &resp); err == nil {
			allOrders = append(allOrders, resp.Data...)
		}
	}

	result, _ := json.Marshal(allOrders)
	return result, nil
}
