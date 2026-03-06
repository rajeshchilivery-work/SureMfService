package structs

type PurchaseOrderRequest struct {
	SchemeID    string  `json:"scheme_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	FolioNumber string  `json:"folio_number"`
}

type SIPOrderRequest struct {
	SchemeID             string  `json:"scheme_id" binding:"required"`
	Amount               float64 `json:"amount" binding:"required"`
	Frequency            string  `json:"frequency" binding:"required"`
	SIPDate              int     `json:"sip_date" binding:"required"`
	NumberOfInstallments int     `json:"number_of_installments"`
	MandateID            string  `json:"mandate_id"`
	FolioNumber          string  `json:"folio_number"`
	Email                string  `json:"email"`
	Mobile               string  `json:"mobile"`
}

type RedemptionOrderRequest struct {
	FolioNumber string  `json:"folio_number" binding:"required"`
	SchemeID    string  `json:"scheme_id" binding:"required"`
	Amount      float64 `json:"amount"`
	Units       float64 `json:"units"`
	RedeemAll   bool    `json:"redeem_all"`
}

type ConfirmOTPRequest struct {
	OTP string `json:"otp" binding:"required"`
}

type OrderResponse struct {
	OrderID string `json:"order_id"`
	State   string `json:"state"`
	Message string `json:"message"`
}

type ConsentUpdateRequest struct {
	Email  string `json:"email" binding:"required"`
	Mobile string `json:"mobile" binding:"required"`
}

type CreatePaymentRequest struct {
	Method string `json:"method" binding:"required"` // "UPI" or "NETBANKING"
}

type SIPConfirmRequest struct {
	Email  string `json:"email" binding:"required"`
	Mobile string `json:"mobile" binding:"required"`
}

type RedemptionConsentRequest struct {
	Email  string `json:"email" binding:"required"`
	Mobile string `json:"mobile" binding:"required"`
}

type CreateMandateRequest struct {
	MandateType  string  `json:"mandate_type"`
	MandateLimit float64 `json:"mandate_limit" binding:"required"`
}

type AuthorizeMandateRequest struct {
	MandateID int `json:"mandate_id" binding:"required"`
}

type CancelSIPRequest struct {
	CancellationCode string `json:"cancellation_code" binding:"required"`
}
