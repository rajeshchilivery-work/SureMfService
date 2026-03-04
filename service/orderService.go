package service

import (
	"SureMFService/database/cloudsql/entity"
	"SureMFService/database/cloudsql/repository"
	"SureMFService/structs"
	"fmt"
	"time"
)

func PlacePurchaseOrder(uid string, fpData *structs.UserFPData, req structs.PurchaseOrderRequest) (*structs.FPOrderResponse, error) {
	if fpData.FpInvestmentAccountID == "" {
		return nil, fmt.Errorf("user has no active investment account")
	}

	fpResp, err := FPCreatePurchaseOrder(structs.FPPurchaseOrderRequest{
		MFInvestmentAccount: fpData.FpInvestmentAccountID,
		SchemeID:            req.SchemeID,
		Amount:              req.Amount,
		FolioNumber:         req.FolioNumber,
	})
	if err != nil {
		return nil, err
	}

	logMfEvent(uid, "purchase_order_created", fpResp.ID, "", req.Amount, 0, nil)
	return fpResp, nil
}

func PlaceSIPOrder(uid string, fpData *structs.UserFPData, req structs.SIPOrderRequest) (*structs.FPOrderResponse, error) {
	if fpData.FpInvestmentAccountID == "" {
		return nil, fmt.Errorf("user has no active investment account")
	}

	fpResp, err := FPCreateSIPOrder(structs.FPSIPOrderRequest{
		MFInvestmentAccount: fpData.FpInvestmentAccountID,
		SchemeID:            req.SchemeID,
		Amount:              req.Amount,
		Frequency:           req.Frequency,
		SIPDate:             req.SIPDate,
	})
	if err != nil {
		return nil, err
	}

	logMfEvent(uid, "sip_order_created", fpResp.ID, "", req.Amount, 0, nil)
	return fpResp, nil
}

func PlaceRedemptionOrder(uid string, fpData *structs.UserFPData, req structs.RedemptionOrderRequest) (*structs.FPOrderResponse, error) {
	if fpData.FpInvestmentAccountID == "" {
		return nil, fmt.Errorf("user has no active investment account")
	}

	fpResp, err := FPCreateRedemptionOrder(structs.FPRedemptionOrderRequest{
		MFInvestmentAccount: fpData.FpInvestmentAccountID,
		FolioNumber:         req.FolioNumber,
		SchemeID:            req.SchemeID,
		Amount:              req.Amount,
		Units:               req.Units,
		RedeemAll:           req.RedeemAll,
	})
	if err != nil {
		return nil, err
	}

	logMfEvent(uid, "redemption_order_created", fpResp.ID, "", req.Amount, 0, nil)
	return fpResp, nil
}

func ConfirmOrderOTP(uid, orderID, orderType, otp string) error {
	if err := FPConfirmOTP(orderType, orderID, otp); err != nil {
		// Update otp_activity as failed
		updateOtpActivity(orderID, "failed", "")
		return err
	}

	updateOtpActivity(orderID, "confirmed", "confirmed")
	logMfEvent(uid, "otp_confirmed", orderID, "", 0, 0, nil)
	return nil
}

func GetUserOrders(fpData *structs.UserFPData) ([]byte, error) {
	if fpData.FpInvestmentAccountID == "" {
		return []byte("[]"), nil
	}
	return FPListOrders(fpData.FpInvestmentAccountID)
}

func logMfEvent(uid, eventType, fpEntityID, isin string, amount, units float64, payload map[string]interface{}) {
	event := &entity.MfEvent{
		UserID:    uid,
		EventType: eventType,
		EventAt:   time.Now(),
	}
	if fpEntityID != "" {
		event.FpEntityID = strPtr(fpEntityID)
	}
	if isin != "" {
		event.ISIN = strPtr(isin)
	}
	if amount != 0 {
		event.Amount = &amount
	}
	if units != 0 {
		event.Units = &units
	}
	if payload != nil {
		event.RawPayload = entity.JSONB(payload)
	}
	_ = repository.CreateMfEvent(event)
}

func updateOtpActivity(fpOrderID, status, orderState string) {
	otp, err := repository.GetOtpActivityByFpOrderID(fpOrderID)
	if err != nil {
		return
	}
	otp.Status = status
	if orderState != "" {
		otp.ResultingOrderState = strPtr(orderState)
	}
	now := time.Now()
	if status == "confirmed" {
		otp.ConfirmedAt = &now
	}
	_ = repository.UpdateOtpActivity(otp)
}
