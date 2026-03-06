# SureMFService

Backend service for SureInvest's mutual fund onboarding and investment platform. Built in Go using the Gin framework, it integrates with FinPrim (FP) APIs for investor KYC, account creation, and order execution.

---

## Features

1. **KYC verification** — validate PAN + identity via FP POA API
2. **Investor profile creation** — register investor with FP
3. **Contact & address registration** — phone, email, address linked to FP profile
4. **Bank account setup** — add bank account + penny-drop verification
5. **Nominee registration** — optional nominee linked to investment account
6. **Account activation** — create MF investment account in FP
7. **Lumpsum purchase** — full 6-step flow (order → OTP → consent → payment → confirm → status)
8. **SIP** — systematic investment plans with mandate support, installment tracking, cancel
9. **Redemption** — redeem by amount, units, or full with consent confirmation
10. **Portfolio** — view folios and holdings via FP v2 API
11. **Mandates** — create, authorize, list, status check, cancel eNACH/UPI autopay mandates
12. **Event audit trail** — comprehensive mf_events logging across 4 lifecycle phases (created, confirmed, completed, cancelled) with terminal event deduplication
13. **Auto-consent** — consent data (email/phone) auto-fetched from FP during confirm flows

---

## Architecture

```
Frontend (Next.js — SureInvest)
    |  UID in URL path (already authenticated via Firebase)
    v
SureMFService (Go / Gin)  :9113
    |-- Firebase Firestore   <- user profile data + FP ID mappings
    |-- PostgreSQL (sure-app) <- pre-verification tracking, order events
    +-- FinPrim APIs
            |-- Tenant API  (https://s.finprim.com)          <- investor profile, orders
            +-- POA API     (https://api.sandbox.cybrilla.com) <- KYC + penny drop
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

# Firebase Admin SDK (inline -- not a file path)
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

# Payment & Mandate callbacks
PAYMENT_POSTBACK_URL=http://localhost:3000/api/v1/payments/callback
MANDATE_POSTBACK_URL=http://localhost:3000/api/v1/mandates/callback
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

## Documentation

- [API Routes & Flows](docs/flow.md) — all endpoints with request/response examples, end-to-end flows
- [Internals & Data Model](docs/claude.md) — architecture details, database schema, FP API integration, patterns
