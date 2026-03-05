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

func FPConfirmOTP(orderType, orderID, otp string) error {
	// orderType determines endpoint: mf_purchases, mf_purchase_plans, mf_redemptions
	endpointMap := map[string]string{
		"purchase":   "mf_purchases",
		"sip":        "mf_purchase_plans",
		"redemption": "mf_redemptions",
	}
	endpoint, ok := endpointMap[orderType]
	if !ok {
		return fmt.Errorf("unknown order type: %s", orderType)
	}

	data := map[string]string{"otp": otp}
	path := fmt.Sprintf("/v2/%s/%s/otp", endpoint, orderID)
	b, status, err := fpRequest(http.MethodPost, path, data)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("fp confirm otp error %d: %s", status, string(b))
	}
	return nil
}

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
