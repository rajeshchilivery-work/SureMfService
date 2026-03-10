package structs

// --- Firebase creditData structs ---

type SlashEMISIP struct {
	EMI int `firestore:"EMI"` // New EMI amount
	IOS int `firestore:"IOS"` // Interest outstanding
	DFA int `firestore:"DFA"` // Debt free age (months) with SIP
}

type SlashTenure struct {
	IOS int `firestore:"IOS"` // Interest outstanding
	RT  int `firestore:"RT"`  // Remaining tenure (months)
	DFA int `firestore:"DFA"` // Debt free age (months) with SIP
}

type ProgressStep struct {
	MID string `firestore:"MID"` // Master enum ID
	CTS int64  `firestore:"CTS"` // Timestamp (epoch ms)
	ETS int64  `firestore:"ETS"` // Event time (epoch ms)
	ICF bool   `firestore:"ICF"` // Is completed flag
}

type PropertyDetails struct {
	PV  int64  `firestore:"PV"`  // Property value
	AL  string `firestore:"AL"`  // Address line
	DIS string `firestore:"DIS"` // District
	PIN string `firestore:"PIN"` // Pincode
	PON string `firestore:"PON"` // PO name
	PTY string `firestore:"PTY"` // Property type
	STA string `firestore:"STA"` // State
}

type FirebaseLoanDetails struct {
	RID int              `firestore:"RID"`          // Retail ID
	BNK string           `firestore:"BNK"`          // Bank name
	TYP string           `firestore:"TYP"`          // Account type
	ATP string           `firestore:"ATP"`          // Actual Account type from CIR
	BAL int              `firestore:"BAL"`          // Balance
	SAN int              `firestore:"SAN"`          // Sanction amount
	RT  int              `firestore:"RT"`           // Remaining tenure (months)
	OT  int              `firestore:"OT"`           // Original tenure (months)
	EMI int              `firestore:"EMI"`          // EMI amount
	IOS int              `firestore:"IOS"`          // Interest outstanding
	POS int              `firestore:"POS"`          // Principal outstanding
	TOS int              `firestore:"TOS"`          // Total outstanding
	ACC string           `firestore:"ACC"`          // Account number
	ROI float64          `firestore:"ROI"`          // Interest rate
	DFD int              `firestore:"DFD"`          // Debt free date (epoch ms)
	OPM int              `firestore:"OPM"`          // Overpaying monthly bank level
	SIP int              `firestore:"SIP"`          // SIP every month
	GN  int              `firestore:"GN"`           // Gain
	LS  int              `firestore:"LS"`           // Loss
	OPT bool             `firestore:"OPT"`          // Is optimized flag
	CTS int64            `firestore:"CTS"`          // Created timestamp
	UTS int64            `firestore:"UTS"`          // Updated timestamp
	SES SlashEMISIP      `firestore:"SES"`          // Savings if optimized with market rate
	STN SlashTenure      `firestore:"STN"`          // Savings if optimized with tenure
	TRK []ProgressStep   `firestore:"TRK"`          // Track progress if optimized
	STS string           `firestore:"STS"`          // Status
	DTO int64            `firestore:"DTO"`          // Date opened (epoch ms)
	DTR int64            `firestore:"DTR"`          // Date reported (epoch ms)
	ATI int              `firestore:"ATI"`          // Account type ID
	DFA int              `firestore:"DFA"`          // Debt free age (months)
	EPC int              `firestore:"EPC"`          // EMI principal component
	EIC int              `firestore:"EIC"`          // EMI interest component
	PPP float64          `firestore:"PPP"`          // Principal paid off percentage
	OIP bool             `firestore:"OIP"`          // Optimisation in progress flag
	PD  *PropertyDetails `firestore:"PD,omitempty"` // Property details (optional)
	ERS string           `firestore:"ERS"`          // EMI ROI Source
}

type FirebaseCreditCardDetails struct {
	RID int     `firestore:"RID"` // Retail ID
	BNK string  `firestore:"BNK"` // Bank name
	TYP string  `firestore:"TYP"` // Account type
	LIM int     `firestore:"LIM"` // Limit
	UTP float64 `firestore:"UTP"` // Utilisation percentage
	ACC string  `firestore:"ACC"` // Account number
	CTS int64   `firestore:"CTS"` // Created timestamp
	UTS int64   `firestore:"UTS"` // Updated timestamp
	STS string  `firestore:"STS"` // Status
	ATP string  `firestore:"ATP"` // Actual account type
	ATI int     `firestore:"ATI"` // Account type ID
}

type FirebaseRetailAccount struct {
	FTS int64                       `firestore:"FTS"` // Financial Type
	SCR int                         `firestore:"SCR"` // Score
	CC  []FirebaseCreditCardDetails `firestore:"CC"`  // Credit Card Details
	LN  []FirebaseLoanDetails       `firestore:"LN"`  // Loan Details
	LPT int64                       `firestore:"LPT"` // Last Pulled timestamp
	SR  string                      `firestore:"SR"`  // Source
	LTS int64                       `firestore:"LTS"` // Last Updated timestamp
	STG int                         `firestore:"STG"` // Stage ID
}

// --- EMI ROI Delta response ---

type EMIROIDeltaItem struct {
	ACC  string  `json:"acc"`   // Account number
	OEMI int     `json:"o_emi"` // Old EMI (current)
	OROI float64 `json:"o_roi"` // Old ROI (current)
	NROI float64 `json:"n_roi"` // New ROI (market rate)
	NEMI int     `json:"n_emi"` // New EMI (at market rate)
}
