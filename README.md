# SureMFService

Backend service for SureInvest's mutual fund onboarding and investment platform. Built in Go using the Gin framework, it integrates with FinPrim (FP) APIs for investor KYC, account creation, and order execution.

---

## Table of Contents

- [Project Overview](#project-overview)
- [Architecture](#architecture)
- [How to Run](#how-to-run)
- [Frontend Integration](#frontend-integration)
- [Data Model](#data-model)
- [API Reference](#api-reference)
- [FP API Integrations](#fp-api-integrations)

---

## Project Overview

SureMFService handles everything after a user completes phone/email authentication in the main SureInvest app:

1. **KYC verification** — validate PAN + identity via FP POA API
2. **Investor profile creation** — register investor with FP
3. **Contact & address registration** — phone, email, address linked to FP profile
4. **Bank account setup** — add bank account + penny-drop verification
5. **Nominee registration** — optional nominee linked to investment account
6. **Account activation** — create MF investment account in FP
7. **Order placement** — lump-sum purchase, SIP creation, redemption

---

## Architecture

```
Frontend (React Native)
    │  UID in URL path (already authenticated via Firebase)
    ▼
SureMFService (Go / Gin)  :9113
    ├── Firebase Firestore   ← user profile data + FP ID mappings
    ├── PostgreSQL (sure-app) ← pre-verification tracking, order events
    └── FinPrim APIs
            ├── Tenant API  (https://s.finprim.com)          ← investor profile, orders
            └── POA API     (https://api.sandbox.cybrilla.com) ← KYC + penny drop
```

**Auth model**: The frontend embeds the Firebase UID in the URL (`/sure-mf/:uid/...`). This service trusts the UID — authentication is handled upstream by the calling app.

---

## How to Run

### Prerequisites

- Go 1.21+
- PostgreSQL (Cloud SQL or local) with `sure-app` database
- Firebase project with Firestore enabled
- FP Tenant + POA API credentials

### Environment Variables

Create a `.env` file in the project root:

```env
# Server
PORT=9113

# PostgreSQL
DB_HOST=<cloud-sql-host>
DB_USER=<db-user>
DB_PASSWORD=<db-password>
DB_NAME=sure-app
DB_PORT=5432
DB_SSL_MODE=require

# Firebase Admin SDK (inline — not a file path)
FIREBASE_PROJECT_ID=<project-id>
FIREBASE_CLIENT_EMAIL=<service-account-email>
FIREBASE_PRIVATE_KEY=<private-key-with-literal-\n>

# FP Tenant API
FP_BASE_URL=https://s.finprim.com
FP_CLIENT_ID=<tenant-client-id>
FP_CLIENT_SECRET=<tenant-client-secret>
FP_TENANT_ID=<tenant-id>

# FP POA API (KYC + penny drop)
FP_POA_BASE_URL=https://api.sandbox.cybrilla.com
FP_POA_AUTH_URL=https://auth.sandbox.cybrilla.com/auth/realms/POA/protocol/openid-connect/token
FP_POA_CLIENT_ID=<poa-client-id>
FP_POA_CLIENT_SECRET=<poa-client-secret>

# MSG91 (OTP for order confirmation)
MSG91_AUTH_KEY=<msg91-auth-key>
MSG91_TEMPLATE_ID=<msg91-template-id>
```

### Build and Run

```bash
go mod tidy
go build ./...
go run main.go
```

Health check:

```bash
curl http://localhost:9113/sure-mf/ping
```

---

## Frontend Integration

### URL Structure

All user-scoped endpoints use the Firebase UID as a path parameter:

```
/sure-mf/:uid/onboarding/...
/sure-mf/:uid/orders/...
```

The frontend must include the authenticated user's Firebase UID in the URL. No `Authorization` header is required — the UID is extracted directly from the URL path.

### Typical Onboarding Flow

```
1. POST /sure-mf/:uid/onboarding/kyc-check           ← verify PAN identity
2. POST /sure-mf/:uid/onboarding/investor-profile     ← register investor with FP
3. POST /sure-mf/:uid/onboarding/phone                ← add phone number
4. POST /sure-mf/:uid/onboarding/email                ← add email
5. POST /sure-mf/:uid/onboarding/address              ← add address
6. POST /sure-mf/:uid/onboarding/bank                 ← penny drop verify + add bank account
7. POST /sure-mf/:uid/onboarding/nominee              ← add nominee (optional)
9. POST /sure-mf/:uid/onboarding/activate             ← create MF investment account
```

After activation the user can place orders:

```
10. POST /sure-mf/:uid/orders/purchase     ← lump-sum buy
11. POST /sure-mf/:uid/orders/sip          ← SIP
12. POST /sure-mf/:uid/orders/redemption   ← redeem
13. POST /sure-mf/:uid/orders/:id/confirm-otp ← confirm order with OTP
```

### Polling Pattern (KYC / Bank Verify)

FP pre-verifications are async. The service polls up to 5 times internally (1s interval) after creating a pre-verification, then saves the current status to the DB.

If the status is still `"pending"` after the initial create response, the frontend should poll:

```
GET /sure-mf/:uid/onboarding/pre-verification/:fp_id
```

Until `status` is `"completed"` or `"failed"`.

---

## Data Model

### Firebase Firestore

#### `users/{uid}` — User profile (populated by main SureInvest app)

| Field | Type   | Description                     |
|-------|--------|---------------------------------|
| NAM   | string | Full name                       |
| EML   | string | Email address                   |
| NUM   | int64  | Phone number                    |
| DOB   | int64  | Date of birth (epoch ms, can be negative for pre-1970) |
| GN    | string | Gender                          |
| STS   | string | Account status (`"ACTIVE"`)     |
| STG   | int    | Onboarding stage                |
| APR   | bool   | Approved flag                   |
| CTS   | int64  | Created timestamp (epoch ms)    |
| RFC   | string | Referral code                   |

#### `user_fp_mapping/{uid}` — FP ID mappings (written by this service)

| Field                    | Type   | Description                        |
|--------------------------|--------|------------------------------------|
| fp_investor_id           | string | FP investor profile ID             |
| fp_phone_id              | string | FP phone resource ID               |
| fp_email_id              | string | FP email resource ID               |
| fp_address_id            | string | FP address resource ID             |
| fp_bank_account_id       | string | FP bank account resource ID        |
| fp_nominee_id            | string | FP nominee resource ID             |
| fp_investment_account_id | string | FP MF investment account ID        |
| onboarding_step          | int    | Progress tracker (1=profile, 2=bank, 3=nominee, 4=activated) |
| is_activated             | bool   | Whether MF account has been created |

### PostgreSQL (`sure-app` database)

> Tables are pre-created — no auto-migration.

#### `sure_user.users` — Read-only user data (owned by main SureInvest service)

| Column       | Type      | Description              |
|--------------|-----------|--------------------------|
| uuid         | string    | Firebase UID             |
| name         | string    | Full name                |
| phone_number | string    | Phone number             |
| pan          | string    | PAN card number          |
| dob          | date      | Date of birth            |
| email        | string    | Email address            |
| gender       | string    | Gender                   |

#### `sure_mf.pre_verification_usage` — KYC and penny-drop tracking

| Column                | Type      | Description                                       |
|-----------------------|-----------|---------------------------------------------------|
| id                    | int64 PK  | Auto-increment                                    |
| uuid                  | string    | Firebase UID                                      |
| verification_type     | string    | `kyc_verification` or `bank_account_verification` |
| fp_pre_verification_id| string    | FP pre-verification ID for polling                |
| pan                   | string    | PAN used                                          |
| status                | string    | `pending` \| `completed` \| `failed`              |
| result                | string    | Optional result detail                            |
| bank_ifsc             | string    | Bank IFSC (bank_account_verification only)        |
| bank_account_number   | string    | Account number (bank_account_verification only)   |
| poll_count            | int16     | Number of times polled                            |
| triggered_by          | string    | `kyc_check` or `bank_verify`                      |
| created_at            | timestamp | Row creation time                                 |

#### `sure_mf.otp_activity` — Order OTP tracking

| Column    | Type      | Description                      |
|-----------|-----------|----------------------------------|
| id        | int64 PK  | Auto-increment                   |
| uuid      | string    | Firebase UID                     |
| order_id  | string    | FP order ID                      |
| status    | string    | `pending` \| `verified` \| `expired` |
| created_at| timestamp | Created time                     |

#### `sure_mf.email_verification` — Email OTP tracking

| Column     | Type      | Description                         |
|------------|-----------|-------------------------------------|
| id         | int64 PK  | Auto-increment                      |
| uuid       | string    | Firebase UID                        |
| email      | string    | Email being verified                |
| otp        | string    | OTP sent                            |
| status     | string    | `pending` \| `verified` \| `expired`|
| created_at | timestamp | Created time                        |

#### `sure_mf.mf_events` — Audit trail of all order events

| Column      | Type      | Description                   |
|-------------|-----------|-------------------------------|
| id          | int64 PK  | Auto-increment                |
| user_id     | string    | Firebase UID                  |
| event_type  | string    | e.g. `purchase`, `sip`, `redemption` |
| fp_entity_id| string    | FP order/entity ID            |
| isin        | string    | Fund ISIN                     |
| amount      | float64   | Order amount                  |
| units       | float64   | Units (redemption)            |
| raw_payload | jsonb     | Full FP response              |
| event_at    | timestamp | Event timestamp               |

---

## API Reference

Base URL: `http://localhost:9113/sure-mf`

### Public

| Method | Path        | Description       |
|--------|-------------|-------------------|
| GET    | `/ping`     | Health check      |
| GET    | `/funds`    | List MF schemes   |

### Onboarding — `/sure-mf/:uid/onboarding`

All endpoints require UID in the URL path.

---

#### `GET /status`

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

---

#### `POST /kyc-check`

Verifies PAN identity against NSDL via FP POA. Fetches user name + DOB from PostgreSQL automatically.

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

> If `status` is `"pending"`, poll `GET /pre-verification/:fp_id` until `completed` or `failed`.

**Data sources:** PAN from request; name + DOB auto-fetched from `sure_user.users`

---

#### `GET /pre-verification/:fp_id`

Fetches latest pre-verification status from DB and refreshes from FP.

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

---

#### `POST /investor-profile`

Creates the investor profile in FP. Must be called after KYC.

**Body:**
```json
{
  "occupation": "business",
  "income_slab": "above_10lakh_upto_25lakh",
  "source_of_wealth": "salary"
}
```

`pan`, `name`, `gender`, `date_of_birth` are auto-fetched from `sure_user.users`. `pep_details` is hardcoded to `"not_applicable"`.

**Response:** `{ "fp_investor_id": "invp_xxx" }`

**Saves to Firestore:** `fp_investor_id`, `onboarding_step: 1`

---

#### `POST /phone`

Adds a phone number to the FP investor profile.

**Body:**
```json
{ "number": "9876543210", "belongs_to": "self" }
```

**Response:** `{ "fp_phone_id": "phn_xxx" }`

> `belongs_to` defaults to `"self"` if omitted.

---

#### `POST /email`

Adds an email to the FP investor profile.

**Body:**
```json
{ "email": "user@example.com", "belongs_to": "self" }
```

**Response:** `{ "fp_email_id": "eml_xxx" }`

---

#### `POST /address`

Adds a residential address.

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

> `country` defaults to `"IN"`, `address_type` defaults to `"residential"` if omitted.

---

#### `POST /bank`

Creates a bank account in FP tenant, then runs penny-drop verification via FP POA API. Saves to Firestore only if verification succeeds. PAN, name auto-fetched from `sure_user.users`.

**Flow:**
1. `POST /v2/bank_accounts` → creates `bac_xxx` in FP tenant
2. `POST /poa/pre_verifications` → starts penny drop, returns `pv_xxx`
3. Poll `/poa/pre_verifications/{pv_xxx}` until `bank_accounts[0].status = "verified"`
4. Save `fp_bank_account_id` + `onboarding_step: 2` to Firestore

**Body:**
```json
{
  "account_number": "981234591199",
  "ifsc": "HDFC0001234",
  "account_type": "savings"
}
```

> `account_type` defaults to `"savings"` if omitted. Valid values: `savings`, `current`, `nre`, `nro`.

**Sandbox test accounts** (account_number suffix):
- `1195–1199` → verified ✓
- `1600` → failed (low_confidence)
- `31XX` → failed (bank_verification_failed)

**Response (success):**
```json
{
  "fp_bank_account_id": "bac_xxx",
  "fp_pre_verification_id": "pv_xxx",
  "verification_status": "completed"
}
```

**Response (verification failed):**
```json
{
  "status": 500,
  "msg": "bank verification failed",
  "fp_pre_verification_id": "pv_xxx"
}
```

**Saves to Firestore:** `fp_bank_account_id`, `onboarding_step: 2`

---

#### `POST /nominee`

Adds a nominee (related party) to the investor profile. One FP API call: `POST /v2/related_parties`.
Frontend is responsible for age validation and determining whether to pass nominee vs guardian fields.

**Body:**
```json
{
  "name": "Priya Kumar",
  "relation": "spouse",
  "date_of_birth": "1992-08-20",
  "pan": "DFGPX3751K",
  "email_address": "nominee@example.com",
  "phone_number": {
    "isd": "91",
    "number": "9876543210"
  },
  "address": {
    "line1": "123, test street",
    "line2": "",
    "city": "Anand",
    "state": "Gujarat",
    "postal_code": "388120",
    "country": "in"
  }
}
```

> **Identity fields** (provide exactly one — must match the `nominee1_identity_proof_type` sent during activate):
> - `pan` — PAN number
> - `aadhaar_number` — Aadhaar number
> - `passport_number` — Passport number
> - `driving_licence_number` — Driving licence number

> **Guardian fields** (for minor nominees): `guardian_name`, `guardian_phone_number`, `guardian_address`, `guardian_email_address`, `guardian_pan`, `guardian_aadhaar_number`, `guardian_passport_number`, `guardian_driving_licence_number`

> `relation` valid values: `father`, `mother`, `spouse`, `son`, `daughter`, `brother`, `sister`, `aunt`, `uncle`, `nephew`, `niece`, `grand_father`, `grand_mother`, `grand_son`, `grand_daughter`, `brother_in_law`, `sister_in_law`, `father_in_law`, `mother_in_law`, `son_in_law`, `daughter_in_law`, `court_appointed_legal_guardian`, `others`

**Response:** `{ "fp_nominee_id": "relp_xxx" }`

**Saves to Firestore:** `fp_nominee_id`, `onboarding_step: 3`

---

#### `POST /activate`

Creates and fully configures the MF investment account. Makes two FP API calls:

1. `POST /v2/mf_investment_accounts` — creates the account (skipped if already exists)
2. `PATCH /v2/mf_investment_accounts` — sets folio defaults (bank, phone, email, address, nominee)

**Body:**
```json
{
  "nominee1_identity_proof_type": "pan"
}
```

> `nominee1_identity_proof_type` — required if nominee was added. Must match the identity field provided during nominee creation. Allowed values: `pan`, `aadhaar`, `driving_licence`, `passport`

**Response:**
```json
{
  "fp_investment_account_id": "mfia_xxx",
  "is_activated": true
}
```

**Folio defaults set automatically from Firestore:**
- `communication_email_address` → fp_email_id
- `communication_mobile_number` → fp_phone_id
- `communication_address` → fp_address_id
- `payout_bank_account` → fp_bank_account_id
- `nominee1` → fp_nominee_id (if set, with allocation 100%)
- `nominee1_identity_proof_type` → from request body
- `nominations_info_visibility` → `show_nomination_status`

**Saves to Firestore:** `fp_investment_account_id`, `onboarding_step: 4`, `is_activated: true`

---

### Orders — `/sure-mf/:uid/orders`

All order endpoints require UID in URL. The FP investment account ID is auto-fetched from Firestore.

---

#### `GET /`

Lists all orders for the user.

---

#### `POST /purchase`

Places a lump-sum mutual fund purchase order.

**Body:**
```json
{
  "scheme_id": "INF090I01239",
  "amount": 5000,
  "folio_number": ""
}
```

**Response:** `{ "order_id": "ord_xxx", "state": "pending" }`

---

#### `POST /sip`

Creates a SIP (Systematic Investment Plan).

**Body:**
```json
{
  "scheme_id": "INF090I01239",
  "amount": 1000,
  "frequency": "monthly",
  "sip_date": 5
}
```

**Response:** `{ "order_id": "ord_xxx", "state": "pending" }`

---

#### `POST /redemption`

Places a redemption (sell) order.

**Body:**
```json
{
  "folio_number": "1234567890",
  "scheme_id": "INF090I01239",
  "amount": 0,
  "units": 10.5,
  "redeem_all": false
}
```

Provide either `amount`, `units`, or `redeem_all: true`.

---

#### `POST /:id/confirm-otp`

Confirms an order by submitting the OTP sent via SMS.

**Body:**
```json
{ "otp": "123456" }
```

---

## FP API Integrations

### Token Caching

Both FP APIs (Tenant and POA) use OAuth2 client credentials. Tokens are **cached in-process memory** — not fetched on every API call.

**Refresh criteria** (checked before every FP request):
```
if cached_token != "" AND now < (token_expiry - 60s)  →  reuse token
else                                                   →  fetch new token
```

The 60-second buffer ensures the token is never used right before it expires. On server restart the cache is empty so the first request fetches a fresh token.

Each API has its own independent cache and mutex:
- `fpAccessToken` / `fpTokenExpiry` — Tenant API token
- `poaAccessToken` / `poaTokenExpiry` — POA API token

---

### FP Tenant API (`FP_BASE_URL` = `https://s.finprim.com`)

Used for investor profile management and order execution.

| FP Endpoint                                    | Used by                |
|------------------------------------------------|------------------------|
| `POST /v2/investor_profiles`                   | CreateInvestorProfile  |
| `POST /v2/phone_numbers`                       | AddPhone               |
| `POST /v2/email_addresses`                     | AddEmail               |
| `POST /v2/addresses`                           | AddAddress             |
| `POST /v2/bank_accounts`                       | AddBankAccount         |
| `POST /v2/related_parties`                     | AddNominee             |
| `POST /v2/mf_investment_accounts`              | ActivateAccount        |
| `GET  /api/oms/fund_schemes`                   | ListFunds              |
| `POST /v2/mf_purchases`                        | PlacePurchaseOrder     |
| `POST /v2/mf_purchase_plans`                   | PlaceSIPOrder          |
| `POST /v2/mf_redemptions`                      | PlaceRedemptionOrder   |
| `POST /v2/mf_purchases/:id/otp`                | ConfirmOTP (purchase)  |
| `POST /v2/mf_purchase_plans/:id/otp`           | ConfirmOTP (SIP)       |
| `POST /v2/mf_redemptions/:id/otp`              | ConfirmOTP (redemption)|

**Auth endpoint:** `POST {FP_BASE_URL}/v2/auth/{FP_TENANT_ID}/token`

---

### FP POA API (`FP_POA_BASE_URL` = `https://api.sandbox.cybrilla.com`)

Used for async identity verification (KYC) and bank penny-drop.

| FP Endpoint                        | Used by                                        |
|------------------------------------|------------------------------------------------|
| `POST /poa/pre_verifications`      | KYCCheck, AddBankAccount                       |
| `GET  /poa/pre_verifications/:id`  | PollPreVerification, GetPreVerificationStatus  |

**Auth endpoint:** `FP_POA_AUTH_URL` (Keycloak — separate from Tenant auth)

**Status flow:**

```
FP Status    → DB Status
"accepted"   → "pending"    (async processing in progress)
"completed"  → "completed"  (unless readiness.status == "failed" → "failed")
"failed"     → "failed"
```

**Bank verification status resolution** (in priority order):
1. `bank_accounts[0].status == "verified"` → `"completed"`
2. `bank_accounts[0].status == "failed"` → `"failed"`
3. `readiness.status == "failed"` → `"failed"`
4. Fall back to top-level status normalization above
