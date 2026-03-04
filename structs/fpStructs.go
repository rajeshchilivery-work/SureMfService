package structs

// ---- FP Auth ----

type FPTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// ---- FP Pre-Verification (KYC + Penny Drop) ----

type FPPreVerificationRequest struct {
	PAN               string `json:"pan"`
	Type              string `json:"type"` // "kyc_verification" | "bank_account_verification"
	BankAccountNumber string `json:"bank_account_number,omitempty"`
	BankIFSC          string `json:"bank_ifsc,omitempty"`
}

type FPPreVerification struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	PAN    string `json:"pan"`
	Type   string `json:"type"`
	Result string `json:"result,omitempty"`
}

// ---- FP Investor Profile ----

type FPInvestorProfileRequest struct {
	PAN          string           `json:"pan"`
	Name         string           `json:"name"`
	Gender       string           `json:"gender"`
	DateOfBirth  string           `json:"date_of_birth"`
	CountryOfBirth string         `json:"country_of_birth"`
	Occupation   string           `json:"occupation"`
	Income       string           `json:"income"`
	PEP          bool             `json:"pep"`
	TaxStatus    string           `json:"tax_status"`
	FATCA        []FPFATCADetail  `json:"fatca_details"`
}

type FPFATCADetail struct {
	TaxResidency string `json:"tax_residency"`
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
	Number            string `json:"number"`
	STDCode           string `json:"std_code,omitempty"`
	ISD               string `json:"isd,omitempty"`
	BelongsTo         string `json:"belongs_to"`
}

type FPPhoneResponse struct {
	ID     string `json:"id"`
	Number string `json:"number"`
}

// ---- FP Email ----

type FPEmailRequest struct {
	InvestorProfileID string `json:"profile"`
	Email             string `json:"email"`
	BelongsTo         string `json:"belongs_to"`
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
	City              string `json:"city"`
	State             string `json:"state"`
	Pincode           string `json:"pincode"`
	Country           string `json:"country"`
	AddressType       string `json:"address_type"`
}

type FPAddressResponse struct {
	ID string `json:"id"`
}

// ---- FP Bank Account ----

type FPBankAccountRequest struct {
	InvestorProfileID string `json:"profile"`
	AccountNumber     string `json:"account_number"`
	IFSC              string `json:"ifsc"`
	AccountType       string `json:"account_type"`
}

type FPBankAccountResponse struct {
	ID            string `json:"id"`
	AccountNumber string `json:"account_number"`
	IFSC          string `json:"ifsc"`
	Status        string `json:"status"`
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
