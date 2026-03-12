# SureMFService

Backend service for SureInvest's mutual fund onboarding and investment platform. Built in Go using the Gin framework, it integrates with FinPrim (FP) APIs for investor KYC, account creation, and order execution.

---

## Features

1. **KYC verification** — validate PAN + identity via FP POA API
2. **Investor profile creation** — register investor with FP
3. **Contact & address registration** — phone and email auto-fetched from PostgreSQL, address from request body, all linked to FP profile
4. **Bank account setup** — add bank account + penny-drop verification
5. **Nominee registration** — optional nominee linked to investment account
6. **Account activation** — create MF investment account in FP
7. **Lumpsum purchase** — full 6-step flow (order → consent → payment → confirm → pay → status)
8. **SIP** — systematic investment plans with mandate support, installment tracking, cancel
9. **Redemption** — redeem by amount, units, or full with consent confirmation
10. **Portfolio** — view folios and holdings via FP v2 API, enriched with scheme names
11. **Mandates** — create, authorize, list, status check, cancel eNACH/UPI autopay mandates
12. **Fund browsing** — list MF schemes with filtering, get fund details by ISIN
13. **Event audit trail** — comprehensive mf_events logging across 4 lifecycle phases (created, confirmed, completed, cancelled) with terminal event deduplication
14. **Auto-consent** — consent data (email/phone) auto-fetched from PostgreSQL during confirm flows
15. **Scheme name enrichment** — orders, SIPs, redemptions, and portfolio responses enriched with fund scheme names
16. **CORS middleware** — cross-origin request support for frontend integration
17. **Audit logging** — all requests/responses logged and published to audit queue

---

## Architecture

```
Frontend (Next.js — SureInvest)
    |  Firebase Auth token in Authorization header
    v
SureMFService (Go / Gin)  :9113
    |-- Middleware: CORS + AuditLog + Auth (Firebase token verification)
    |-- Firebase Firestore   <- user profile data + FP ID mappings
    |-- PostgreSQL (sure-app) <- pre-verification tracking, order events
    +-- FinPrim APIs
            |-- Tenant API  (https://s.finprim.com)          <- investor profile, orders
            +-- POA API     (https://api.sandbox.cybrilla.com) <- KYC + penny drop
```

**Auth model**: User-scoped routes (`/:uid/...`) are protected by `AuthMiddleware` which verifies the Firebase Auth token from the `Authorization` header and extracts the UID. Public routes (`/ping`, `/funds`) and callback routes (`/callbacks/*`) do not require auth.

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
PAYMENT_POSTBACK_URL=http://localhost:3000/invest/payment-callback
MANDATE_POSTBACK_URL=http://localhost:3000/invest/mandate-callback

# DB Connection Pool (optional)
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=1h
DB_CONN_MAX_IDLE_TIME=10m
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
- [API Routes Reference](docs/routes.md) — detailed request/response for every endpoint with FP API payloads
- [Integration Learnings](docs/learnings.md) — hard-won debugging insights, gotchas, and FP integration tips
- [Frontend Integration](SureInvest_FLOW.md) — frontend-to-backend action mapping, page routes, real API integration
