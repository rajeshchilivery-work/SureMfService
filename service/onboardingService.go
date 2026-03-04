package service

import (
	"SureMFService/database/cloudsql/entity"
	"SureMFService/database/cloudsql/repository"
	"SureMFService/database/firebase"
	"SureMFService/structs"
	"fmt"
)

const userFpCollection = "user_fp_collection"

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

func KYCCheck(uid, pan string) (*entity.PreVerificationUsage, error) {
	fpPV, err := POACreatePreVerification(structs.FPPreVerificationRequest{
		PAN:  pan,
		Type: "kyc_verification",
	})
	if err != nil {
		return nil, fmt.Errorf("kyc pre_verification failed: %w", err)
	}

	fpID := fpPV.ID
	pv := &entity.PreVerificationUsage{
		UUID:                uid,
		VerificationType:    "kyc_verification",
		Pan:                 pan,
		Status:              fpPV.Status,
		FpPreVerificationID: &fpID,
		TriggeredBy:         strPtr("kyc_check"),
	}
	if err := repository.CreatePreVerification(pv); err != nil {
		return nil, err
	}
	return pv, nil
}

func CreateInvestorProfile(uid string, req structs.InvestorProfileRequest) (string, error) {
	fpResp, err := FPCreateInvestorProfile(structs.FPInvestorProfileRequest{
		PAN:            req.PAN,
		Name:           req.Name,
		Gender:         req.Gender,
		DateOfBirth:    req.DateOfBirth,
		CountryOfBirth: "IN",
		Occupation:     req.Occupation,
		Income:         req.IncomeSlab,
		PEP:            req.PEP,
		TaxStatus:      "resident_individual",
		FATCA:          []structs.FPFATCADetail{{TaxResidency: "IN"}},
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

func AddPhone(uid, investorID string, req structs.PhoneRequest) (string, error) {
	belongsTo := req.BelongsTo
	if belongsTo == "" {
		belongsTo = "self"
	}
	fpResp, err := FPAddPhone(structs.FPPhoneRequest{
		InvestorProfileID: investorID,
		Number:            req.Number,
		BelongsTo:         belongsTo,
	})
	if err != nil {
		return "", err
	}
	if err := saveUserFPFields(uid, map[string]interface{}{"fp_phone_id": fpResp.ID}); err != nil {
		return "", err
	}
	return fpResp.ID, nil
}

func AddEmail(uid, investorID string, req structs.EmailRequest) (string, error) {
	belongsTo := req.BelongsTo
	if belongsTo == "" {
		belongsTo = "self"
	}
	fpResp, err := FPAddEmail(structs.FPEmailRequest{
		InvestorProfileID: investorID,
		Email:             req.Email,
		BelongsTo:         belongsTo,
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
	addrType := req.AddressType
	if addrType == "" {
		addrType = "residential"
	}
	fpResp, err := FPAddAddress(structs.FPAddressRequest{
		InvestorProfileID: investorID,
		Line1:             req.Line1,
		Line2:             req.Line2,
		City:              req.City,
		State:             req.State,
		Pincode:           req.Pincode,
		Country:           country,
		AddressType:       addrType,
	})
	if err != nil {
		return "", err
	}
	if err := saveUserFPFields(uid, map[string]interface{}{"fp_address_id": fpResp.ID}); err != nil {
		return "", err
	}
	return fpResp.ID, nil
}

func AddBankAccount(uid, investorID string, req structs.BankAccountRequest) (string, error) {
	acType := req.AccountType
	if acType == "" {
		acType = "savings"
	}
	fpResp, err := FPAddBankAccount(structs.FPBankAccountRequest{
		InvestorProfileID: investorID,
		AccountNumber:     req.AccountNumber,
		IFSC:              req.IFSC,
		AccountType:       acType,
	})
	if err != nil {
		return "", err
	}
	if err := saveUserFPFields(uid, map[string]interface{}{
		"fp_bank_account_id": fpResp.ID,
		"onboarding_step":    2,
	}); err != nil {
		return "", err
	}
	return fpResp.ID, nil
}

func VerifyBankAccount(uid, pan string, req structs.BankVerifyRequest) (*entity.PreVerificationUsage, error) {
	fpPV, err := POACreatePreVerification(structs.FPPreVerificationRequest{
		PAN:               pan,
		Type:              "bank_account_verification",
		BankAccountNumber: req.AccountNumber,
		BankIFSC:          req.IFSC,
	})
	if err != nil {
		return nil, err
	}

	fpID := fpPV.ID
	pv := &entity.PreVerificationUsage{
		UUID:                uid,
		VerificationType:    "bank_account_verification",
		Pan:                 pan,
		Status:              fpPV.Status,
		FpPreVerificationID: &fpID,
		BankIFSC:            strPtr(req.IFSC),
		BankAccountNumber:   strPtr(req.AccountNumber),
		TriggeredBy:         strPtr("bank_verify"),
	}
	if err := repository.CreatePreVerification(pv); err != nil {
		return nil, err
	}
	return pv, nil
}

func AddNominee(uid, investorID string, req structs.NomineeRequest) (string, error) {
	fpResp, err := FPAddNominee(structs.FPNomineeRequest{
		InvestorProfileID: investorID,
		Name:              req.Name,
		Relation:          req.Relation,
		DateOfBirth:       req.DateOfBirth,
		AllocationPercent: req.AllocationPercent,
		IsMajor:           req.IsMajor,
	})
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

func ActivateAccount(uid string, fpData *structs.UserFPData) (string, error) {
	nominees := []string{}
	if fpData.FpNomineeID != "" {
		nominees = []string{fpData.FpNomineeID}
	}

	fpResp, err := FPCreateMFInvestmentAccount(structs.FPMFInvestmentAccountRequest{
		InvestorProfileID: fpData.FpInvestorID,
		PrimaryBankID:     fpData.FpBankAccountID,
		HoldingType:       "single",
		Nominees:          nominees,
	})
	if err != nil {
		return "", err
	}

	if err := saveUserFPFields(uid, map[string]interface{}{
		"fp_investment_account_id": fpResp.ID,
		"onboarding_step":          4,
		"is_activated":             true,
	}); err != nil {
		return "", err
	}
	return fpResp.ID, nil
}

func GetOnboardingStatus(uid string) (*structs.UserFPData, error) {
	return GetUserFPData(uid)
}

func strPtr(s string) *string {
	return &s
}
