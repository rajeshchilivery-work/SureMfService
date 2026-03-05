package structs

// Firestore "users" collection document
type UserProfile struct {
	Name   string `firestore:"NAM" json:"name"`
	Email  string `firestore:"EML" json:"email"`
	Phone  int64  `firestore:"NUM" json:"phone"`
	DOB    int64  `firestore:"DOB" json:"dob"`   // epoch ms (negative = pre-1970)
	Gender string `firestore:"GN" json:"gender"`
	Status string `firestore:"STS" json:"status"` // "ACTIVE"
	Stage  int    `firestore:"STG" json:"stage"`
	APR    bool   `firestore:"APR" json:"apr"`
	CTS    int64  `firestore:"CTS" json:"cts"`    // created timestamp ms
	RFC    string `firestore:"RFC" json:"rfc"`    // referral code
}

// Firestore document for user FP references
type UserFPData struct {
	FpInvestorID          string `firestore:"fp_investor_id" json:"fp_investor_id"`
	FpPhoneID             string `firestore:"fp_phone_id" json:"fp_phone_id"`
	FpEmailID             string `firestore:"fp_email_id" json:"fp_email_id"`
	FpAddressID           string `firestore:"fp_address_id" json:"fp_address_id"`
	FpBankAccountID       string `firestore:"fp_bank_account_id" json:"fp_bank_account_id"`
	FpNomineeID           string `firestore:"fp_nominee_id" json:"fp_nominee_id"`
	FpInvestmentAccountID string `firestore:"fp_investment_account_id" json:"fp_investment_account_id"`
	OnboardingStep        int    `firestore:"onboarding_step" json:"onboarding_step"`
	IsActivated           bool   `firestore:"is_activated" json:"is_activated"`
}

// OnboardingStatusResponse combines Firestore FP data with pending pre-verification IDs from Postgres
type OnboardingStatusResponse struct {
	UserFPData
	PendingKYCPreVerifID  string `json:"pending_kyc_pre_verification_id,omitempty"`
	PendingBankPreVerifID string `json:"pending_bank_pre_verification_id,omitempty"`
}

// KYC Check
type KYCCheckRequest struct {
	PAN string `json:"pan" binding:"required"`
}

// Investor Profile
// pan, name, gender, date_of_birth are auto-fetched from Postgres (sure_user.users)
type InvestorProfileRequest struct {
	Occupation     string `json:"occupation" binding:"required"`
	IncomeSlab     string `json:"income_slab" binding:"required"`
	SourceOfWealth string `json:"source_of_wealth"`
}

// Address
type AddressRequest struct {
	Line1   string `json:"line1" binding:"required"`
	Line2   string `json:"line2"`
	City    string `json:"city"`
	State   string `json:"state"`
	Pincode string `json:"pincode" binding:"required"`
	Country string `json:"country"`
	Nature  string `json:"nature"`
}

// Bank Account
type BankAccountRequest struct {
	AccountNumber string `json:"account_number" binding:"required"`
	IFSC          string `json:"ifsc" binding:"required"`
	AccountType   string `json:"account_type"`
}

// Nominee
type NomineeRequest struct {
	Name        string `json:"name" binding:"required"`
	Relation    string `json:"relation" binding:"required"`
	DateOfBirth string `json:"date_of_birth"`
	// Identity fields (adult nominees)
	PAN                  string              `json:"pan"`
	EmailAddress         string              `json:"email_address"`
	AadhaarNumber        string              `json:"aadhaar_number"`
	PassportNumber       string              `json:"passport_number"`
	DrivingLicenceNumber string              `json:"driving_licence_number"`
	PhoneNumber          *NomineePhoneNumber `json:"phone_number"`
	Address              *NomineeAddress     `json:"address"`
	// Guardian fields (minor nominees)
	GuardianName                 string `json:"guardian_name"`
	GuardianPhoneNumber          string `json:"guardian_phone_number"`
	GuardianAddress              string `json:"guardian_address"`
	GuardianEmailAddress         string `json:"guardian_email_address"`
	GuardianPAN                  string `json:"guardian_pan"`
	GuardianAadhaarNumber        string `json:"guardian_aadhaar_number"`
	GuardianPassportNumber       string `json:"guardian_passport_number"`
	GuardianDrivingLicenceNumber string `json:"guardian_driving_licence_number"`
}

type NomineePhoneNumber struct {
	ISD    string `json:"isd"`
	Number string `json:"number"`
}

type NomineeAddress struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// Activate
// Nominee1IdentityProofType: pan | aadhaar | driving_licence | passport
type ActivateRequest struct {
	Nominee1IdentityProofType string `json:"nominee1_identity_proof_type"`
}
