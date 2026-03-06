# API Routes & Flows

Base URL: `http://localhost:9113/sure-mf`

---

## Route Summary

### Public

| Method | Path | Description |
|--------|------|-------------|
| GET | `/ping` | Health check |
| GET | `/funds` | List MF schemes |

### Onboarding — `/:uid/onboarding`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/status` | Get onboarding progress |
| POST | `/kyc-check` | Verify PAN identity |
| GET | `/pre-verification/:fp_id` | Poll pre-verification status |
| POST | `/investor-profile` | Create investor profile |
| POST | `/phone` | Add phone number |
| POST | `/email` | Add email |
| POST | `/address` | Add address |
| POST | `/bank` | Add bank account + penny drop |
| POST | `/nominee` | Add nominee |
| POST | `/activate` | Create MF investment account |

### Orders — `/:uid/orders`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | List all orders |
| POST | `/purchase` | Create lumpsum purchase |
| POST | `/:id/confirm-otp` | Confirm order OTP |
| PATCH | `/:id/consent` | Update purchase consent |
| POST | `/:id/payment` | Create payment (get token_url) |
| PATCH | `/:id/confirm` | Confirm purchase state |
| GET | `/:id/status` | Get purchase status |
| POST | `/sip` | Create SIP |
| GET | `/sips` | List all SIPs |
| GET | `/sips/:id` | Get SIP detail |
| PATCH | `/sips/:id/confirm` | Confirm SIP with consent |
| GET | `/sips/:id/installments` | List SIP installments |
| POST | `/sips/:id/cancel` | Cancel SIP |
| POST | `/redemption` | Create redemption |
| GET | `/redemptions` | List all redemptions |
| GET | `/redemptions/:id` | Get redemption detail |
| PATCH | `/redemptions/:id/confirm` | Confirm redemption with consent |

### Portfolio — `/:uid/portfolio`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Get all folios (v2 API) |
| GET | `/:id` | Get folio detail |

### Holdings — `/:uid/holdings`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/?folio=XXX` | Get holdings by folio (legacy OMS API) |

### Mandates — `/:uid/mandates`

| Method | Path | Description |
|--------|------|-------------|
| POST | `/` | Create mandate |
| POST | `/authorize` | Authorize mandate (get token_url) |
| GET | `/` | List mandates |
| GET | `/:id` | Get mandate status |
| POST | `/:id/cancel` | Cancel mandate |

---

## Frontend Integration

All user-scoped endpoints use the Firebase UID as a path parameter:

```
/sure-mf/:uid/onboarding/...
/sure-mf/:uid/orders/...
/sure-mf/:uid/portfolio/...
/sure-mf/:uid/mandates/...
```

The frontend must include the authenticated user's Firebase UID in the URL. No `Authorization` header is required.

### Polling Pattern (KYC / Bank Verify)

FP pre-verifications are async. The service polls up to 5 times internally (1s interval), then saves the current status to the DB. If still `"pending"`, the frontend should poll:

```
GET /sure-mf/:uid/onboarding/pre-verification/:fp_id
```

Until `status` is `"completed"` or `"failed"`.

---

## Onboarding Endpoints

### `GET /status`

Returns the user's current FP ID mappings and onboarding progress.

**Response:**
```json
{
  "fp_investor_id": "mfp_xxx",
  "fp_bank_account_id": "bna_xxx",
  "onboarding_step": 2,
  "is_activated": false
}
```

### `POST /kyc-check`

Verifies PAN identity against NSDL via FP POA. Name + DOB auto-fetched from PostgreSQL.

**Body:**
```json
{ "pan": "ABCDE1234F" }
```

**Response:**
```json
{
  "fp_pre_verification_id": "pv_xxx",
  "status": "pending"
}
```

**FP API:** `POST /poa/pre_verifications` with `investor_identifier`, `pan.value`, `name.value`, `date_of_birth.value`

### `GET /pre-verification/:fp_id`

**Response:**
```json
{
  "fp_pre_verification_id": "pv_xxx",
  "verification_type": "kyc_verification",
  "status": "completed",
  "fp_status": "completed",
  "pan": { "status": "verified" },
  "readiness": { "status": "completed" },
  "bank_accounts": []
}
```

### `POST /investor-profile`

Creates investor profile in FP. `pan`, `name`, `gender`, `date_of_birth` auto-fetched from DB.

**Body:**
```json
{
  "occupation": "business",
  "income_slab": "above_10lakh_upto_25lakh",
  "source_of_wealth": "salary"
}
```

**Response:** `{ "fp_investor_id": "invp_xxx" }`

**FP API:** `POST /v2/investor_profiles` — includes FATCA fields (`country_of_birth`, `nationality_country`, `citizenship_countries`) hardcoded to `"IN"`.

**Saves to Firestore:** `fp_investor_id`, `onboarding_step: 1`

### `POST /phone`

**Body:**
```json
{ "number": "9876543210", "belongs_to": "self" }
```

**Response:** `{ "fp_phone_id": "phn_xxx" }`

**FP API:** `POST /v2/phone_numbers` with `profile`, `isd` (no "+"), `number`

### `POST /email`

**Body:**
```json
{ "email": "user@example.com", "belongs_to": "self" }
```

**Response:** `{ "fp_email_id": "eml_xxx" }`

**FP API:** `POST /v2/email_addresses` with `profile`, `email`

### `POST /address`

**Body:**
```json
{
  "line1": "123 Main Street",
  "line2": "Apt 4B",
  "city": "Mumbai",
  "state": "MH",
  "pincode": "400001",
  "country": "IN",
  "address_type": "residential"
}
```

**Response:** `{ "fp_address_id": "adr_xxx" }`

**FP API:** `POST /v2/addresses` with `profile`, `postal_code`, `country`

### `POST /bank`

Creates bank account + runs penny-drop verification.

**Body:**
```json
{
  "account_number": "981234591199",
  "ifsc": "HDFC0001234",
  "account_type": "savings"
}
```

> `account_type` defaults to `"savings"`. Valid: `savings`, `current`, `nre`, `nro`.

**FP API flow:**
1. `POST /v2/bank_accounts` with `profile`, `primary_account_holder_name`, `account_number`, `type`, `ifsc_code`
2. `POST /poa/pre_verifications` with `pan`, `name`, `bank_accounts[].value.{account_number, ifsc_code, account_type}`
3. Poll until `bank_accounts[0].status = "verified"`

**Response (success):**
```json
{
  "fp_bank_account_id": "bac_xxx",
  "fp_pre_verification_id": "pv_xxx",
  "verification_status": "completed"
}
```

**Saves to Firestore:** `fp_bank_account_id`, `onboarding_step: 2`

### `POST /nominee`

**Body:**
```json
{
  "name": "Priya Kumar",
  "relation": "spouse",
  "date_of_birth": "1992-08-20",
  "pan": "DFGPX3751K",
  "email_address": "nominee@example.com",
  "phone_number": { "isd": "91", "number": "9876543210" },
  "address": {
    "line1": "123, test street",
    "city": "Anand",
    "state": "Gujarat",
    "postal_code": "388120",
    "country": "in"
  }
}
```

> **Identity fields** (provide exactly one — must match `nominee1_identity_proof_type` in activate):
> `pan`, `aadhaar_number`, `passport_number`, `driving_licence_number`

> **Adult nominees require:** identity proof + email + phone + address. Nominee PAN must differ from investor PAN.

> **Guardian fields** (for minor nominees): `guardian_name`, `guardian_phone_number`, `guardian_address`, `guardian_email_address`, `guardian_pan`

**FP API:** `POST /v2/related_parties`

**Response:** `{ "fp_nominee_id": "relp_xxx" }`

**Saves to Firestore:** `fp_nominee_id`, `onboarding_step: 3`

### `POST /activate`

Creates and configures the MF investment account.

**Body:**
```json
{ "nominee1_identity_proof_type": "pan" }
```

> Must match the identity field provided during nominee creation. Values: `pan`, `aadhaar`, `driving_licence`, `passport`

**FP API flow:**
1. `POST /v2/mf_investment_accounts` (skipped if already exists)
2. `PATCH /v2/mf_investment_accounts` — sets folio defaults (bank, phone, email, address, nominee)

**Response:**
```json
{
  "fp_investment_account_id": "mfia_xxx",
  "is_activated": true
}
```

**Saves to Firestore:** `fp_investment_account_id`, `onboarding_step: 4`, `is_activated: true`

---

## Lumpsum Purchase Flow

### Step 1: `POST /:uid/orders/purchase`

**Body:**
```json
{
  "scheme_id": "INF090I01239",
  "amount": 5000,
  "folio_number": ""
}
```

**FP API:** `POST /v2/mf_purchases` with `mf_investment_account`, `scheme`, `amount`, `user_ip`, `gateway: "ondc"`

**Response:** FP order response with `id`, `state: "pending"`

### Step 2: `POST /:uid/orders/:id/confirm-otp`

**Query params:** `?type=purchase|sip|redemption` (defaults to `purchase`)

**Body:**
```json
{ "otp": "123456" }
```

**FP API:** `POST /v2/mf_purchases/{id}/otp`

### Step 3: `PATCH /:uid/orders/:id/consent`

No request body required. Email and phone are auto-fetched from FP using stored `fp_phone_id` and `fp_email_id`.

**FP API:** `PATCH /v2/mf_purchases` with `id`, `consent: {email, isd_code: "91", mobile}`

### Step 4: `POST /:uid/orders/:id/payment`

**Body:**
```json
{ "method": "NETBANKING" }
```

> `method` — `"NETBANKING"` or `"UPI"`

Server fetches `old_id` from FP for both purchase order and bank account.

**FP API:** `POST /api/pg/payments/netbanking` with `amc_order_ids: [old_id]`, `method`, `payment_postback_url`, `bank_account_id: old_id`

**Response:**
```json
{
  "id": 12345,
  "token_url": "https://payments.finprim.com/..."
}
```

### Step 5: `PATCH /:uid/orders/:id/confirm`

Sets `state: "confirmed"` which enables the payment link.

**FP API:** `PATCH /v2/mf_purchases` with `id`, `state: "confirmed"`

### Step 6: User completes payment

Open `token_url` in browser/webview. FP processes: `submitted` -> `successful`

### Step 7: `GET /:uid/orders/:id/status`

Poll until `state` is `"successful"` or `"failed"`.

**FP API:** `GET /v2/mf_purchases/{id}`

### Step 8: `GET /:uid/holdings?folio=XXX`

View holdings after settlement (T+1 or T+2).

**FP API:** `GET /api/oms/reports/holdings?folios={folio}`

**Response:**
```json
{
  "data": [
    {
      "folio_number": "1234567890",
      "scheme_code": "INF090I01239",
      "scheme_name": "Axis Bluechip Fund",
      "units": 29.5126,
      "nav": 101.6516,
      "market_value": 2999.99
    }
  ]
}
```

---

## SIP Flow

### `POST /:uid/orders/sip` — Create SIP

**Body:**
```json
{
  "scheme_id": "INF084M01044",
  "amount": 1000,
  "frequency": "monthly",
  "sip_date": 5,
  "number_of_installments": 12,
  "mandate_id": "mnd_xxx",
  "folio_number": ""
}
```

> `mandate_id` is optional. If provided, sets `payment_method: "mandate"` and `payment_source: mandate_id`.

**FP API:** `POST /v2/mf_purchase_plans` with:
- `mf_investment_account`, `scheme`, `amount`, `frequency`, `sip_date`
- `systematic: true`, `user_ip`, `auto_generate_installments: true`
- `payment_method: "mandate"`, `payment_source: "mnd_xxx"` (if mandate provided)
- `consent: {email, isd_code, mobile}` (if provided)
- `number_of_installments`, `folio_number`

### `PATCH /:uid/orders/sips/:id/confirm` — Confirm SIP

No request body required. Email and phone are auto-fetched from FP using stored `fp_phone_id` and `fp_email_id`.

**FP API:** `PATCH /v2/mf_purchase_plans` with `id`, `state: "confirmed"`, `consent: {email, isd_code: "91", mobile}`

**Response:** SIP transitions `created` -> `confirmed` -> `active`

### `GET /:uid/orders/sips` — List SIPs

**FP API:** `GET /v2/mf_purchase_plans?mf_investment_account={mfia_id}`

**Response:** Array of SIP detail objects.

### `GET /:uid/orders/sips/:id` — Get SIP Detail

**FP API:** `GET /v2/mf_purchase_plans/{id}`

**Response:**
```json
{
  "id": "mfpp_xxx",
  "state": "active",
  "systematic": true,
  "amount": 1000,
  "frequency": "monthly",
  "sip_date": 5,
  "number_of_installments": 12,
  "remaining_installments": 9,
  "next_instalment_date": "2026-04-05",
  "payment_method": "mandate",
  "payment_source": "mnd_xxx"
}
```

### `GET /:uid/orders/sips/:id/installments` — Get Installments

**FP API:** `GET /v2/mf_purchase_installments?mf_purchase_plan={id}`

**Response:**
```json
{
  "data": [
    {
      "id": "mfpi_xxx",
      "state": "successful",
      "amount": 1000,
      "mf_purchase_plan": "mfpp_xxx",
      "scheduled_on": "2026-03-05",
      "executed_at": "2026-03-05T10:00:00Z"
    }
  ]
}
```

**Installment states:** `pending`, `submitted`, `successful`, `failed`, `skipped`

### `POST /:uid/orders/sips/:id/cancel` — Cancel SIP

**Body:**
```json
{ "cancellation_code": "investment_goal_complete" }
```

> `cancellation_code` is required. Valid values: `amount_not_available`, `investment_returns_not_as_expected`, `exit_load_not_as_expected`, `switch_to_other_scheme`, `fund_manager_changed`, `investment_goal_complete`, `mandate_not_ready`

**FP API:** `POST /v2/mf_purchase_plans/cancel` with `id`, `cancellation_code`

**Response:** SIP detail with `state: "cancelled"`

---

## Redemption Flow

### `POST /:uid/orders/redemption` — Create Redemption

**Mode 1: By amount**
```json
{
  "folio_number": "12345678",
  "scheme_id": "INF084M01044",
  "amount": 2000
}
```

**Mode 2: By units**
```json
{
  "folio_number": "12345678",
  "scheme_id": "INF084M01044",
  "units": 100
}
```

**Mode 3: Full redemption**
```json
{
  "folio_number": "12345678",
  "scheme_id": "INF084M01044",
  "redeem_all": true
}
```

**FP API:** `POST /v2/mf_redemptions` with `mf_investment_account`, `folio_number`, `scheme`, `amount`/`units`/neither (for full), `user_ip`

**Redemption states:** `under_review` -> `pending` -> `confirmed` -> `submitted` -> `successful`

### `PATCH /:uid/orders/redemptions/:id/confirm` — Confirm Redemption

No request body required. Email and phone are auto-fetched from FP using stored `fp_phone_id` and `fp_email_id`.

**FP API:** `PATCH /v2/mf_redemptions` with `id`, `state: "confirmed"`, `consent: {email, isd_code: "91", mobile}`

### `GET /:uid/orders/redemptions` — List Redemptions

**FP API:** `GET /v2/mf_redemptions?mf_investment_account={mfia_id}`

### `GET /:uid/orders/redemptions/:id` — Get Redemption Detail

**FP API:** `GET /v2/mf_redemptions/{id}`

---

## Portfolio

### `GET /:uid/portfolio` — Get All Folios

**FP API:** `GET /v2/mf_folios?mf_investment_account={mfia_id}`

**Response:**
```json
{
  "data": [
    {
      "id": "folio_xxx",
      "number": "12345678",
      "amc": "P",
      "holding_pattern": "single",
      "holdings": {
        "units": 324.25,
        "nav": 15.4321,
        "market_value": 5002.51,
        "invested_value": 5000.00,
        "redeemable_units": 324.25,
        "redeemable_value": 5002.51
      },
      "payout_details": [
        { "scheme": "INF084M01044", "scheme_code": "ABCFP-GR" }
      ]
    }
  ]
}
```

> `holdings` object is only present if units exist. New users may have empty folios.

### `GET /:uid/portfolio/:id` — Get Folio Detail

**FP API:** `GET /v2/mf_folios/{id}`

---

## Mandate Flow

### `POST /:uid/mandates` — Create Mandate

**Body:**
```json
{
  "mandate_type": "E_MANDATE",
  "mandate_limit": 50000
}
```

> `mandate_type` defaults to `"E_MANDATE"`. Values: `E_MANDATE`, `N_MANDATE`

Server fetches bank account `old_id` from FP.

**FP API:** `POST /api/pg/mandates` with `mf_investment_account`, `bank_account_id` (int old_id), `mandate_type`, `mandate_limit`

**Response:**
```json
{
  "id": 99999,
  "status": "created",
  "mandate_type": "E_MANDATE",
  "mandate_limit": 50000
}
```

**Mandate states:** `created` -> `submitted` -> `approved` (or `rejected`, `cancelled`)

### `POST /:uid/mandates/authorize` — Authorize Mandate

**Body:**
```json
{ "mandate_id": 99999 }
```

**FP API:** `POST /api/pg/payments/emandate/auth` with `mandate_id`, `payment_postback_url`

**Response:**
```json
{
  "id": 12345,
  "token_url": "https://bank-gateway/authorize?token=abc123"
}
```

> Open `token_url` in browser for bank authorization.

### `GET /:uid/mandates` — List Mandates

**FP API:** `GET /api/pg/mandates?mf_investment_account={mfia_id}`

**Response:**
```json
{
  "data": [
    {
      "id": 99999,
      "state": "approved",
      "mandate_type": "E_MANDATE",
      "mandate_limit": 50000,
      "umrn": "SBIN12345678901234",
      "start_date": "2026-02-16",
      "end_date": "2036-02-16"
    }
  ]
}
```

### `GET /:uid/mandates/:id` — Get Mandate Status

**FP API:** `GET /api/pg/mandates/{id}`

**Response:**
```json
{
  "id": 99999,
  "status": "approved",
  "mandate_type": "E_MANDATE",
  "mandate_limit": 50000,
  "umrn": "SBIN12345678901234"
}
```

> Logs terminal events: `mandate_approved` (when status is `approved`), `mandate_failed` (when status is `failed` or `rejected`).

### `POST /:uid/mandates/:id/cancel` — Cancel Mandate

**FP API:** `POST /api/pg/mandates/{id}/cancel`

**Response:** `{ "message": "mandate cancelled" }`

---

## Event Audit Trail (mf_events)

Every transaction lifecycle is logged to `sure_mf.mf_events` across 4 phases:

| Phase | Purchase | SIP | Redemption | Mandate |
|-------|----------|-----|------------|---------|
| **Created** | `purchase_order_created` | `sip_order_created` | `redemption_order_created` | `mandate_created` |
| **Confirmed** | `purchase_confirmed` | `sip_confirmed` | `redemption_confirmed` | — |
| **Completed** | `purchase_successful` / `purchase_failed` | `sip_active` / `sip_failed` | `redemption_successful` / `redemption_failed` | `mandate_approved` / `mandate_failed` |
| **Cancelled** | — | `sip_cancelled` | — | `mandate_cancelled` |

**Deduplication:** Terminal events (completed + cancelled phases) are deduplicated — if a terminal event already exists for an FP entity, it won't be logged again. This prevents duplicates from repeated status polling.

**ISIN tracking:** Created events include the fund ISIN (scheme ID) for traceability.

---

## End-to-End Flow Summary

### Onboarding
```
1. POST /kyc-check           -> verify PAN
2. POST /investor-profile    -> create FP profile (with FATCA fields)
3. POST /phone               -> add phone
4. POST /email               -> add email
5. POST /address             -> add address
6. POST /bank                -> bank account + penny drop
7. POST /nominee             -> add nominee (optional)
8. POST /activate            -> create MF investment account
```

### Lumpsum Purchase
```
1. POST /orders/purchase        -> create order
2. POST /orders/:id/confirm-otp -> confirm OTP
3. PATCH /orders/:id/consent    -> update consent
4. POST /orders/:id/payment     -> create payment (get token_url)
5. PATCH /orders/:id/confirm    -> confirm state
6. User completes payment via token_url
7. GET /orders/:id/status       -> poll until successful
8. GET /holdings?folio=XXX      -> view holdings
```

### SIP
```
1. POST /mandates                      -> create mandate (optional)
2. POST /mandates/authorize            -> authorize (get token_url)
3. User authorizes mandate in browser
4. POST /orders/sip                    -> create SIP (with mandate_id)
5. PATCH /orders/sips/:id/confirm      -> confirm SIP (auto-consent)
6. GET /orders/sips/:id                -> poll until active
7. GET /orders/sips/:id/installments   -> track installments
8. POST /orders/sips/:id/cancel        -> cancel SIP (optional)
```

### Redemption
```
1. POST /orders/redemption                  -> create redemption
2. PATCH /orders/redemptions/:id/confirm    -> confirm with consent
3. GET /orders/redemptions/:id              -> poll until successful
```
