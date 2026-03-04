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
	poaTokenMu     sync.Mutex
	poaAccessToken string
	poaTokenExpiry time.Time
)

func getPoaToken() (string, error) {
	poaTokenMu.Lock()
	defer poaTokenMu.Unlock()

	if poaAccessToken != "" && time.Now().Before(poaTokenExpiry) {
		return poaAccessToken, nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", config.AppConfig.FPPoaClientID)
	data.Set("client_secret", config.AppConfig.FPPoaClientSecret)

	resp, err := http.PostForm(config.AppConfig.FPPoaAuthURL, data)
	if err != nil {
		return "", fmt.Errorf("poa auth request failed: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp structs.FPTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("poa auth decode failed: %w", err)
	}

	poaAccessToken = tokenResp.AccessToken
	poaTokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)
	return poaAccessToken, nil
}

func poaRequest(method, path string, body interface{}) ([]byte, int, error) {
	token, err := getPoaToken()
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

	req, err := http.NewRequest(method, config.AppConfig.FPPoaBaseURL+path, reqBody)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("poa request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBytes, resp.StatusCode, nil
}

func POACreatePreVerification(req structs.FPPreVerificationRequest) (*structs.FPPreVerification, error) {
	b, status, err := poaRequest(http.MethodPost, "/poa/pre_verifications", req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("poa pre_verification error %d: %s", status, string(b))
	}
	var resp structs.FPPreVerification
	return &resp, json.Unmarshal(b, &resp)
}

func POAGetPreVerification(verificationID string) (*structs.FPPreVerification, error) {
	b, status, err := poaRequest(http.MethodGet, "/poa/pre_verifications/"+verificationID, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("poa get pre_verification error %d: %s", status, string(b))
	}
	var resp structs.FPPreVerification
	return &resp, json.Unmarshal(b, &resp)
}

// PollPreVerification polls until status is "completed" or "failed" (not "accepted"/"pending")
func PollPreVerification(verificationID string, maxAttempts int) (*structs.FPPreVerification, error) {
	for i := 0; i < maxAttempts; i++ {
		pv, err := POAGetPreVerification(verificationID)
		if err != nil {
			return nil, err
		}
		if pv.Status == "completed" || pv.Status == "failed" {
			return pv, nil
		}
		time.Sleep(1 * time.Second)
	}
	// Return last known state even if still in progress
	return POAGetPreVerification(verificationID)
}
