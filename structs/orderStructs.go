package structs

type PurchaseOrderRequest struct {
	SchemeID    string  `json:"scheme_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	FolioNumber string  `json:"folio_number"`
}

type SIPOrderRequest struct {
	SchemeID  string  `json:"scheme_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	Frequency string  `json:"frequency" binding:"required"`
	SIPDate   int     `json:"sip_date" binding:"required"`
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
