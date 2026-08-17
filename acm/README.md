# ACM → Investment Platform: Build Roadmap

This document maps what's already implemented in the `acm` codebase, what's
still needed to turn it into a full investment/portfolio platform, and which
frameworks/libraries fit each piece. It's meant to be split up — each section
under **Features To Add** is scoped so one engineer can pick it up
independently.

---

## 1. What's Already Built

| Area | Status | Notes |
|---|---|---|
| User registration/login | ✅ Done | bcrypt password hashing, basic validators |
| Sessions | ✅ Done (needs hardening) | In-memory map, cookie-based, no expiry |
| Deposit / Withdraw | ✅ Done | Balance derived from transaction ledger |
| Transfer between users | ✅ Done | |
| Transaction history | ✅ Done | |
| Profile edit / password change / delete account | ✅ Done | |
| Admin view (users + transactions) | ⚠️ Done but unauthenticated | Needs access control — see §2.1 |
| Storage | ✅ SQLite | Fine for dev; revisit before production scale |

**Stack currently in use:** Go standard library (`net/http`, `html/template`),
`golang.org/x/crypto/bcrypt`, SQLite.

---

## 2. Features To Add

### 2.1 Authentication & Access Control (Priority: High)
- [ ] **Role-based access control (RBAC)** — replace the hardcoded
  `admin@acm.com` check with a proper `role` column (`user`, `admin`,
  `support`) and middleware that gates routes by role.
- [ ] **Session hardening** — move sessions out of an in-memory map into
  Redis or the DB so they survive restarts and can expire. Set
  `HttpOnly`, `Secure`, `SameSite=Strict` on the cookie.
- [ ] **Auth middleware** — a single `RequireAuth` / `RequireAdmin` wrapper
  instead of repeating the cookie-lookup block in every handler.
- [ ] **2FA (TOTP)** — recommended before handling real money movement.
  Library: `github.com/pquerna/otp`.
- [ ] **Rate limiting on login/register** — prevent credential stuffing.
  Library: `golang.org/x/time/rate` or a middleware like `tollbooth`.

### 2.2 Investment Plans / Portfolios (Priority: High)
This is the core feature gap between "bank clone" and "investment platform."

- [ ] **`Plan` model** — name, asset class (crypto/forex/stocks/ETF/etc.),
  duration, rate structure, min/max deposit.
- [ ] **`Investment` model** — links `UserID` + `PlanID`, tracks principal,
  start date, end date, status (`active`, `matured`, `withdrawn`).
- [ ] **Valuation/accrual engine** — a scheduled job (cron or Go
  `time.Ticker`) that calculates accrued value per active investment and
  writes it to the ledger.
- [ ] **Maturity handling** — auto-flag investments as maturable and surface
  a "withdraw principal + return" flow once the term ends.
- [ ] **Portfolio bundling** — allow a single investment to span multiple
  plans/assets with weighted allocation.

> Engineering note: however this is priced, the return calculation should be
> transparent and traceable to something real (market data, a stated fixed
> rate disclosed up front, etc.) — not just an arbitrary number written to
> the DB. That distinction matters both technically (auditability) and
> legally.

### 2.3 Loans Against Deposits (Priority: Medium)
- [ ] **`Loan` model** — principal, interest rate, term, repayment schedule,
  status (`pending`, `approved`, `active`, `paid`, `defaulted`).
- [ ] **Eligibility rules engine** — configurable thresholds (e.g. minimum
  account balance/tenure) rather than hardcoded numbers.
- [ ] **Approval workflow** — admin or automated approval step before funds
  are released.
- [ ] **Repayment scheduler** — monthly interest job similar to the accrual
  engine in §2.2.

### 2.4 Market Data Integration (Priority: Medium)
- [ ] **Live price feeds** for crypto/forex/stocks so displayed portfolio
  values reflect real market movement instead of static numbers.
  - Crypto: `CoinGecko` or `CoinMarketCap` API
  - Stocks/ETFs: `Alpha Vantage`, `Polygon.io`, or `IEX Cloud`
  - Forex: `exchangerate.host` or `OANDA`
- [ ] **Caching layer** for price data — Redis with a short TTL to avoid
  rate-limit issues.
- [ ] **Background price-sync worker** — separate goroutine/service that
  pulls prices on an interval and stores them for the accrual engine to use.

### 2.5 Payments / Deposits & Withdrawals (Priority: High)
- [ ] **Real payment rails** — current deposit/withdraw just writes a DB row;
  a production system needs an actual money-movement integration:
  - Fiat: Stripe, Plaid (for bank linking), or a banking-as-a-service
    provider (Unit, Synapse)
  - Crypto: a wallet/custody provider (Fireblocks, BitGo) or direct
    on-chain wallet integration (`go-ethereum`, `btcd` libraries)
- [ ] **Withdrawal approval queue** — large withdrawals should route through
  a review step, not auto-approve.
- [ ] **Idempotency keys** on all money-movement endpoints to prevent
  double-submission.

### 2.6 Compliance & Risk (Priority: High before any real launch)
- [ ] **KYC/AML identity verification** — Persona, Sumsub, or Onfido
  integration at signup/first-deposit.
- [ ] **Audit logging** — immutable log of every balance-affecting action
  (who, what, when, before/after values).
- [ ] **Terms of service / risk disclosure acceptance** tracked per user with
  timestamp and version.
- [ ] **Regulatory review** — depending on jurisdiction, pooling public
  deposits into "investment plans" with promised returns may require
  licensing (money transmitter, investment adviser, securities
  registration). This is a legal question, not just an engineering one —
  loop in counsel before this goes live with real funds.

### 2.7 Frontend / UX (Priority: Medium)
- [ ] Replace server-rendered `html/template` pages with a proper SPA if a
  richer dashboard (charts, live price tickers) is wanted.
  - React + TypeScript, or Vue, with `recharts`/`chart.js` for portfolio
    performance graphs.
- [ ] **Notifications** — email (SendGrid/Postmark) or in-app for deposit
  confirmations, maturity alerts, loan approvals.
- [ ] **Referral system**, if wanted — referral code on signup, tracked
  payouts to referrer on referee deposits.

### 2.8 Infrastructure (Priority: Medium)
- [ ] **Move off SQLite** to Postgres for concurrent write safety and better
  tooling (`pgx` or `gorm` as the driver/ORM).
- [ ] **Migrations** — `golang-migrate` or `goose` instead of ad hoc
  `CreateTable()` calls.
- [ ] **Containerization** — Dockerfile + docker-compose for local dev
  parity (app + Postgres + Redis).
- [ ] **CI** — GitHub Actions running `go vet`, `go test`, `golangci-lint`
  on every PR.
- [ ] **Structured logging & monitoring** — swap `log.Println` for
  `zerolog`/`zap`, add Prometheus metrics + Grafana dashboards for
  transaction volume, error rates, login attempts.
- [ ] **Environment config** — move the hardcoded `:8080` port, DB path,
  etc. into env vars (`github.com/joho/godotenv` for local dev).

---

## 3. Suggested Repo Structure Additions

```
acm/
  internal/
    models/
      plan.go            # new
      investment.go      # new
      loan.go            # new
    repository/
      plan.go            # new
      investment.go       # new
      loan.go            # new
    handlers/
      plans.go            # new
      invest.go           # new
      loans.go            # new
      admin_auth.go        # new — RBAC middleware
    services/
      accrual/             # new — scheduled valuation engine
      pricing/              # new — market data client + cache
      payments/              # new — payment provider integrations
    middleware/
      auth.go               # new — RequireAuth/RequireAdmin
      ratelimit.go            # new
  migrations/                 # new — goose/golang-migrate files
  docker-compose.yml           # new
  Dockerfile                    # new
```

---

## 4. How to Pick a Task

Each checklist item above is scoped to be workable independently as long as
the underlying `models` for that area exist first. Suggested order for a
small team:

1. **RBAC + auth middleware** (§2.1) — unblocks admin work safely.
2. **Plan/Investment models + basic CRUD** (§2.2) — the core new domain
   object everything else hangs off of.
3. **Loans** (§2.3) and **market data** (§2.4) can be built in parallel once
   §2.2 lands, since both consume the `Investment` model.
4. **Payments** (§2.5) and **Compliance** (§2.6) should be scoped with
   legal/product before writing code — these carry real regulatory weight.
5. **Frontend, infra, and notifications** (§2.7–2.8) can proceed in parallel
   with the above at any point.
