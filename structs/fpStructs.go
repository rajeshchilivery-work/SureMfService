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
	CountryOfBirth          string         `json:"country_of_birth"`
	NationalityCountry      string         `json:"nationality_country"`
	CitizenshipCountries    []string       `json:"citizenship_countries"`
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
	OldID         int    `json:"old_id,omitempty"`
	AccountNumber string `json:"account_number"`
	IFSCCode      string `json:"ifsc_code"`
	Type          string `json:"type"`
}

// ---- FP Related Party (Nominee) ----

type FPNomineeRequest struct {
	InvestorProfileID string `json:"profile"`
	Name              string `json:"name"`
	Relationship      string `json:"relationship"`
	DateOfBirth       string `json:"date_of_birth,omitempty"`
	// Identity fields (allowed only if nominee is 18+)
	PAN                    string          `json:"pan,omitempty"`
	EmailAddress           string          `json:"email_address,omitempty"`
	AadhaarNumber          string          `json:"aadhaar_number,omitempty"`
	PassportNumber         string          `json:"passport_number,omitempty"`
	DrivingLicenceNumber   string          `json:"driving_licence_number,omitempty"`
	PhoneNumber            *FPNomineePhone `json:"phone_number,omitempty"`
	Address                *FPNomineeAddr  `json:"address,omitempty"`
	// Guardian fields (for minor nominees)
	GuardianName                  string `json:"guardian_name,omitempty"`
	GuardianPhoneNumber           string `json:"guardian_phone_number,omitempty"`
	GuardianAddress               string `json:"guardian_address,omitempty"`
	GuardianEmailAddress          string `json:"guardian_email_address,omitempty"`
	GuardianPAN                   string `json:"guardian_pan,omitempty"`
	GuardianAadhaarNumber         string `json:"guardian_aadhaar_number,omitempty"`
	GuardianPassportNumber        string `json:"guardian_passport_number,omitempty"`
	GuardianDrivingLicenceNumber  string `json:"guardian_driving_licence_number,omitempty"`
}

type FPNomineePhone struct {
	ISD    string `json:"isd"`
	Number string `json:"number"`
}

type FPNomineeAddr struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type FPNomineeResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ---- FP Investment Account ----

type FPMFInvestmentAccountRequest struct {
	PrimaryInvestor string `json:"primary_investor"`
	HoldingPattern  string `json:"holding_pattern"`
}

type FPMFInvestmentAccountResponse struct {
	ID            string           `json:"id"`
	Status        string           `json:"status"`
	OldID         int              `json:"old_id"`
	FolioDefaults *FPFolioDefaults `json:"folio_defaults,omitempty"`
}

type FPFolioDefaults struct {
	CommunicationEmailAddress string `json:"communication_email_address,omitempty"`
	CommunicationMobileNumber string `json:"communication_mobile_number,omitempty"`
	CommunicationAddress      string `json:"communication_address,omitempty"`
	PayoutBankAccount         string `json:"payout_bank_account,omitempty"`
	Nominee1                  string `json:"nominee1,omitempty"`
	Nominee1AllocationPercent float64 `json:"nominee1_allocation_percentage,omitempty"`
	Nominee1IdentityProofType string `json:"nominee1_identity_proof_type,omitempty"`
	NominationsInfoVisibility string `json:"nominations_info_visibility,omitempty"`
}

type FPMFInvestmentAccountPatchRequest struct {
	ID            string          `json:"id"`
	FolioDefaults FPFolioDefaults `json:"folio_defaults"`
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
	UserIP              string  `json:"user_ip"`
}

type FPSIPOrderRequest struct {
	MFInvestmentAccount      string  `json:"mf_investment_account"`
	SchemeID                 string  `json:"scheme"`
	Amount                   float64 `json:"amount"`
	Frequency                string  `json:"frequency"`
	InstallmentDay           int     `json:"installment_day"`
	Systematic               bool    `json:"systematic,omitempty"`
	UserIP                   string  `json:"user_ip,omitempty"`
	PaymentMethod            string  `json:"payment_method,omitempty"`
	PaymentSource            string  `json:"payment_source,omitempty"`
	AutoGenerateInstallments bool    `json:"auto_generate_installments,omitempty"`
	NumberOfInstallments     int     `json:"number_of_installments,omitempty"`
}

type FPRedemptionOrderRequest struct {
	MFInvestmentAccount string  `json:"mf_investment_account"`
	FolioNumber         string  `json:"folio_number"`
	SchemeID            string  `json:"scheme"`
	Units  float64 `json:"units,omitempty"`
	Amount float64 `json:"amount,omitempty"`
	UserIP string  `json:"user_ip,omitempty"`
}

type FPOrderResponse struct {
	ID                  string  `json:"id"`
	OldID               int     `json:"old_id,omitempty"`
	State               string  `json:"state"`
	Amount              float64 `json:"amount,omitempty"`
	FolioNumber         string  `json:"folio_number,omitempty"`
	Scheme              string  `json:"scheme,omitempty"`
	SchemeName          string  `json:"scheme_name,omitempty"`
	MFInvestmentAccount string  `json:"mf_investment_account,omitempty"`
	CreatedAt           string  `json:"created_at,omitempty"`
}

// ---- FP Consent Update (PATCH /v2/mf_purchases) ----

type FPConsentUpdateRequest struct {
	ID      string          `json:"id"`
	Consent FPConsentDetail `json:"consent"`
}

type FPConsentDetail struct {
	Email   string `json:"email"`
	ISDCode string `json:"isd_code"`
	Mobile  string `json:"mobile"`
}

// ---- FP Payment (POST /api/pg/payments/netbanking) ----

type FPCreatePaymentRequest struct {
	AMCOrderIDs        []int  `json:"amc_order_ids"`
	Method             string `json:"method"`
	PaymentPostbackURL string `json:"payment_postback_url"`
	BankAccountID      int    `json:"bank_account_id"`
	ProviderName       string `json:"provider_name,omitempty"`
}

type FPCreatePaymentResponse struct {
	ID       int    `json:"id"`
	TokenURL string `json:"token_url"`
}

// ---- FP Confirm State (PATCH /v2/mf_purchases) ----

type FPConfirmStateRequest struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

// ---- FP Transaction Report Request (shared by report endpoints) ----

type FPTransactionReportRequest struct {
	MFInvestmentAccount string `json:"mf_investment_account"`
}

// ---- FP SIP Confirm (PATCH /v2/mf_purchase_plans) ----

type FPSIPConfirmRequest struct {
	ID      string          `json:"id"`
	State   string          `json:"state"`
	Consent FPConsentDetail `json:"consent"`
}

type FPSIPDetailResponse struct {
	ID                       string  `json:"id"`
	OldID                    int     `json:"old_id,omitempty"`
	State                    string  `json:"state"`
	Systematic               bool    `json:"systematic"`
	MFInvestmentAccount      string  `json:"mf_investment_account,omitempty"`
	Scheme                   string  `json:"scheme,omitempty"`
	SchemeName               string  `json:"scheme_name,omitempty"`
	FolioNumber              string  `json:"folio_number,omitempty"`
	Amount                   float64 `json:"amount,omitempty"`
	Frequency                string  `json:"frequency,omitempty"`
	SIPDate                  int     `json:"sip_date,omitempty"`
	InstallmentDay           int     `json:"installment_day,omitempty"`
	PaymentMethod            string  `json:"payment_method,omitempty"`
	PaymentSource            string  `json:"payment_source,omitempty"`
	NumberOfInstallments     int     `json:"number_of_installments,omitempty"`
	InstalmentStartDate      string  `json:"instalment_start_date,omitempty"`
	NextInstalmentDate       string  `json:"next_instalment_date,omitempty"`
	RemainingInstallments    int     `json:"remaining_installments,omitempty"`
	AutoGenerateInstallments bool    `json:"auto_generate_installments,omitempty"`
	CreatedAt                string  `json:"created_at,omitempty"`
}

// ---- FP Redemption Confirm (PATCH /v2/mf_redemptions) ----

type FPRedemptionConfirmRequest struct {
	ID      string          `json:"id"`
	State   string          `json:"state"`
	Consent FPConsentDetail `json:"consent"`
}

type FPRedemptionDetailResponse struct {
	ID                  string  `json:"id"`
	OldID               int     `json:"old_id,omitempty"`
	State               string  `json:"state"`
	MFInvestmentAccount string  `json:"mf_investment_account,omitempty"`
	Scheme              string  `json:"scheme,omitempty"`
	SchemeName          string  `json:"scheme_name,omitempty"`
	FolioNumber         string  `json:"folio_number,omitempty"`
	Amount              float64 `json:"amount,omitempty"`
	Units               float64 `json:"units,omitempty"`
	CreatedAt           string  `json:"created_at,omitempty"`
}

// ---- FP Folios / Portfolio (GET /v2/mf_folios) ----

type FPFolioListResponse struct {
	Data []FPFolio `json:"data"`
}

type FPFolio struct {
	ID                  string          `json:"id"`
	FolioNumber         string          `json:"number,omitempty"`
	MFInvestmentAccount string          `json:"mf_investment_account,omitempty"`
	AMC                 string          `json:"amc,omitempty"`
	HoldingPattern      string          `json:"holding_pattern,omitempty"`
	Holdings            *FPFolioHolding `json:"holdings,omitempty"`
	PayoutDetails       []FPPayoutDetail `json:"payout_details,omitempty"`
	SchemeName          string          `json:"scheme_name,omitempty"`
	FundCategory        string          `json:"fund_category,omitempty"`
}

type FPFolioHolding struct {
	Units           float64 `json:"units"`
	NAV             float64 `json:"nav"`
	MarketValue     float64 `json:"market_value"`
	InvestedValue   float64 `json:"invested_value"`
	RedeemableUnits float64 `json:"redeemable_units"`
	RedeemableValue float64 `json:"redeemable_value"`
}

type FPPayoutDetail struct {
	Scheme     string `json:"scheme,omitempty"`
	SchemeCode string `json:"scheme_code,omitempty"`
}

// ---- FP Mandates (POST /api/pg/mandates) ----

type FPCreateMandateRequest struct {
	BankAccountID int     `json:"bank_account_id"`
	MandateType   string  `json:"mandate_type,omitempty"`
	MandateLimit  float64 `json:"mandate_limit,omitempty"`
	ProviderName  string  `json:"provider_name,omitempty"`
}

type FPMandateResponse struct {
	ID              int     `json:"id"`
	Status          string  `json:"status,omitempty"`
	MandateStatus   string  `json:"mandate_status,omitempty"`
	State           string  `json:"state,omitempty"`
	BankAccountID   int     `json:"bank_account_id,omitempty"`
	MandateRef      string  `json:"mandate_ref,omitempty"`
	MandateType     string  `json:"mandate_type,omitempty"`
	MandateLimit    float64 `json:"mandate_limit,omitempty"`
	UMRN            string  `json:"umrn,omitempty"`
	ValidFrom       string  `json:"valid_from,omitempty"`
	ValidTo         string  `json:"valid_to,omitempty"`
	CreatedAt       string  `json:"created_at,omitempty"`
	ApprovedAt      string  `json:"approved_at,omitempty"`
}

type FPMandateListResponse struct {
	Mandates []FPMandateResponse `json:"mandates"`
}

type FPMandateAuthRequest struct {
	MandateID          int    `json:"mandate_id"`
	PaymentPostbackURL string `json:"payment_postback_url,omitempty"`
}

type FPMandateAuthResponse struct {
	ID       int    `json:"id"`
	TokenURL string `json:"token_url"`
}
