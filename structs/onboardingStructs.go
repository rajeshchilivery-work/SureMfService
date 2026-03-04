package structs

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

// KYC Check
type KYCCheckRequest struct {
	PAN string `json:"pan" binding:"required"`
}

// Investor Profile
type InvestorProfileRequest struct {
	PAN          string `json:"pan" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Gender       string `json:"gender" binding:"required"`
	DateOfBirth  string `json:"date_of_birth" binding:"required"`
	Occupation   string `json:"occupation" binding:"required"`
	IncomeSlab   string `json:"income_slab" binding:"required"`
	PEP          bool   `json:"pep"`
	IsIndianResident bool `json:"is_indian_resident"`
}

// Phone
type PhoneRequest struct {
	Number    string `json:"number" binding:"required"`
	BelongsTo string `json:"belongs_to"`
}

// Email
type EmailRequest struct {
	Email     string `json:"email" binding:"required"`
	BelongsTo string `json:"belongs_to"`
}

// Address
type AddressRequest struct {
	Line1       string `json:"line1" binding:"required"`
	Line2       string `json:"line2"`
	City        string `json:"city" binding:"required"`
	State       string `json:"state" binding:"required"`
	Pincode     string `json:"pincode" binding:"required"`
	Country     string `json:"country"`
	AddressType string `json:"address_type"`
}

// Bank Account
type BankAccountRequest struct {
	AccountNumber string `json:"account_number" binding:"required"`
	IFSC          string `json:"ifsc" binding:"required"`
	AccountType   string `json:"account_type"`
}

type BankVerifyRequest struct {
	AccountNumber string `json:"account_number" binding:"required"`
	IFSC          string `json:"ifsc" binding:"required"`
	PAN           string `json:"pan" binding:"required"`
}

// Nominee
type NomineeRequest struct {
	Name              string  `json:"name" binding:"required"`
	Relation          string  `json:"relation" binding:"required"`
	DateOfBirth       string  `json:"date_of_birth"`
	AllocationPercent float64 `json:"allocation_percentage"`
	IsMajor           bool    `json:"is_major"`
}

// Activate
type ActivateRequest struct {
	AgreedTnC bool `json:"agreed_tnc" binding:"required"`
}
