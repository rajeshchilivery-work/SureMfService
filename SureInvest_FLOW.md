# SureInvest — User Navigation & Transaction Flow

> Reference document for integrating SureMFService backend with the SureInvest frontend.
> Frontend runs at `http://localhost:3000` | Backend at `http://localhost:9113`

---

## Quick Test

```
GET http://localhost:9113/sure-mf/ping
→ { "status": 200, "msg": "pong", "service": "SureMFService" }
```

---

## App Entry Point

`/invest` — Dashboard page

- If user has **no completed transaction**: shows loan optimization card (Old Rate 8.5% → New Rate 7.3%)
- If user has **at least one successful order**: shows Portfolio Overview card with P&L
- Always shows: Explore Funds, Dashboard Analysis, SURE's Top Picks (3 funds)

---

## Full SIP Investment Flow

### Step 0 — Dashboard
```
/invest
  └── "Explore Funds" → /invest/funds
```

### Step 1 — Browse Funds
```
/invest/funds
  - Search bar filters by name, AMC, category
  - Each fund card shows: NAV, min SIP, AUM, expense ratio
  - Click any fund card → /invest/funds/{fund_id}
```

### Step 2 — Fund Detail
```
/invest/funds/{fund_id}
  State: 'detail'
  - Shows fund info: category, plan, exit load, lock-in
  - Shows EMI savings projections (3/5/7/10 year)
  - Scheme selector: Growth / IDCW Payout / IDCW Reinvestment
  - Two CTAs: "Start SIP" | "Invest Lumpsum"
```

### Step 3 — Profile Check (if not complete)
```
Clicking "Start SIP" or "Invest Lumpsum" triggers isProfileComplete() check.

If profile.isComplete !== true:
  → Redirect to /invest/profile?returnTo=/invest/funds/{fund_id}

Profile is 4 steps:
  Step 1: KYC / Confirm Details
    - PAN, Name, DOB, Gender, Mobile, Email (verify), Address
    - Occupation, Income Slab, Source of Wealth
    - Declaration checkbox (Indian Resident, not PEP, T&C)

  Step 2: Bank Account
    - Pre-filled bank card shown (verified via penny drop)
    - Click card to select → auto-advances after 500ms

  Step 3: Nominee
    - Name, Relation, DOB, Phone, Email, Address, ID Proof
    - "Skip for now" option available

  Step 4: Activation
    - Checklist: KYC ✓, Profile ✓, Bank ✓, Nominee ✓ (or Skipped)
    - T&C checkbox
    - "Activate Account →" → redirects back to returnTo URL
```

### Step 4 — SIP Form
```
/invest/funds/{fund_id}
  State: 'invest' → SIP tab

  Fields:
    - Amount input (min SIP shown below)
    - Frequency: Monthly / Weekly / Quarterly (if fund supports)
    - SIP Date: numbered buttons (1–31 for monthly, etc.)
    - Duration: 6 Months / 1 Year / 2 Years / 3 Years / Perpetual
    - Shows: Min SIP Amount, Total Installments

  Checkboxes (both required):
    □ "I authorize this purchase and consent to OTP"
    □ "I have read and agree to T&Cs"

  Button: "Start SIP of ₹{amount}/—" (enabled when all filled + both checked)
```

### Step 5 — OTP Verification
```
State: 'otp'
  - 6-digit OTP input fields (auto-focus, auto-advance)
  - Timer: 27 seconds countdown → "Resend OTP" available after expiry
  - Demo OTP: 000000
  - Error on wrong OTP: "Invalid OTP. Use 000000 for demo."
  - "Verify OTP" button (enabled when all 6 digits filled)

On success:
  - Creates Order { id: mfp_[28hex], status: 'pending', ... }
  - Moves to state: 'order'
```

### Step 6 — Order Pending Screen
```
State: 'order'
  - Status: Pending ⏱ (spinner)
  - Displays: Order ID, Scheme ISIN, Amount, Timestamp
  - "Pay Now" expands payment section:
      - Select: Net Banking | UPI
      - Amount to pay shown
      - "Proceed to Pay" button
  - Optional: "Simulate Payment Failure" (demo only)

On Proceed to Pay:
  - txnResult = 'success' (always in demo)
  - order.status = 'success'
  - Folio number assigned (12 alphanumeric chars)
  - Order saved to localStorage (sure_orders)
  - Moves to state: 'result'
```

### Step 7 — Result Screen
```
State: 'result'
  - SUCCESS: Green banner "Investment Placed Successfully!"
    - Shows: Order ID, ISIN, Amount, Folio Number, Timestamp
    - CTAs: "View All Orders" | "Explore More Funds"

  - FAILURE: Red banner
    - "Retry Payment" returns to state: 'order'
```

### Step 8 — Orders Page
```
/invest/orders
  Tab: Investments
  - Shows all orders (newest first)
  - Per order: Fund name, type (SIP/Lumpsum), amount, status badge, folio
  - SIP orders additionally show: SIP Date, Frequency

  Status badges:
    ✓ Success  (green)
    ✗ Failed   (red)
    ⏱ Pending  (yellow)
```

---

## Redemption Flow

### From Dashboard
```
/invest  (after at least one successful order)
  "Redeem" → /invest/redemption
```

### Redemption Page
```
/invest/redemption
  - Lists all holdings (successful orders, aggregated by fund)
  - For each holding:
      - Fund name, folio number, units, current value
      - Redemption options:
          Amount: Enter ₹ amount
          Units:  Enter number of units
          All:    Redeem entire holding

  - OTP required (same 000000 demo flow)
  - Creates Redemption { id: mfr_[28hex], status: 'processing' }
  - Saved to localStorage (sure_redemptions)
```

### Check Redemptions
```
/invest/orders  → Tab: Redemptions
  - Per redemption: Fund name, type, amount, units, NAV, date, Redemption ID, Folio
  - Status: Processing (yellow) → (in real flow: Success/Failed)
```

---

## localStorage Data Schema (Current Frontend Mock)

| Key | Type | Purpose |
|-----|------|---------|
| `sure_orders` | `Order[]` | All investment orders |
| `sure_redemptions` | `Redemption[]` | All redemption requests |
| `sure_investor_profile` | `InvestorProfile` | KYC + bank + nominee data |

---

## Backend API Mapping (SureMFService → SureInvest)

| Frontend Action | Current (Mock) | SureMFService Endpoint |
|----------------|----------------|----------------------|
| Health check | — | `GET /sure-mf/ping` |
| Display funds | Static `data/funds.ts` | `GET /sure-mf/funds?category=&search=` |
| Fund schemes | Static data | `GET /sure-mf/funds/:id/schemes` |
| KYC check | Local state | `POST /sure-mf/onboarding/kyc-check` |
| Create profile | localStorage | `POST /sure-mf/onboarding/investor-profile` |
| Add phone | localStorage | `POST /sure-mf/onboarding/phone` |
| Add email | localStorage | `POST /sure-mf/onboarding/email` |
| Add address | localStorage | `POST /sure-mf/onboarding/address` |
| Add bank | localStorage | `POST /sure-mf/onboarding/bank` |
| Verify bank (penny drop) | UI only | `POST /sure-mf/onboarding/bank/verify` |
| Add nominee | localStorage | `POST /sure-mf/onboarding/nominee` |
| Activate account | localStorage | `POST /sure-mf/onboarding/activate` |
| Onboarding status | localStorage | `GET /sure-mf/onboarding/status` |
| Place SIP | generateOrderId() | `POST /sure-mf/orders/sip` |
| Place Lumpsum | generateOrderId() | `POST /sure-mf/orders/purchase` |
| Confirm OTP | Check === "000000" | `POST /sure-mf/orders/:id/confirm-otp?type=sip` |
| Get orders | localStorage | `GET /sure-mf/orders` |
| Place redemption | generateRedemptionId() | `POST /sure-mf/orders/redemption` |

### Auth Header (all onboarding + order endpoints)
```
Authorization: Bearer <Firebase ID Token>
```

---

## Firebase Firestore — user_fp_collection

All FP reference IDs stored per user at:
```
user_fp_collection/{firebase_uid}
{
  fp_investor_id:          string   // from POST /v2/investor_profiles
  fp_phone_id:             string   // from POST /v2/phone_numbers
  fp_email_id:             string   // from POST /v2/email_addresses
  fp_address_id:           string   // from POST /v2/addresses
  fp_bank_account_id:      string   // from POST /v2/bank_accounts
  fp_nominee_id:           string   // from POST /v2/related_parties
  fp_investment_account_id: string  // from POST /v2/mf_investment_accounts
  onboarding_step:         int      // 0 → 4
  is_activated:            bool
}
```

---

## PostgreSQL — sure_mf Schema

| Table | Used For |
|-------|---------|
| `pre_verification_usage` | KYC check + penny drop bank verify tracking |
| `otp_activity` | Order OTP initiation and confirmation tracking |
| `email_verification` | Email OTP flow |
| `mf_events` | Audit trail of all order events (created, confirmed, failed) |

---

## Test Credentials (Demo)

| Field | Value |
|-------|-------|
| OTP | `000000` |
| PAN (prefilled) | `BTSPP3751K` |
| Mobile (prefilled) | `9876543210` |
| Email (prefilled) | `test@example.com` |
| Bank Account | pre-filled (penny drop verified) |
