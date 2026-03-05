package service

import (
	"SureMFService/database/cloudsql/entity"
	"SureMFService/database/cloudsql/repository"
	"SureMFService/database/firebase"
	"SureMFService/structs"
	"fmt"
	"time"
)

const userFpCollection = "user_fp_mapping"
const usersCollection = "users"

// GetUserProfile fetches user details from Firestore "users" collection by uid.
func GetUserProfile(uid string) (*structs.UserProfile, error) {
	var profile structs.UserProfile
	found, err := firebase.GetDoc(usersCollection, uid, &profile)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("user profile not found for uid: %s", uid)
	}
	return &profile, nil
}

func GetUserFPData(uid string) (*structs.UserFPData, error) {
	var data structs.UserFPData
	_, err := firebase.GetDoc(userFpCollection, uid, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func saveUserFPFields(uid string, fields map[string]interface{}) error {
	return firebase.SetDocFields(userFpCollection, uid, fields)
}

// epochMsToDate converts epoch milliseconds (can be negative) to "YYYY-MM-DD".
func epochMsToDate(ms int64) string {
	return time.UnixMilli(ms).UTC().Format("2006-01-02")
}

// normalisePreVerifStatus maps FP status to DB-allowed values: pending | completed | failed.
// For KYC checks, readiness.status determines actual compliance even when top-level status is "completed".
func normalisePreVerifStatus(pv *structs.FPPreVerification) string {
	switch pv.Status {
	case "completed":
		if pv.Readiness != nil && pv.Readiness.Status == "failed" {
			return "failed"
		}
		return "completed"
	case "failed":
		return "failed"
	default: // "accepted" or anything else
		return "pending"
	}
}

// deriveBankVerifStatus checks bank_accounts[0].status first, then falls back to readiness.status.
func deriveBankVerifStatus(pv *structs.FPPreVerification) string {
	if len(pv.BankAccounts) > 0 {
		switch pv.BankAccounts[0].Status {
		case "verified":
			return "completed"
		case "failed":
			return "failed"
		}
	}
	// readiness.status == "failed" also means the account is invalid
	if pv.Readiness != nil && pv.Readiness.Status == "failed" {
		return "failed"
	}
	return normalisePreVerifStatus(pv)
}

func KYCCheck(uid string, req structs.KYCCheckRequest) (*entity.PreVerificationUsage, error) {
	user, err := repository.GetSureUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}

	fpPV, err := POACreatePreVerification(structs.FPPreVerificationRequest{
		InvestorIdentifier: req.PAN,
		PAN:                &structs.FPPreVerifField{Value: req.PAN},
		Name:               &structs.FPPreVerifField{Value: user.Name},
		DateOfBirth:        &structs.FPPreVerifField{Value: user.DOB.Format("2006-01-02")},
	})
	if err != nil {
		return nil, fmt.Errorf("kyc pre_verification failed: %w", err)
	}

	// Save initial row with pending status
	fpID := fpPV.ID
	pv := &entity.PreVerificationUsage{
		UUID:                uid,
		VerificationType:    "kyc_verification",
		Pan:                 req.PAN,
		Status:              "pending",
		FpPreVerificationID: &fpID,
		TriggeredBy:         strPtr("kyc_check"),
	}
	if err := repository.CreatePreVerification(pv); err != nil {
		return nil, err
	}

	// Poll FP for final status (max 5 attempts, 1s interval)
	if polled, pollErr := PollPreVerification(fpID, 5); pollErr == nil {
		pv.Status = normalisePreVerifStatus(polled)
		pv.PollCount++
		_ = repository.UpdatePreVerification(pv)
	}

	return pv, nil
}

func CreateInvestorProfile(uid string, req structs.InvestorProfileRequest) (string, error) {
	user, err := repository.GetSureUserByUID(uid)
	if err != nil {
		return "", fmt.Errorf("failed to load user: %w", err)
	}

	fpResp, err := FPCreateInvestorProfile(structs.FPInvestorProfileRequest{
		Type:                    "individual",
		TaxStatus:               "resident_individual",
		Name:                    user.Name,
		DateOfBirth:             user.DOB.Format("2006-01-02"),
		Gender:                  normGender(user.Gender),
		Occupation:              req.Occupation,
		PAN:                     user.PAN,
		PlaceOfBirth:            "IN",
		UseDefaultTaxResidences: "false",
		FirstTaxResidency: structs.FPTaxResidency{
			Country:     "IN",
			TaxIDType:   "pan",
			TaxIDNumber: user.PAN,
		},
		SourceOfWealth: req.SourceOfWealth,
		IncomeSlab:     req.IncomeSlab,
		PEPDetails:     "not_applicable",
	})
	if err != nil {
		return "", err
	}
	if err := saveUserFPFields(uid, map[string]interface{}{
		"fp_investor_id":  fpResp.ID,
		"onboarding_step": 1,
	}); err != nil {
		return "", err
	}
	return fpResp.ID, nil
}

func AddPhone(uid, investorID string) (string, error) {
	user, err := repository.GetSureUserByUID(uid)
	if err != nil {
		return "", fmt.Errorf("failed to load user: %w", err)
	}
	fpResp, err := FPAddPhone(structs.FPPhoneRequest{
		InvestorProfileID: investorID,
		ISD:               "91",
		Number:            user.PhoneNumber,
	})
	if err != nil {
		return "", err
	}
	if err := saveUserFPFields(uid, map[string]interface{}{"fp_phone_id": fpResp.ID}); err != nil {
		return "", err
	}
	return fpResp.ID, nil
}

func AddEmail(uid, investorID string) (string, error) {
	user, err := repository.GetSureUserByUID(uid)
	if err != nil {
		return "", fmt.Errorf("failed to load user: %w", err)
	}
	fpResp, err := FPAddEmail(structs.FPEmailRequest{
		InvestorProfileID: investorID,
		Email:             user.Email,
	})
	if err != nil {
		return "", err
	}
	if err := saveUserFPFields(uid, map[string]interface{}{"fp_email_id": fpResp.ID}); err != nil {
		return "", err
	}
	return fpResp.ID, nil
}

func AddAddress(uid, investorID string, req structs.AddressRequest) (string, error) {
	country := req.Country
	if country == "" {
		country = "IN"
	}
	nature := req.Nature
	if nature == "" {
		nature = "residential"
	}
	fpResp, err := FPAddAddress(structs.FPAddressRequest{
		InvestorProfileID: investorID,
		Line1:             req.Line1,
		Line2:             req.Line2,
		City:              req.City,
		State:             req.State,
		PostalCode:        req.Pincode,
		Country:           country,
		Nature:            nature,
	})
	if err != nil {
		return "", err
	}
	if err := saveUserFPFields(uid, map[string]interface{}{"fp_address_id": fpResp.ID}); err != nil {
		return "", err
	}
	return fpResp.ID, nil
}

// AddBankAccount verifies the bank account via POA penny drop first.
// On success it registers the account with FP tenant and saves to Firestore.
// Firestore is NOT updated if verification fails.
func AddBankAccount(uid, investorID string, req structs.BankAccountRequest) (string, string, error) {
	user, err := repository.GetSureUserByUID(uid)
	if err != nil {
		return "", "", fmt.Errorf("failed to load user: %w", err)
	}
	if user.PAN == "" {
		return "", "", fmt.Errorf("PAN not found for user")
	}

	acType := req.AccountType
	if acType == "" {
		acType = "savings"
	}

	// Step 1: Create FP bank account
	fpBank, err := FPAddBankAccount(structs.FPBankAccountRequest{
		Profile:                  investorID,
		PrimaryAccountHolderName: user.Name,
		AccountNumber:            req.AccountNumber,
		Type:                     acType,
		IFSCCode:                 req.IFSC,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to register bank account: %w", err)
	}

	// Step 2: POA penny drop verification
	fpPV, err := POACreatePreVerification(structs.FPPreVerificationRequest{
		InvestorIdentifier: user.PAN,
		PAN:                &structs.FPPreVerifField{Value: user.PAN},
		Name:               &structs.FPPreVerifField{Value: user.Name},
		BankAccounts: []structs.FPPreVerifBankAccountItem{
			{
				Value: structs.FPPreVerifBankValue{
					AccountNumber: req.AccountNumber,
					IFSCCode:      req.IFSC,
					AccountType:   acType,
				},
				VerifyManuallyIfRequired: true,
			},
		},
	})
	if err != nil {
		return fpBank.ID, "", fmt.Errorf("bank verification failed: %w", err)
	}

	fpID := fpPV.ID
	pv := &entity.PreVerificationUsage{
		UUID:                uid,
		VerificationType:    "bank_account_verification",
		Pan:                 user.PAN,
		Status:              "pending",
		FpPreVerificationID: &fpID,
		BankIFSC:            strPtr(req.IFSC),
		BankAccountNumber:   strPtr(req.AccountNumber),
		TriggeredBy:         strPtr("bank_verify"),
	}
	if err := repository.CreatePreVerification(pv); err != nil {
		return fpBank.ID, "", err
	}

	// Poll for final status
	if polled, pollErr := PollPreVerification(fpID, 5); pollErr == nil {
		pv.Status = deriveBankVerifStatus(polled)
		pv.PollCount++
		_ = repository.UpdatePreVerification(pv)
	}

	if pv.Status != "completed" {
		return fpID, "", fmt.Errorf("bank verification %s", pv.Status)
	}

	// Step 3: Save to Firestore only on success
	if err := saveUserFPFields(uid, map[string]interface{}{
		"fp_bank_account_id": fpBank.ID,
		"onboarding_step":    2,
	}); err != nil {
		return fpID, "", err
	}

	return fpID, fpBank.ID, nil
}

func AddNominee(uid, investorID string, req structs.NomineeRequest) (string, error) {
	fpReq := structs.FPNomineeRequest{
		InvestorProfileID:            investorID,
		Name:                         req.Name,
		Relationship:                 req.Relation,
		DateOfBirth:                  req.DateOfBirth,
		PAN:                          req.PAN,
		EmailAddress:                 req.EmailAddress,
		AadhaarNumber:                req.AadhaarNumber,
		PassportNumber:               req.PassportNumber,
		DrivingLicenceNumber:         req.DrivingLicenceNumber,
		GuardianName:                 req.GuardianName,
		GuardianPhoneNumber:          req.GuardianPhoneNumber,
		GuardianAddress:              req.GuardianAddress,
		GuardianEmailAddress:         req.GuardianEmailAddress,
		GuardianPAN:                  req.GuardianPAN,
		GuardianAadhaarNumber:        req.GuardianAadhaarNumber,
		GuardianPassportNumber:       req.GuardianPassportNumber,
		GuardianDrivingLicenceNumber: req.GuardianDrivingLicenceNumber,
	}
	if req.PhoneNumber != nil {
		fpReq.PhoneNumber = &structs.FPNomineePhone{
			ISD:    req.PhoneNumber.ISD,
			Number: req.PhoneNumber.Number,
		}
	}
	if req.Address != nil {
		fpReq.Address = &structs.FPNomineeAddr{
			Line1:      req.Address.Line1,
			Line2:      req.Address.Line2,
			City:       req.Address.City,
			State:      req.Address.State,
			PostalCode: req.Address.PostalCode,
			Country:    req.Address.Country,
		}
	}
	fpResp, err := FPAddNominee(fpReq)
	if err != nil {
		return "", err
	}
	if err := saveUserFPFields(uid, map[string]interface{}{
		"fp_nominee_id":   fpResp.ID,
		"onboarding_step": 3,
	}); err != nil {
		return "", err
	}
	return fpResp.ID, nil
}

func ActivateAccount(uid string, fpData *structs.UserFPData, nomineeIdentityProofType string) (string, error) {
	// Step 1: Create or reuse MF investment account
	accountID := fpData.FpInvestmentAccountID
	if accountID == "" {
		fpResp, err := FPCreateMFInvestmentAccount(structs.FPMFInvestmentAccountRequest{
			PrimaryInvestor: fpData.FpInvestorID,
			HoldingPattern:  "single",
		})
		if err != nil {
			return "", err
		}
		accountID = fpResp.ID
	}

	// Step 2: Set folio defaults (bank, phone, email, address, nominee)
	folio := structs.FPFolioDefaults{
		CommunicationEmailAddress: fpData.FpEmailID,
		CommunicationMobileNumber: fpData.FpPhoneID,
		CommunicationAddress:      fpData.FpAddressID,
		PayoutBankAccount:         fpData.FpBankAccountID,
	}
	if fpData.FpNomineeID != "" {
		folio.Nominee1 = fpData.FpNomineeID
		folio.Nominee1AllocationPercent = "100"
		folio.Nominee1IdentityProofType = nomineeIdentityProofType
		folio.NominationsInfoVisibility = "show_nomination_status"
	}
	if err := FPPatchMFInvestmentAccount(accountID, structs.FPMFInvestmentAccountPatchRequest{
		ID:            accountID,
		FolioDefaults: folio,
	}); err != nil {
		return "", fmt.Errorf("folio defaults update failed: %w", err)
	}

	if err := saveUserFPFields(uid, map[string]interface{}{
		"fp_investment_account_id": accountID,
		"onboarding_step":          4,
		"is_activated":             true,
	}); err != nil {
		return "", err
	}
	return accountID, nil
}

func GetOnboardingStatus(uid string) (*structs.OnboardingStatusResponse, error) {
	fpData, err := GetUserFPData(uid)
	if err != nil {
		return nil, err
	}

	resp := &structs.OnboardingStatusResponse{UserFPData: *fpData}

	if kycPV, err := repository.GetLatestPreVerificationByUUIDAndType(uid, "kyc_verification"); err == nil && kycPV.Status == "pending" && kycPV.FpPreVerificationID != nil {
		resp.PendingKYCPreVerifID = *kycPV.FpPreVerificationID
	}
	if bankPV, err := repository.GetLatestPreVerificationByUUIDAndType(uid, "bank_account_verification"); err == nil && bankPV.Status == "pending" && bankPV.FpPreVerificationID != nil {
		resp.PendingBankPreVerifID = *bankPV.FpPreVerificationID
	}

	return resp, nil
}

// GetPreVerificationStatus fetches latest status from DB + FP and updates the DB row.
func GetPreVerificationStatus(fpID string) (*entity.PreVerificationUsage, *structs.FPPreVerification, error) {
	pv, err := repository.GetPreVerificationByFpID(fpID)
	if err != nil {
		return nil, nil, fmt.Errorf("pre-verification not found: %w", err)
	}

	fpPV, err := POAGetPreVerification(fpID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch status from FP: %w", err)
	}

	// Update DB if status changed
	var newStatus string
	if pv.VerificationType == "bank_account_verification" {
		newStatus = deriveBankVerifStatus(fpPV)
	} else {
		newStatus = normalisePreVerifStatus(fpPV)
	}
	if newStatus != pv.Status {
		pv.Status = newStatus
		pv.PollCount++
		_ = repository.UpdatePreVerification(pv)
	}

	return pv, fpPV, nil
}

func strPtr(s string) *string {
	return &s
}

// normGender maps Postgres gender values to FP-accepted values: male | female | transgender
func normGender(g string) string {
	switch g {
	case "M", "m", "male":
		return "male"
	case "F", "f", "female":
		return "female"
	default:
		return "transgender"
	}
}
