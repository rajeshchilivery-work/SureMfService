# API Routes Reference

Base URL: `http://localhost:9113/sure-mf`

All user-scoped endpoints use `/:uid/` path prefix. Firebase Auth token required in `Authorization` header.

---

## Public

### `GET /ping`
Health check. No auth required.

### `GET /funds`
List mutual fund schemes. No auth required.

**Query params:** `investment_option`, `plan_type`, `amc_id`, `page` (default 0), `size` (default 20)

**FP API:** `GET /api/oms/fund_schemes?investment_option=GROWTH&plan_type=Direct&page=0&size=20`

---

### `GET /funds/:isin`
Get fund details by ISIN. No auth required.

**FP API:** `GET /api/oms/fund_schemes/{isin}`

---

## Callbacks (no auth)

### `GET|POST /callbacks/payment`
Payment postback from FP after user completes payment.

### `GET|POST /callbacks/mandate`
Mandate postback from FP after user authorizes mandate.

---

## Onboarding — `/:uid/onboarding`

### `GET /status`
Returns onboarding progress + any pending pre-verification IDs.

No request body.

---

### `POST /kyc-check`

No request body. PAN, name, and DOB are all auto-fetched from PostgreSQL `sure_user.users`.

**FP API:** `POST /poa/pre_verifications`
```json
{
  "investor_identifier": "ARRPP3751N",
  "pan": { "value": "ARRPP3751N" },
  "name": { "value": "John Doe" },
  "date_of_birth": { "value": "1990-01-15" }
}
```

---

### `GET /pre-verification/:fp_id`
Poll pre-verification status. No request body.

**FP API:** `GET /poa/pre_verifications/{fp_id}`

---

### `POST /investor-profile`

**Request body:**
```json
{
  "occupation": "business",
  "income_slab": "above_10lakh_upto_25lakh",
  "source_of_wealth": "salary"
}
```

**FP API:** `POST /v2/investor_profiles`
```json
{
  "type": "individual",
  "tax_status": "individual",
  "name": "John Doe",
  "date_of_birth": "1990-01-15",
  "gender": "male",
  "occupation": "business",
  "pan": "ARRPP3751N",
  "place_of_birth": "IN",
  "country_of_birth": "IN",
  "nationality_country": "IN",
  "citizenship_countries": ["IN"],
  "use_default_tax_residences": "true",
  "first_tax_residency": { "country": "IN", "taxid_type": "pan", "taxid_number": "ARRPP3751N" },
  "source_of_wealth": "salary",
  "income_slab": "above_10lakh_upto_25lakh",
  "pep_details": "not_applicable"
}
```

> PAN, name, DOB, gender auto-fetched from PostgreSQL.

---

### `POST /phone`
No request body. Phone number auto-fetched from PostgreSQL `sure_user.users`.

**FP API:** `POST /v2/phone_numbers`
```json
{
  "profile": "invp_xxx",
  "isd": "91",
  "number": "9876543210"
}
```

---

### `POST /email`
No request body. Email auto-fetched from PostgreSQL `sure_user.users`.

**FP API:** `POST /v2/email_addresses`
```json
{
  "profile": "invp_xxx",
  "email": "user@example.com"
}
```

---

### `POST /address`

**Request body:**
```json
{
  "line1": "123 Main Street",
  "line2": "Apt 4B",
  "city": "Mumbai",
  "state": "MH",
  "pincode": "400001",
  "country": "IN",
  "nature": "residential"
}
```

**FP API:** `POST /v2/addresses`
```json
{
  "profile": "invp_xxx",
  "line1": "123 Main Street",
  "line2": "Apt 4B",
  "city": "Mumbai",
  "state": "MH",
  "postal_code": "400001",
  "country": "IN",
  "nature": "residential"
}
```

---

### `POST /bank`

**Request body:**
```json
{
  "account_number": "981234591199",
  "ifsc": "HDFC0001234",
  "account_type": "savings"
}
```

> `account_type` defaults to `"savings"`. Valid: `savings`, `current`, `nre`, `nro`.

**FP API flow:**

1. `POST /v2/bank_accounts`
```json
{
  "profile": "invp_xxx",
  "primary_account_holder_name": "John Doe",
  "account_number": "981234591199",
  "type": "savings",
  "ifsc_code": "HDFC0001234"
}
```

2. `POST /poa/pre_verifications` (penny drop)
```json
{
  "pan": { "value": "ARRPP3751N" },
  "name": { "value": "John Doe" },
  "bank_accounts": [
    {
      "value": {
        "account_number": "981234591199",
        "ifsc_code": "HDFC0001234",
        "account_type": "savings"
      }
    }
  ]
}
```

3. Poll `GET /poa/pre_verifications/{id}` until bank verified.

---

### `POST /nominee`

**Request body:**
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

> Adult nominees require: identity proof (one of `pan`, `aadhaar_number`, `passport_number`, `driving_licence_number`) + email + phone + address.

**FP API:** `POST /v2/related_parties`
```json
{
  "profile": "invp_xxx",
  "name": "Priya Kumar",
  "relationship": "spouse",
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

---

### `POST /activate`

**Request body:**
```json
{ "nominee1_identity_proof_type": "pan" }
```

> Must match the identity field provided during nominee creation. Values: `pan`, `aadhaar`, `driving_licence`, `passport`.

**FP API flow:**

1. `POST /v2/mf_investment_accounts` (skipped if already exists)
```json
{
  "primary_investor": "invp_xxx",
  "holding_pattern": "single"
}
```

2. `PATCH /v2/mf_investment_accounts`
```json
{
  "id": "mfia_xxx",
  "folio_defaults": {
    "communication_email_address": "eml_xxx",
    "communication_mobile_number": "phn_xxx",
    "communication_address": "adr_xxx",
    "payout_bank_account": "bac_xxx",
    "nominee1": "relp_xxx",
    "nominee1_allocation_percentage": "100",
    "nominee1_identity_proof_type": "pan",
    "nominations_info_visibility": "show_nomination_status"
  }
}
```

---

## Lumpsum Purchase — `/:uid/orders`

### `GET /`
List all orders (purchases + SIPs + redemptions). Enriched with `scheme_name` from FP fund scheme lookup.

No request body.

**FP API:** `GET /v2/mf_purchases`, `/v2/mf_purchase_plans`, `/v2/mf_redemptions` (all with `?mf_investment_account={mfia_id}`)

---

### `GET /purchases`
List lumpsum purchase orders only. Enriched with `scheme_name`.

No request body.

**FP API:** `GET /v2/mf_purchases?mf_investment_account={mfia_id}`

---

### `POST /purchase`

**Request body:**
```json
{
  "scheme_id": "INF090I01239",
  "amount": 5000
}
```

**FP API:** `POST /v2/mf_purchases`
```json
{
  "mf_investment_account": "mfia_xxx",
  "scheme": "INF090I01239",
  "amount": 5000,
  "user_ip": "10.0.128.12"
}
```

---

### `PATCH /:id/consent`
No request body. Email and phone auto-fetched from PostgreSQL `sure_user.users`.

**FP API:** `PATCH /v2/mf_purchases`
```json
{
  "id": "mfp_xxx",
  "consent": {
    "email": "user@example.com",
    "isd_code": "91",
    "mobile": "9876543210"
  }
}
```

---

### `POST /:id/payment`

**Request body:**
```json
{ "method": "NETBANKING" }
```

> Valid: `NETBANKING`, `UPI` — **must be uppercase**. Backend also applies `strings.ToUpper()` as a safety net.

**FP API:** `POST /api/pg/payments/netbanking`
```json
{
  "amc_order_ids": [12345],
  "method": "NETBANKING",
  "payment_postback_url": "http://localhost:3000/invest/payment-callback?order_id=mfp_xxx&uid=firebase_uid",
  "bank_account_id": 131,
  "provider_name": "ONDC"
}
```

> `amc_order_ids` and `bank_account_id` use integer `old_id` values fetched from FP.
> `provider_name: "ONDC"` is **required** — FP rejects payments without it.
> `payment_postback_url` includes `order_id` and `uid` query params so the callback page can fetch order details.

**Response includes `token_url`** — open in browser for payment. Format: `https://sure.s.finprim.com/api/pg/payments/netbanking/ondc?txnId=...`

---

### `PATCH /:id/confirm`
No request body. Sets purchase state to `confirmed`.

**FP API:** `PATCH /v2/mf_purchases`
```json
{
  "id": "mfp_xxx",
  "state": "confirmed"
}
```

---

### `GET /:id/status`
Poll purchase status. No request body.

**FP API:** `GET /v2/mf_purchases/{id}`

---

## SIP — `/:uid/orders`

### `POST /sip`

**Request body (without mandate):**
```json
{
  "scheme_id": "INF084M01044",
  "amount": 1000,
  "frequency": "monthly",
  "sip_date": 5,
  "number_of_installments": 12
}
```

**Request body (with mandate):**
```json
{
  "scheme_id": "INF084M01044",
  "amount": 1000,
  "frequency": "monthly",
  "sip_date": 5,
  "number_of_installments": 12,
  "mandate_id": "mnd_xxx"
}
```

**FP API:** `POST /v2/mf_purchase_plans`
```json
{
  "mf_investment_account": "mfia_xxx",
  "scheme": "INF084M01044",
  "amount": 1000,
  "frequency": "monthly",
  "installment_day": 5,
  "systematic": true,
  "user_ip": "10.0.128.12",
  "auto_generate_installments": true,
  "number_of_installments": 12,
  "payment_method": "mandate",
  "payment_source": "mnd_xxx"
}
```

> `payment_method` and `payment_source` only sent if `mandate_id` is provided.

---

### `PATCH /sips/:id/confirm`
No request body. Auto-consent with email + phone (from PostgreSQL `sure_user.users`).

**FP API:** `PATCH /v2/mf_purchase_plans`
```json
{
  "id": "mfpp_xxx",
  "state": "confirmed",
  "consent": {
    "email": "user@example.com",
    "isd_code": "91",
    "mobile": "9876543210"
  }
}
```

---

### `GET /sips`
List all SIPs. No request body. Enriched with `scheme_name`.

**FP API:** `GET /v2/mf_purchase_plans?mf_investment_account={mfia_id}`

---

### `GET /sips/:id`
Get SIP detail. No request body.

**FP API:** `GET /v2/mf_purchase_plans/{id}`

---

### `POST /sips/:id/cancel`

**Request body:**
```json
{ "cancellation_code": "investment_goal_complete" }
```

> Valid: `amount_not_available`, `investment_returns_not_as_expected`, `exit_load_not_as_expected`, `switch_to_other_scheme`, `fund_manager_changed`, `investment_goal_complete`, `mandate_not_ready`

**FP API:** `POST /v2/mf_purchase_plans/cancel`
```json
{
  "id": "mfpp_xxx",
  "cancellation_code": "investment_goal_complete"
}
```

---

## Redemption — `/:uid/orders`

### `POST /redemption`

**By amount:**
```json
{
  "folio_number": "12345678",
  "scheme_id": "INF084M01044",
  "amount": 2000
}
```

**By units:**
```json
{
  "folio_number": "12345678",
  "scheme_id": "INF084M01044",
  "units": 100
}
```

**Full redemption:**
```json
{
  "folio_number": "12345678",
  "scheme_id": "INF084M01044",
  "redeem_all": true
}
```

**FP API:** `POST /v2/mf_redemptions`
```json
{
  "mf_investment_account": "mfia_xxx",
  "folio_number": "12345678",
  "scheme": "INF084M01044",
  "amount": 2000,
  "user_ip": "10.0.128.12"
}
```

> `folio_number` is required for redemptions — you need to know which folio to redeem from.

---

### `PATCH /redemptions/:id/confirm`
No request body. Auto-consent with email + phone (from PostgreSQL `sure_user.users`).

**FP API:** `PATCH /v2/mf_redemptions`
```json
{
  "id": "mfr_xxx",
  "state": "confirmed",
  "consent": {
    "email": "user@example.com",
    "isd_code": "91",
    "mobile": "9876543210"
  }
}
```

---

### `GET /redemptions`
List all redemptions. No request body. Enriched with `scheme_name`.

**FP API:** `GET /v2/mf_redemptions?mf_investment_account={mfia_id}`

---

### `GET /redemptions/:id`
Get redemption detail. No request body.

**FP API:** `GET /v2/mf_redemptions/{id}`

---

## Portfolio — `/:uid/portfolio`

### `GET /`
Get all folios. No request body.

**FP API:** `GET /v2/mf_folios?mf_investment_account={mfia_id}`

---

---

## Holdings — `/:uid/holdings`

### `GET /`
Get holdings for user. No request body or query params. Investment account old_id is resolved server-side.

**FP API:** `GET /v2/mf_investment_accounts/{mfia_id}` (fetch old_id) → `GET /api/oms/reports/holdings?investment_account_id={old_id}`

---

## Reports — `/:uid/reports`

### `GET /scheme-returns`
Get scheme-wise PnL and returns. No request body.

**FP API:** `POST /v2/transactions/reports/scheme_wise_returns`
```json
{
  "mf_investment_account": "mfia_xxx"
}
```

---

### `GET /account-returns`
Get investment account-level PnL summary. No request body.

**FP API:** `POST /v2/transactions/reports/investment_account_wise_returns`
```json
{
  "mf_investment_account": "mfia_xxx"
}
```

---

## Mandates — `/:uid/mandates`

### `POST /`

**Request body:**
```json
{
  "mandate_type": "E_MANDATE",
  "mandate_limit": 50000
}
```

> `mandate_type` defaults to `"E_MANDATE"`. Values: `E_MANDATE`, `N_MANDATE`.

**FP API:** `POST /api/pg/mandates`
```json
{
  "bank_account_id": 131,
  "mandate_type": "E_MANDATE",
  "mandate_limit": 50000,
  "provider_name": "CYBRILLAPOA"
}
```

> `bank_account_id` is the integer `old_id` fetched from FP bank account.
> `provider_name: "CYBRILLAPOA"` is set by the backend.

---

### `POST /authorize`

**Request body:**
```json
{ "mandate_id": 53 }
```

**FP API:** `POST /api/pg/payments/emandate/auth`
```json
{
  "mandate_id": 53,
  "payment_postback_url": "http://localhost:3000/invest/mandate-callback?mandate_id=53&uid=firebase_uid"
}
```

**Response includes `token_url`** — open in browser for bank authorization.

---

### `GET /`
List all mandates. No request body. Service fetches bank account `old_id` from FP internally.

**FP API:** `GET /v2/bank_accounts/{fp_bank_account_id}` → `GET /api/pg/mandates?bank_account_id={old_id}`

---

### `GET /:id`
Get mandate status. No request body.

**FP API:** `GET /api/pg/mandates/{id}`

---

### `POST /:id/cancel`
Cancel mandate. No request body.

**FP API:** `POST /api/pg/mandates/{id}/cancel`

---

## Credit — `/:uid/credit`

### `GET /emi-roi-delta`

Returns EMI and ROI comparison for eligible loans (ATI 2, 3, 4: Home Loan, HL Top Up, LAP).

No request body.

**Data sources:**
1. PostgreSQL `sure_user.users` — resolve UID to user ID
2. PostgreSQL `sure_credit_report.credit_details` — get credit score by user ID
3. Firebase `creditData/{uid}` — get retail account with loan details
4. PostgreSQL `sure_credit_report.interest_rates_v2` — get market rate by account type ID and credit score range

**Response:**
```json
[
  {
    "acc": "LOAN123456",
    "o_emi": 25000,
    "o_roi": 9.5,
    "n_roi": 8.3,
    "n_emi": 23500
  }
]
```

| Field | Description |
|-------|-------------|
| `acc` | Account number |
| `o_emi` | Old (current) EMI |
| `o_roi` | Old (current) rate of interest |
| `n_roi` | New (market) rate of interest from `interest_rates_v2` |
| `n_emi` | New EMI at market rate (from Firebase `SES.EMI`) |
