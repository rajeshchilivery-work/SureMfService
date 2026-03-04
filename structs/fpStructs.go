package structs

// ---- FP Auth ----

type FPTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// ---- FP Pre-Verification (KYC + Penny Drop) ----

// Request: POST /poa/pre_verifications
type FPPreVerificationRequest struct {
	InvestorIdentifier string                     `json:"investor_identifier,omitempty"`
	PAN                *FPPreVerifField            `json:"pan,omitempty"`
	Name               *FPPreVerifField            `json:"name,omitempty"`
	DateOfBirth        *FPPreVerifField            `json:"date_of_birth,omitempty"`
	BankAccounts       []FPPreVerifBankAccountItem `json:"bank_accounts,omitempty"`
}

type FPPreVerifField struct {
	Value string `json:"value"`
}

type FPPreVerifBankAccountItem struct {
	Value                    FPPreVerifBankValue `json:"value"`
	VerifyManuallyIfRequired bool                `json:"verify_manually_if_required,omitempty"`
}

type FPPreVerifBankValue struct {
	AccountNumber string `json:"account_number"`
	IFSCCode      string `json:"ifsc_code"`
	AccountType   string `json:"account_type"`
}

// Response: GET/POST /poa/pre_verifications
type FPPreVerification struct {
	ID                 string                    `json:"id"`
	Status             string                    `json:"status"` // "accepted" | "completed"
	InvestorIdentifier string                    `json:"investor_identifier"`
	Readiness          *FPPreVerifResult         `json:"readiness,omitempty"`
	PAN                *FPPreVerifResult         `json:"pan,omitempty"`
	Name               *FPPreVerifResult         `json:"name,omitempty"`
	DateOfBirth        *FPPreVerifResult         `json:"date_of_birth,omitempty"`
	BankAccounts       []FPPreVerifBankResult    `json:"bank_accounts,omitempty"`
}

type FPPreVerifResult struct {
	Status string `json:"status"` // "verified" | "failed" | "pending"
	Code   string `json:"code,omitempty"`
	Reason string `json:"reason,omitempty"`
	Value  string `json:"value,omitempty"`
}

type FPPreVerifBankResult struct {
	Status string              `json:"status"` // "verified" | "failed"
	Code   string              `json:"code,omitempty"`
	Reason string              `json:"reason,omitempty"`
	Value  FPPreVerifBankValue `json:"value"`
}

// ---- FP Investor Profile ----

type FPInvestorProfileRequest struct {
	Type                    string         `json:"type"`
	TaxStatus               string         `json:"tax_status"`
	Name                    string         `json:"name"`
	DateOfBirth             string         `json:"date_of_birth"`
	Gender                  string         `json:"gender"`
	Occupation              string         `json:"occupation"`
	PAN                     string         `json:"pan"`
	PlaceOfBirth            string         `json:"place_of_birth"`
	UseDefaultTaxResidences string         `json:"use_default_tax_residences"`
	FirstTaxResidency       FPTaxResidency `json:"first_tax_residency"`
	SourceOfWealth          string         `json:"source_of_wealth,omitempty"`
	IncomeSlab              string         `json:"income_slab"`
	PEPDetails              string         `json:"pep_details"`
}

type FPTaxResidency struct {
	Country     string `json:"country"`
	TaxIDType   string `json:"taxid_type"`
	TaxIDNumber string `json:"taxid_number"`
}

type FPInvestorProfileResponse struct {
	ID     string `json:"id"`
	PAN    string `json:"pan"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ---- FP Phone ----

type FPPhoneRequest struct {
	InvestorProfileID string `json:"profile"`
	ISD               string `json:"isd"`
	Number            string `json:"number"`
}

type FPPhoneResponse struct {
	ID     string `json:"id"`
	Number string `json:"number"`
}

// ---- FP Email ----

type FPEmailRequest struct {
	InvestorProfileID string `json:"profile"`
	Email             string `json:"email"`
}

type FPEmailResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// ---- FP Address ----

type FPAddressRequest struct {
	InvestorProfileID string `json:"profile"`
	Line1             string `json:"line1"`
	Line2             string `json:"line2,omitempty"`
	Line3             string `json:"line3,omitempty"`
	City              string `json:"city,omitempty"`
	State             string `json:"state,omitempty"`
	PostalCode        string `json:"postal_code"`
	Country           string `json:"country"`
	Nature            string `json:"nature,omitempty"`
}

type FPAddressResponse struct {
	ID string `json:"id"`
}

// ---- FP Bank Account ----

type FPBankAccountRequest struct {
	Profile                  string `json:"profile"`
	PrimaryAccountHolderName string `json:"primary_account_holder_name"`
	AccountNumber            string `json:"account_number"`
	Type                     string `json:"type"`
	IFSCCode                 string `json:"ifsc_code"`
}

type FPBankAccountResponse struct {
	ID            string `json:"id"`
	AccountNumber string `json:"account_number"`
	IFSCCode      string `json:"ifsc_code"`
	Type          string `json:"type"`
}

// ---- FP Related Party (Nominee) ----

type FPNomineeRequest struct {
	InvestorProfileID string  `json:"profile"`
	Name              string  `json:"name"`
	Relation          string  `json:"relation"`
	DateOfBirth       string  `json:"date_of_birth,omitempty"`
	AllocationPercent float64 `json:"allocation_percentage"`
	IsMajor           bool    `json:"is_major"`
	Guardian          *FPGuardian `json:"guardian,omitempty"`
}

type FPGuardian struct {
	Name     string `json:"name"`
	Relation string `json:"relation"`
}

type FPNomineeResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ---- FP Investment Account ----

type FPMFInvestmentAccountRequest struct {
	InvestorProfileID string   `json:"profile"`
	PrimaryBankID     string   `json:"primary_bank_account"`
	HoldingType       string   `json:"holding_pattern"`
	Nominees          []string `json:"nominees,omitempty"`
}

type FPMFInvestmentAccountResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ---- FP Fund Schemes (OMS API: GET /api/oms/fund_schemes) ----

type FPFundSchemeListResponse struct {
	FundSchemes    []FPFundScheme `json:"fund_schemes"`
	TotalPages     int            `json:"total_pages"`
	TotalElements  int            `json:"total_elements"`
	Size           int            `json:"size"`
	Number         int            `json:"number"`
	Last           bool           `json:"last"`
	First          bool           `json:"first"`
}

type FPFundScheme struct {
	FundSchemeID                  int                        `json:"fund_scheme_id"`
	Name                          string                     `json:"name"`
	InvestmentOption              string                     `json:"investment_option"`
	PlanType                      string                     `json:"plan_type"`
	MinInitialInvestment          float64                    `json:"min_initial_investment"`
	MinAdditionalInvestment       float64                    `json:"min_additional_investment"`
	InitialInvestmentMultiples    float64                    `json:"initial_investment_multiples"`
	AdditionalInvestmentMultiples float64                    `json:"additional_investment_multiples"`
	MaxInitialInvestment          float64                    `json:"max_initial_investment"`
	MaxAdditionalInvestment       float64                    `json:"max_additional_investment"`
	MinWithdrawalAmount           float64                    `json:"min_withdrawal_amount"`
	MinWithdrawalUnits            float64                    `json:"min_withdrawal_units"`
	MaxWithdrawalAmount           float64                    `json:"max_withdrawal_amount"`
	MaxWithdrawalUnits            float64                    `json:"max_withdrawal_units"`
	SIPAllowed                    bool                       `json:"sip_allowed"`
	SIPFrequencySpecificData      map[string]FPSIPFreqData   `json:"sip_frequency_specific_data"`
	AMCID                         int                        `json:"amc_id"`
	AMCName                       string                     `json:"amc_name"`
	FundCategory                  string                     `json:"fund_category"`
	FundSubCategory               string                     `json:"fund_sub_category"`
	ISIN                          string                     `json:"isin"`
	AMFI                          string                     `json:"amfi_code"`
	SchemeCode                    string                     `json:"scheme_code"`
	DeliveryMode                  string                     `json:"delivery_mode"`
	SwitchAllowed                 bool                       `json:"switch_allowed"`
	RedemptionAllowed             bool                       `json:"redemption_allowed"`
}

type FPSIPFreqData struct {
	MinInstallmentAmount float64 `json:"min_installment_amount"`
	MaxInstallmentAmount float64 `json:"max_installment_amount"`
	InstallmentMultiples float64 `json:"installment_multiples"`
	MinInstallmentCount  int     `json:"min_installment_count"`
	MaxInstallmentCount  int     `json:"max_installment_count"`
	Dates                []int   `json:"dates"`
}

// ---- FP Orders ----

type FPPurchaseOrderRequest struct {
	MFInvestmentAccount string  `json:"mf_investment_account"`
	SchemeID            string  `json:"scheme"`
	Amount              float64 `json:"amount"`
	FolioNumber         string  `json:"folio_number,omitempty"`
}

type FPSIPOrderRequest struct {
	MFInvestmentAccount string  `json:"mf_investment_account"`
	SchemeID            string  `json:"scheme"`
	Amount              float64 `json:"amount"`
	Frequency           string  `json:"frequency"`
	SIPDate             int     `json:"sip_date"`
	InstalmentStartDate string  `json:"instalment_start_date,omitempty"`
}

type FPRedemptionOrderRequest struct {
	MFInvestmentAccount string  `json:"mf_investment_account"`
	FolioNumber         string  `json:"folio_number"`
	SchemeID            string  `json:"scheme"`
	Units               float64 `json:"units,omitempty"`
	Amount              float64 `json:"amount,omitempty"`
	RedeemAll           bool    `json:"redeem_all,omitempty"`
}

type FPOrderResponse struct {
	ID          string  `json:"id"`
	State       string  `json:"state"`
	Amount      float64 `json:"amount,omitempty"`
	FolioNumber string  `json:"folio_number,omitempty"`
}

type FPOTPRequest struct {
	OrderID string `json:"id"`
	OTP     string `json:"otp"`
}
