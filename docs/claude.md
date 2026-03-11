# SureMFService — Internals & Data Model

Architecture details, database schema, FP API integration patterns, and development notes.

---

## Architecture

### 3-Layer Pattern

```
Controller (routes + request parsing)
    ↓
Service (business logic + Firestore/DB ops)
    ↓
FP API wrappers (fpService.go, fpPoaService.go)
```

- **Controllers** (`controller/`) — bind JSON, extract UID/params, call service, return via `utils.HandleResponse(c, data, err, "MF")`
- **Services** (`service/`) — orchestrate FP API calls, read/write Firestore, log events to PostgreSQL
- **FP wrappers** — `fpRequest()` and `poaRequest()` generic HTTP helpers with Bearer token auth

### Key Files

| File | Purpose |
|------|---------|
| `service/fpService.go` | FP Tenant API wrappers (investor, orders, payments, SIP, redemption, portfolio, mandates) |
| `service/fpPoaService.go` | FP POA API wrappers (KYC pre-verification, bank penny drop) |
| `service/onboardingService.go` | Onboarding flow: KYC, profile, contacts, bank, nominee, activation |
| `service/orderService.go` | Order flows: purchase, SIP, redemption, consent, payment, portfolio, cancel |
| `service/mandateService.go` | Mandate CRUD: create, authorize, list, status, cancel |
| `service/creditService.go` | EMI ROI delta: compare current vs market loan rates |
| `database/cloudsql/repository/MfEventsRepo.go` | Event logging + terminal event dedup guard |
| `database/firebase/connection.go` | Firebase init + `SetDocFields()` / `GetDoc()` helpers |
| `database/cloudsql/connection.go` | PostgreSQL (GORM) connection setup |
| `middleware/authMiddleware.go` | Firebase Auth token verification |
| `middleware/corsMiddleware.go` | CORS headers for cross-origin requests |
| `middleware/auditLogMiddleware.go` | Request/response audit logging to message queue |
| `service/fundService.go` | Fund scheme listing and lookup by ISIN |
| `config/config.go` | All env vars loaded into `AppConfig` struct (includes DB connection pool settings) |

---

## Data Model

### Firestore Collections

#### `user_fp_mapping/{uid}`

Stores all FP resource IDs created during onboarding. Written incrementally as each step completes.

| Field | Type | Set at |
|-------|------|--------|
| `fp_investor_id` | string | Investor profile creation |
| `fp_phone_id` | string | Phone registration |
| `fp_email_id` | string | Email registration |
| `fp_address_id` | string | Address registration |
| `fp_bank_account_id` | string | Bank account (after penny drop passes) |
| `fp_nominee_id` | string | Nominee registration |
| `fp_investment_account_id` | string | Account activation |
| `onboarding_step` | int | Progress tracker (1-4) |
| `is_activated` | bool | `true` after activation |

#### `creditData/{uid}`

Retail account data including loans and credit cards. Read-only — written by credit report service. Contains `FirebaseRetailAccount` struct with `LN` (loans) and `CC` (credit cards) arrays.

#### `users/{uid}`

User profile data (name, PAN, DOB, phone, email, gender). Read-only from this service — written by the frontend/auth layer.

### PostgreSQL (`sure-app` database, `sure_mf` schema)

Tables are pre-created — **no auto-migrate**. GORM is used for queries only.

#### `sure_mf.pre_verification_usage`

Tracks KYC and bank penny-drop verifications via FP POA API.

| Column | Type | Notes |
|--------|------|-------|
| `id` | bigserial | PK |
| `uuid` | text | Firebase UID |
| `verification_type` | text | `kyc_verification` or `bank_account_verification` |
| `fp_pre_verification_id` | text | FP POA resource ID |
| `pan` | text | User's PAN |
| `status` | text | `pending` / `completed` / `failed` |
| `result` | text | Raw result detail |
| `bank_ifsc` | text | For bank verifications |
| `bank_account_number` | text | For bank verifications |
| `poll_count` | smallint | Number of status polls |
| `triggered_by` | text | `kyc_check` or `bank_verify` |

#### `sure_mf.otp_activity`

Tracks order OTP confirmation lifecycle.

| Column | Type | Notes |
|--------|------|-------|
| `id` | bigserial | PK |
| `uuid` | text | Firebase UID |
| `order_type` | text | `purchase` / `sip` / `redemption` |
| `fp_order_id` | text | FP order resource ID |
| `status` | text | `initiated` / `confirmed` / `failed` |
| `initiated_at` | timestamp | When OTP was sent |
| `confirmed_at` | timestamp | When OTP was confirmed |
| `resulting_order_state` | text | FP order state after OTP |

#### `sure_mf.email_verification`

Email OTP verification tracking.

| Column | Type | Notes |
|--------|------|-------|
| `id` | bigserial | PK |
| `uuid` | text | Firebase UID |
| `email` | text | Email address |
| `method` | text | Default `otp` |
| `token_hash` | text | Hashed OTP/token |
| `status` | text | `pending` / `verified` |
| `attempt_count` | smallint | Current attempts |
| `max_attempts` | smallint | Default 3 |
| `expires_at` | timestamp | Token expiry |
| `verified_at` | timestamp | When verified |

#### `sure_credit_report.credit_details` (read-only)

| Column | Type | Notes |
|--------|------|-------|
| `id` | bigserial | PK |
| `user_id` | integer | FK to `sure_user.users.id` |
| `score` | bigint | Credit score |

#### `sure_credit_report.interest_rates_v2` (read-only)

| Column | Type | Notes |
|--------|------|-------|
| `id` | bigserial | PK |
| `min_score` | bigint | Minimum credit score range |
| `max_score` | bigint | Maximum credit score range |
| `market_rate` | numeric | Market interest rate |
| `account_type_id` | bigint | 2=Home Loan, 3=HL Top Up, 4=LAP, 5=Personal Loan |
| `is_active` | boolean | Active rate flag |

#### `sure_mf.mf_events`

Audit trail for all order and mandate events.

| Column | Type | Notes |
|--------|------|-------|
| `id` | bigserial | PK |
| `user_id` | text | Firebase UID |
| `event_type` | text | e.g. `purchase_order_created`, `sip_confirmed`, `mandate_cancelled` |
| `fp_entity_id` | text | FP resource ID |
| `isin` | text | Fund ISIN (if applicable) |
| `amount` | numeric | Order amount |
| `units` | numeric | Redemption units |
| `raw_payload` | jsonb | Extra context (payment_id, token_url, etc.) |
| `event_at` | timestamp | When event occurred |

---

## FP API Integration

### Token Caching

Both FP Tenant and POA tokens are cached in memory with `sync.Mutex`. Tokens are refreshed 60 seconds before expiry.

```go
// fpService.go — Tenant token
fpTokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

// fpPoaService.go — POA token
poaTokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)
```

### FP Tenant API (`FP_BASE_URL`)

Base: `https://s.finprim.com`

| Function | Method | Endpoint | Notes |
|----------|--------|----------|-------|
| `FPCreateInvestorProfile` | POST | `/v2/investor_profiles` | Requires FATCA fields |
| `FPAddPhone` | POST | `/v2/phone_numbers` | `isd` without "+" prefix |
| `FPAddEmail` | POST | `/v2/email_addresses` | |
| `FPAddAddress` | POST | `/v2/addresses` | |
| `FPAddBankAccount` | POST | `/v2/bank_accounts` | Field: `profile` (not `investor_profile`) |
| `FPAddNominee` | POST | `/v2/related_parties` | Adult nominees need PAN + contact info |
| `FPCreateMFInvestmentAccount` | POST | `/v2/mf_investment_accounts` | |
| `FPPatchMFInvestmentAccount` | PATCH | `/v2/mf_investment_accounts` | Sets folio defaults + nominee |
| `FPCreatePurchaseOrder` | POST | `/v2/mf_purchases` | |
| `FPCreateSIPOrder` | POST | `/v2/mf_purchase_plans` | `systematic: true` required |
| `FPCreateRedemptionOrder` | POST | `/v2/mf_redemptions` | |
| `FPUpdatePurchaseConsent` | PATCH | `/v2/mf_purchases` | Consent with email + mobile |
| `FPConfirmSIP` | PATCH | `/v2/mf_purchase_plans` | State + consent |
| `FPConfirmRedemption` | PATCH | `/v2/mf_redemptions` | State + consent |
| `FPConfirmPurchaseState` | PATCH | `/v2/mf_purchases` | `state: "confirmed"` |
| `FPGetPurchaseOrder` | GET | `/v2/mf_purchases/{id}` | |
| `FPGetSIPDetail` | GET | `/v2/mf_purchase_plans/{id}` | |
| `FPGetRedemption` | GET | `/v2/mf_redemptions/{id}` | |
| `FPListSIPs` | GET | `/v2/mf_purchase_plans?mf_investment_account={id}` | |
| `FPListRedemptions` | GET | `/v2/mf_redemptions?mf_investment_account={id}` | |
| `FPGetFolios` | GET | `/v2/mf_folios?mf_investment_account={id}` | v2 portfolio API |
| `FPGetSchemeWiseReturns` | POST | `/v2/transactions/reports/scheme_wise_returns` | Body: `{mf_investment_account}` |
| `FPGetInvestmentAccountReturns` | POST | `/v2/transactions/reports/investment_account_wise_returns` | Body: `{mf_investment_account}` |
| `FPGetBankAccount` | GET | `/v2/bank_accounts/{id}` | Fetches `old_id` for payments |
| `FPGetPhone` | GET | `/v2/phone_numbers/{id}` | Auto-consent: fetch phone number |
| `FPGetEmail` | GET | `/v2/email_addresses/{id}` | Auto-consent: fetch email |
| `FPGetMFInvestmentAccount` | GET | `/v2/mf_investment_accounts/{id}` | Fetch account details |
| `FPCancelSIP` | POST | `/v2/mf_purchase_plans/cancel` | Body: `{id, cancellation_code}` |
| `FPGetMandate` | GET | `/api/pg/mandates/{id}` | Single mandate status |
| `FPGetHoldings` | GET | `/api/oms/reports/holdings?investment_account_id={old_id}` | Uses investment account `old_id` (integer) |
| `FPGetFundScheme` | GET | `/api/oms/fund_schemes/{isin}` | Single fund by ISIN |
| `FPListFundSchemes` | GET | `/api/oms/fund_schemes` | With query params (pagination, filters) |
| `FPCreatePayment` | POST | `/api/pg/payments/netbanking` | Uses `old_id` integers |
| `FPCreateMandate` | POST | `/api/pg/mandates` | |
| `FPAuthorizeMandate` | POST | `/api/pg/payments/emandate/auth` | Returns `token_url` |
| `FPListMandates` | GET | `/api/pg/mandates?bank_account_id={old_id}` | Uses bank account `old_id` (integer) |
| `FPCancelMandate` | POST | `/api/pg/mandates/{id}/cancel` | |
| `FPListPurchases` | GET | `/v2/mf_purchases?mf_investment_account={id}` | Lumpsum purchases only |
| `FPListOrders` | GET | `/v2/mf_purchases` + `mf_purchase_plans` + `mf_redemptions` | Combined all order types |
| `FPPatchNominee` | PATCH | `/v2/related_parties/{id}` | Update nominee details |

### FP POA API (`FP_POA_BASE_URL`)

Base: `https://api.sandbox.cybrilla.com`

| Function | Method | Endpoint | Notes |
|----------|--------|----------|-------|
| `POACreatePreVerification` | POST | `/poa/pre_verifications` | KYC check or bank penny drop |
| `POAGetPreVerification` | GET | `/poa/pre_verifications/{id}` | Poll for status |

POA auth uses a separate OAuth endpoint (`FP_POA_AUTH_URL`) with its own client credentials.

### Status Flows

**Pre-verification (KYC/Bank):**
```
accepted → completed (readiness.status = ready)
accepted → completed (readiness.status = failed) → treated as "failed"
accepted → failed
```

**Purchase order (actual observed in sandbox):**
```
under_review → pending (auto, ~1-3s delay) → confirmed (after consent + confirm PATCH) → submitted → successful
```

> **Important:** FP does NOT instantly transition to `pending` after creation. There is a ~1-3 second delay while the order is `under_review`. Consent calls will fail with "order is not in pending state" if called too quickly. See Race Condition Handling below.

**SIP (purchase plan):**
```
pending → confirmed (consent + state) → active → completed
                                       → cancelled (via cancel API)
```

**Redemption:**
```
pending → confirmed (consent + state) → submitted → successful
```

**Mandate:**
```
created → approved (after eNACH auth) → cancelled (optional)
```

---

## Key Patterns

### `fpRequest` / `poaRequest` helpers

Generic HTTP function: `(method, path, body) → ([]byte, statusCode, error)`. Handles token injection, JSON marshaling, 30s timeout. All FP wrapper functions follow the same pattern:

1. Call `fpRequest(method, path, body)`
2. Check `status >= 400` → return error with response body
3. Unmarshal into typed response struct

### List endpoint parsing

FP list endpoints return `{"data": [...]}`. Wrappers unmarshal into an anonymous struct:
```go
var resp struct { Data []T `json:"data"` }
```

### `old_id` pattern

Payment and mandate APIs use integer `old_id` instead of string IDs. The service fetches these server-side:
- `FPGetPurchaseOrder(id)` → `resp.OldID`
- `FPGetBankAccount(id)` → `resp.OldID`

### Mandate creation requirements

`FPCreateMandate` requires:
- `provider_name: "CYBRILLAPOA"` — set by the backend
- `bank_account_id` uses integer `old_id` fetched from FP bank account

### Mandate authorize postback URL

`AuthorizeMandate` constructs the postback URL with dynamic query params:
```go
postbackURL := fmt.Sprintf("%s?mandate_id=%d&uid=%s", config.AppConfig.MandatePostbackURL, mandateID, uid)
```

### Payment creation requirements

`FPCreatePayment` requires:
- `provider_name: "ONDC"` — **mandatory**, FP rejects without it
- `method` must be **uppercase**: `NETBANKING` or `UPI` (backend applies `strings.ToUpper()` as safety net)
- `payment_postback_url` includes dynamic query params: `?order_id={purchaseID}&uid={uid}` — needed by the frontend callback page
- `amc_order_ids` and `bank_account_id` use integer `old_id` values, not string FP IDs

### Race condition handling — consent retry

FP takes ~1-3 seconds to transition a new purchase order from `under_review` to `pending`. The consent PATCH (`/v2/mf_purchases` with consent data) will fail with `"order is not in pending state"` if called immediately after creation.

`UpdatePurchaseConsent` uses a retry loop:
```go
for attempt := 0; attempt < 3; attempt++ {
    fpResp, err = FPUpdatePurchaseConsent(req)
    if err == nil { return fpResp, nil }
    if strings.Contains(err.Error(), "not in pending state") && attempt < 2 {
        time.Sleep(2 * time.Second) // wait for FP state transition
        continue
    }
    return nil, err
}
```

### SIP confirm retry — review_completed

`ConfirmSIP` uses a retry loop similar to purchase consent, but waits for the SIP to reach `review_completed` state:
```go
for attempt := 0; attempt < 3; attempt++ {
    fpResp, err = FPConfirmSIP(req)
    if err == nil { return fpResp, nil }
    if strings.Contains(err.Error(), "review_completed") && attempt < 2 {
        time.Sleep(2 * time.Second)
        continue
    }
}
```

### Redemption confirm retry

`ConfirmRedemption` also retries (3x, 2s delay) when FP returns "not in pending state" — same pattern as purchase consent.

### "Already confirmed" graceful handling

`ConfirmPurchaseState` handles the case where an order is already confirmed (e.g., from a retry or race):
```go
if err != nil {
    existingOrder, getErr := FPGetPurchaseOrder(purchaseID)
    if getErr == nil && existingOrder.State == "confirmed" {
        // Treat as success — order is already in the desired state
        return existingOrder, nil
    }
    return nil, err
}
```

### Silent error pattern (CRITICAL)

The backend returns **HTTP 200 for ALL responses**, including errors. The `CommonResponse` wrapper:
```json
{ "status": 500, "api-status": "error", "msg": "some error message", "data": null }
```
is sent with `c.JSON(http.StatusOK, ...)`. The frontend MUST check the internal `status` field, not just HTTP status. See `SureInvest/src/lib/api.ts` `request()` function which detects `json.status >= 400`.

### Debug logging

`fpService.go` logs raw FP responses for key operations:
- `FPConfirmPurchaseState`: `[DEBUG] FP confirm state: status=%d body=%s`
- `FPCreatePayment`: `[DEBUG] FP create payment: status=%d body=%s`
- `FPGetFolios`: `[DEBUG] FP get folios: status=%d body=%s`
- `FPGetHoldings`: `[DEBUG] FP get holdings (OMS): status=%d body=%s`
- `FPCreateMandate`: `[DEBUG] FP create mandate: status=%d body=%s`
- `FPAuthorizeMandate`: `[DEBUG] FP authorize mandate: status=%d body=%s`

These are invaluable for diagnosing FP API failures that the backend otherwise wraps into generic error messages.

### Event logging

Two helpers write to `sure_mf.mf_events`:

- `logMfEvent(uid, eventType, fpEntityID, isin, amount, units, payload)` — logs any event unconditionally
- `logTerminalEvent(uid, eventType, fpEntityID, isin, amount, units, payload)` — checks `HasTerminalEvent()` first to prevent duplicates on repeated polling

**4 lifecycle phases logged:**

| Phase | Events | Logger |
|-------|--------|--------|
| Created | `*_order_created`, `mandate_created` | `logMfEvent` |
| Confirmed | `*_confirmed` | `logMfEvent` |
| Completed | `*_successful`, `*_failed`, `*_active`, `mandate_approved` | `logTerminalEvent` (deduped) |
| Cancelled | `sip_cancelled`, `mandate_cancelled` | `logMfEvent` |

### Scheme name enrichment

List endpoints (`GetUserOrders`, `ListSIPs`, `ListRedemptions`, `ListPurchases`, `GetPortfolio`) enrich responses with `scheme_name` (and `fund_category` for portfolio) by looking up ISINs via `FPGetFundScheme`. A scheme cache deduplicates lookups for the same ISIN within a single request.

### Auto-consent

`getConsentData(fpData)` fetches email and phone from FP using stored `fp_phone_id` and `fp_email_id` via `FPGetPhone` and `FPGetEmail`. Used by `UpdatePurchaseConsent`, `ConfirmSIP`, and `ConfirmRedemption` — no email/mobile needed in request body.

### Firebase helpers

Service layer uses `firebase.SetDocFields()` and `firebase.GetDoc()` — never imports the Firestore SDK directly. `SetDocFields` uses `MergeAll` for upsert behavior.

### Audit logging middleware

`AuditLogMiddleware` captures request body and response body for every request, then publishes an `AuditLog` struct to the `AUDIT_LOG` message queue via `SureCommonService/clients`. Includes: service name, endpoint URL, HTTP method, request/response bodies, user ID, IP address, timestamp.

### CORS middleware

`CORSMiddleware` sets `Access-Control-Allow-Origin: *` and allows `GET, POST, PUT, PATCH, DELETE, OPTIONS` methods. Handles preflight `OPTIONS` requests with 204 No Content.

### Polling

`PollPreVerification(fpID, maxAttempts)` polls FP POA every 1 second until status is `completed` or `failed`, up to `maxAttempts` retries.

---

## FATCA Requirements

Investor profile creation **must** include these fields for orders to work:

```go
CountryOfBirth:       "IN",
NationalityCountry:   "IN",
CitizenshipCountries: []string{"IN"},
```

Without FATCA fields, FP rejects purchase/SIP orders even though profile creation succeeds.

---

## Sandbox Test Data

| Resource | Value | Notes |
|----------|-------|-------|
| KYC PAN | `ARRPP3751N` | Returns `readiness.status = ready` |
| Bank accounts | Ending in `1195`-`1199` | Pass penny-drop verification |
| Payment method | `NETBANKING` | Must be uppercase. `UPI` also supported |
| Provider name | `ONDC` | Required in payment creation request |
| Mandate type | `E_MANDATE` | Default; eNACH-based |

---

## Environment Variables

See [README.md](../README.md) for full list. Key groupings:

- **Server**: `PORT`
- **PostgreSQL**: `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`, `DB_SSL_MODE`
- **Firebase**: `FIREBASE_PROJECT_ID`, `FIREBASE_CLIENT_EMAIL`, `FIREBASE_PRIVATE_KEY` (inline, not file path)
- **FP Tenant**: `FP_BASE_URL`, `FP_CLIENT_ID`, `FP_CLIENT_SECRET`, `FP_TENANT_ID`
- **FP POA**: `FP_POA_BASE_URL`, `FP_POA_AUTH_URL`, `FP_POA_CLIENT_ID`, `FP_POA_CLIENT_SECRET`
- **MSG91**: `MSG91_AUTH_KEY`, `MSG91_TEMPLATE_ID`
- **Callbacks**: `PAYMENT_POSTBACK_URL`, `MANDATE_POSTBACK_URL`
- **DB Connection Pool** (optional): `DB_MAX_OPEN_CONNS` (default 100), `DB_MAX_IDLE_CONNS` (default 10), `DB_CONN_MAX_LIFETIME` (default 1h), `DB_CONN_MAX_IDLE_TIME` (default 10m)
