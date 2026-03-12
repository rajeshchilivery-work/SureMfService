# Integration Learnings & Gotchas

Hard-won debugging insights from integrating SureMFService with FinPrim (FP) APIs and the SureInvest frontend.

---

## 1. Silent Error Pattern (CRITICAL)

**Problem:** The backend returns HTTP 200 for ALL responses, including errors. The `CommonResponse` wraps errors as:
```json
{ "status": 500, "api-status": "error", "msg": "some error", "data": null, "error": "" }
```
but sends it with `c.JSON(http.StatusOK, ...)`.

**Impact:** The frontend's `fetch()` sees `res.ok === true` (HTTP 200) and proceeds as if the call succeeded. Errors like "confirm failed" or "payment creation failed" were completely invisible.

**Fix:** Frontend `request()` in `src/lib/api.ts` must check the internal `json.status` field:
```typescript
if (json.status && json.status >= 400) {
  return { data: null, error: json.msg || json.error || `Internal error ${json.status}` };
}
```

**Lesson:** Never trust HTTP status alone when the backend uses a response wrapper pattern. Always check internal status codes.

---

## 2. FP Order State Machine — Race Condition

**Problem:** After creating a purchase order (`POST /v2/mf_purchases`), FP returns the order in `under_review` state. It takes ~1-3 seconds to auto-transition to `pending`. The consent PATCH fails with `"order is not in pending state"` if called immediately.

**Observed state flow:**
```
under_review → pending (~1-3s auto) → confirmed (consent + confirm) → submitted → successful
```

**Fix:** `UpdatePurchaseConsent` in `orderService.go` retries up to 3 times with 2-second delays:
```go
for attempt := 0; attempt < 3; attempt++ {
    fpResp, err = FPUpdatePurchaseConsent(req)
    if err == nil { return fpResp, nil }
    if strings.Contains(err.Error(), "not in pending state") && attempt < 2 {
        time.Sleep(2 * time.Second)
        continue
    }
}
```

**Lesson:** FP state transitions are async. Always build retry logic for state-dependent operations.

---

## 3. Payment Method Must Be Uppercase

**Problem:** FP rejects payment creation with lowercase method values like `netbanking`.

**Fix:** Backend applies `strings.ToUpper(method)` before sending to FP. Frontend also sends uppercase (`NETBANKING`, `UPI`).

---

## 4. `provider_name: "ONDC"` Is Required

**Problem:** FP rejects `POST /api/pg/payments/netbanking` without `provider_name` field.

**Fix:** Added `ProviderName: "ONDC"` to `FPCreatePaymentRequest` in `structs/fpStructs.go` and set it in `orderService.go`.

---

## 5. Payment Postback URL Must Include Query Params

**Problem:** After payment completion, FP redirects the user to `PAYMENT_POSTBACK_URL`. The frontend callback page needs to know which order to look up.

**Fix:** The postback URL is constructed dynamically with order and user info:
```go
PaymentPostbackURL: config.AppConfig.PaymentPostbackURL + "?order_id=" + purchaseID + "&uid=" + uid
```

**Frontend callback page** (`/invest/payment-callback`) reads `order_id` and `uid` from URL params, then polls `GET /orders/:id/status` every 5 seconds to show real transaction details.

---

## 6. `scheme_id` Must Be a String (ISIN)

**Problem:** Frontend was sending `scheme_id` as a number (fund's numeric ID). Backend expects a string (ISIN like `INF090I01239`).

**Fix:** Use `fund.isin || isin` (the ISIN from the URL path) instead of `fund.fund_scheme_id` (numeric).

---

## 7. `window.location.href` vs `window.open` for Payment

**Problem:** `window.open(token_url)` caused popup blocker issues and created a confusing UX with two windows.

**Fix:** Use `window.location.href = token_url` to navigate the current tab to the payment gateway. After payment, FP redirects back to the postback URL.

---

## 8. "Already Confirmed" Graceful Handling

**Problem:** If the confirm call (`PATCH state: "confirmed"`) fails but the order is actually already confirmed (e.g., from a retry), the flow gets stuck.

**Fix:** `ConfirmPurchaseState` catches the error, fetches the current order state, and if it's already `confirmed`, returns success:
```go
if err != nil {
    existingOrder, getErr := FPGetPurchaseOrder(purchaseID)
    if getErr == nil && existingOrder.State == "confirmed" {
        return existingOrder, nil  // already in desired state
    }
}
```

---

## 9. `old_id` Pattern for Payment/Mandate APIs

FP payment and mandate APIs use legacy integer IDs (`old_id`), not the string resource IDs used everywhere else. The service fetches these via:
- `FPGetPurchaseOrder(id)` → `resp.OldID` for `amc_order_ids`
- `FPGetBankAccount(id)` → `resp.OldID` for `bank_account_id`

---

## 10. Debug Logging Locations

Key debug log lines added to `fpService.go`:
- `FPConfirmPurchaseState`: `[DEBUG] FP confirm state: status=%d body=%s`
- `FPCreatePayment`: `[DEBUG] FP create payment: status=%d body=%s`

These show the raw FP response before the service processes it — essential for diagnosing FP-side failures.

---

## 11. Auto-Consent Pattern

Consent data (email/phone) is **not** sent by the frontend. The backend auto-fetches it from PostgreSQL `sure_user.users` via `repository.GetSureUserByUID(uid)`:
1. Read `PhoneNumber` and `Email` from PostgreSQL `sure_user.users` (WHERE uuid=?)
2. Include in consent: `{ email, isd_code: "91", mobile }`

This applies to `UpdatePurchaseConsent`, `ConfirmSIP`, and `ConfirmRedemption`.

> **Previously** this used two FP API calls (`FPGetPhone` + `FPGetEmail`) to fetch phone/email from FP using stored `fp_phone_id` and `fp_email_id`. Changed to read directly from PostgreSQL since the values are identical (set once during onboarding and never change).

---

## 12. Frontend Proxy Configuration

SureInvest proxies API calls to the backend via Next.js rewrites in `next.config.ts`:
```
/api/sure-mf/:path* → http://localhost:9113/sure-mf/:path*
```

Frontend API calls use the prefix `/api/sure-mf/` — Next.js strips `/api` and forwards to the Go backend.

---

## Environment Configuration

| Variable | Correct Value | Notes |
|----------|--------------|-------|
| `PAYMENT_POSTBACK_URL` | `http://localhost:3000/invest/payment-callback` | Frontend callback page URL (NOT backend callback) |
| `MANDATE_POSTBACK_URL` | Backend callback URL | For mandate authorization redirects |

---

## Debugging Checklist — Payment Flow

If payment is stuck at "Redirecting to payment screen...":

1. **Check backend logs** for `[DEBUG] FP create payment` and `[DEBUG] FP confirm state`
2. **Check `mf_events` table** — is there a `purchase_confirmed` event? If not, confirm is failing silently
3. **Verify payment method is uppercase** — `NETBANKING` not `netbanking`
4. **Verify `provider_name: "ONDC"` is present** in the payment request
5. **Check if consent succeeded** — look for "not in pending state" errors (race condition)
6. **Check frontend `api.ts`** — is it detecting internal error status codes (`json.status >= 400`)?
7. **Try the full flow with curl** to isolate frontend vs backend issues:
   ```bash
   # 1. Create order
   curl -X POST http://localhost:9113/sure-mf/{uid}/orders/purchase \
     -H "Content-Type: application/json" \
     -d '{"scheme_id":"INF090I01239","amount":500}'

   # 2. Consent (wait 3s after creation for state transition)
   curl -X PATCH http://localhost:9113/sure-mf/{uid}/orders/{order_id}/consent

   # 3. Create payment
   curl -X POST http://localhost:9113/sure-mf/{uid}/orders/{order_id}/payment \
     -H "Content-Type: application/json" \
     -d '{"method":"NETBANKING"}'

   # 4. Confirm
   curl -X PATCH http://localhost:9113/sure-mf/{uid}/orders/{order_id}/confirm
   ```
