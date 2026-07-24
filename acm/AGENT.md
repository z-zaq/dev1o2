# AGENT.md — Build Handover Specification

**Purpose of this document:** This is a handover note for the next engineer
or LLM agent picking up this codebase. It assumes zero prior context beyond
"here's a Go repo, here's what exists, here's what's next." Each phase below
is self-contained: **feature analysis → implementation → test/verify →
definition of done**, so whoever picks it up can confirm their work is
correct before moving to the next phase.

**Non-negotiable ground rule for whoever builds this:** any feature that
promises users a return on deposited funds (investment plans, portfolios,
loans against balance) must have its numbers traceable to something real —
a disclosed fixed rate calculated transparently, or actual market data —
and logged in an auditable way. Do not implement a feature that silently
fabricates "profit" with no underlying source; that's the line between a
legitimate fintech product and a Ponzi mechanism, and it's a legal problem
as much as an engineering one. If a feature spec below seems to ask for
that, stop and flag it rather than building it.

---

## 0. Current State (read this first)

**Stack:** Go (stdlib `net/http`, `html/template`), SQLite, bcrypt for
passwords, in-memory session map.

**Repo layout:**
```
acm/
  cmd/server/main.go
  internal/
    auth/         # session.go, current_user.go
    database/     # sqlite.go
    handlers/      # one file per route
    models/         # user.go, transaction.go
    repository/      # user.go, transaction.go
    validators/       # user.go
    views/             # render.go
  templates/            # html/template files
  static/css/
```

**Working today:** registration, login/logout, deposit, withdraw, transfer,
transaction history, profile edit, password change, account deletion, admin
page listing all users/transactions.

**Known issues to fix as part of Phase 1, not ignore:**
- `/admin` route has no auth check — anyone can load it.
- Sessions live in an unbounded in-memory map with no expiry.
- Session cookie has no `Secure`/`HttpOnly`/`SameSite` flags.
- Admin role is granted by hardcoded email match (`admin@acm.com`).

Run it locally to confirm baseline before touching anything:
```bash
cd acm
go run cmd/server/main.go
# visit http://localhost:8080, register a user, deposit/withdraw, confirm dashboard updates
```
**Verify:** balance shown on `/dashboard` matches sum of transactions on `/history`.

---

## Phase 1 — Auth Hardening & RBAC

### Feature analysis
Current admin/user distinction is a hardcoded email check with no route
protection. Need a real role system and reusable middleware before building
anything else on top of it, since every later phase (plans, loans, admin
approvals) depends on knowing who's allowed to do what.

### Implementation
1. Add `Role string` (`"user"` / `"admin"` / `"support"`) to `models.User`,
   replacing `IsAdmin bool`. Migrate existing rows.
2. Write `internal/middleware/auth.go`:
   ```go
   func RequireAuth(next http.HandlerFunc) http.HandlerFunc
   func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc
   ```
   Both should read the session cookie, resolve the user, and either call
   `next` or redirect/403.
3. Wrap `/admin` and any future admin-only routes in `RequireRole("admin", ...)`
   in `main.go`.
4. Replace the in-memory `auth.Sessions` map with a store that supports
   expiry (Redis, or a `sessions` table with `expires_at`). Sweep expired
   sessions on a timer or on read.
5. Set `HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode` on
   the session cookie.

### Test / verify
- Unit test: `RequireRole("admin", ...)` returns 403 for a non-admin
  session, 200 for an admin session, 302→/login for no session.
- Manual: log in as a normal user, hit `/admin` directly via curl with that
  session cookie — expect redirect/403, not the user list.
- Manual: log in, wait past the session TTL (or manually expire it in the
  store), hit `/dashboard` — expect redirect to `/login`.
- `curl -i http://localhost:8080/login -c cookies.txt -d "email=...&password=..."`
  then inspect the `Set-Cookie` header for `HttpOnly`, `Secure`, `SameSite`.

### Definition of done
- [ ] No route reachable by role escalation via direct URL hit.
- [ ] Sessions expire and are evictable.
- [ ] All existing handlers still pass manual smoke test (register → login →
  deposit → withdraw → transfer → logout).

---

## Phase 2 — Investment Plan / Portfolio Domain

### Feature analysis
This is the core new domain object. A `Plan` describes an offering (asset
class, duration, rate structure). An `Investment` is a user's position in a
plan — principal, start/end dates, current value. This phase is pure data
model + CRUD; the accrual/valuation logic comes in Phase 3.

### Implementation
1. `internal/models/plan.go`:
   ```go
   type Plan struct {
       ID          int
       Name        string
       AssetClass  string  // "crypto", "forex", "stocks", "etf", "index"
       DurationDays int
       RateType    string  // "fixed" or "market_linked"
       RateValue   float64 // annualized rate if fixed; unused if market_linked
       MinDeposit  float64
       MaxDeposit  float64
   }
   ```
2. `internal/models/investment.go`:
   ```go
   type Investment struct {
       ID          int
       UserID      int
       PlanID      int
       Principal   float64
       StartDate   time.Time
       EndDate     time.Time
       Status      string // "active", "matured", "withdrawn"
       CurrentValue float64
   }
   ```
3. `internal/repository/plan.go` and `investment.go` — standard CRUD
   following the pattern already established in `repository/transaction.go`.
4. `internal/handlers/plans.go`:
   - `GET /plans` — list available plans
   - `POST /plans/invest` — create an `Investment` for the logged-in user,
     debit their balance via a `Transaction` of type `"investment"`
   - `GET /investments` — list the user's active/past investments
5. Admin CRUD for plans (`RequireRole("admin", ...)`-wrapped) so plans can
   be created without touching the DB directly.

### Test / verify
- Unit test: creating an investment with `amount < Plan.MinDeposit` or
  `> MaxDeposit` is rejected.
- Unit test: creating an investment with `amount > user balance` is
  rejected (mirror the existing withdraw balance check).
- Manual: as admin, create a plan via `/admin/plans`. As a user, invest in
  it, confirm balance decreases by the principal and `/investments` shows
  the new row with `status = "active"`.
- Confirm a `Transaction` row was written for the investment (auditability).

### Definition of done
- [ ] Plan CRUD works end-to-end from admin UI.
- [ ] Investing debits balance correctly and is rejected outside min/max.
- [ ] Every investment action has a corresponding transaction log entry.

---

## Phase 3 — Valuation / Accrual Engine

### Feature analysis
This is where an investment's `CurrentValue` actually updates over time.
Per the ground rule above: `fixed` rate plans compute a disclosed,
deterministic accrual; `market_linked` plans require Phase 4 (market data)
to be in place first.

### Implementation
1. `internal/services/accrual/engine.go` — a function
   `RunAccrual(db)` that, for every `active` investment:
   - If `Plan.RateType == "fixed"`: compute pro-rated value based on
     `RateValue`, `DurationDays`, and days elapsed. Formula must be
     documented in code comments — no magic numbers.
   - If `Plan.RateType == "market_linked"`: pull latest price data (Phase 4)
     for the plan's underlying asset(s) and compute value based on actual
     price movement since `StartDate`.
   - Write the updated `CurrentValue` and log a `Transaction` (type
     `"accrual"`) reflecting the change, so the change is auditable, not
     silent.
2. Run this on a schedule: a `time.Ticker` goroutine started in `main.go`
   (daily is reasonable for a first pass), or a separate cron-triggered
   binary if you want it decoupled from the web server.
3. On `EndDate`, flip `Status` to `"matured"` and surface it in `/investments`
   as ready to withdraw.
4. `POST /investments/{id}/withdraw` — moves `CurrentValue` back into the
   user's withdrawable balance via a `Transaction` (type `"payout"`), only
   allowed when `Status == "matured"` (or earlier with an early-withdrawal
   penalty, if that's a feature you want — spec it explicitly if so).

### Test / verify
- Unit test: fixed-rate accrual math against hand-calculated expected
  values for a few day-offsets (day 0, halfway, full term).
- Unit test: an investment past `EndDate` flips to `"matured"` on the next
  accrual run.
- Manual: create a short-duration plan (e.g. `DurationDays: 1` for testing),
  invest, manually trigger `RunAccrual`, confirm `CurrentValue` and the
  transaction log update as expected.
- Manual: attempt withdrawal before maturity — expect rejection; after
  maturity — expect success and balance credit.

### Definition of done
- [ ] Accrual math is documented, deterministic, and unit-tested.
- [ ] Every value change has a matching transaction log entry.
- [ ] Maturity + withdrawal flow works end-to-end.

---

## Phase 4 — Market Data Integration (for `market_linked` plans)

### Feature analysis
Needed only if `market_linked` plans are in scope. Provides real price data
so accrual in Phase 3 reflects actual market movement rather than an
arbitrary number.

### Implementation
1. `internal/services/pricing/client.go` — thin client for one provider per
   asset class to start:
   - Crypto: CoinGecko public API (no auth needed for basic price lookups)
   - Stocks/ETFs: Alpha Vantage or IEX Cloud (requires API key — put it in
     env config)
2. Cache responses (Redis or in-memory with TTL) — respect provider rate
   limits.
3. Background worker (`time.Ticker`, separate from the accrual engine) that
   refreshes prices for every asset referenced by an active `market_linked`
   plan.
4. Accrual engine (Phase 3) reads from this cache, never calls the provider
   directly inline with a user-facing request.

### Test / verify
- Unit test: pricing client parses a mocked provider response correctly
  (use `httptest.Server` to stub the API, don't hit the real API in CI).
- Manual: confirm cached price updates on the expected interval by logging
  timestamps.
- Manual: confirm accrual engine output changes when the cached price
  changes (mock the cache value, re-run accrual, check `CurrentValue`).

### Definition of done
- [ ] No user-facing request makes a live call to an external pricing API.
- [ ] Rate limits respected (log/alert if a provider call fails/throttles).
- [ ] Accrual values move correctly with price changes.

---

## Phase 5 — Loans Against Deposits

### Feature analysis
Optional feature: users with sufficient balance/investment can request a
loan repaid over time. Needs an approval workflow — do not auto-disburse.

### Implementation
1. `internal/models/loan.go`:
   ```go
   type Loan struct {
       ID              int
       UserID          int
       Principal       float64
       InterestRate    float64
       TermMonths      int
       Status          string // "pending", "approved", "active", "paid", "defaulted", "rejected"
       RequestedAt     time.Time
       ApprovedAt      *time.Time
   }
   ```
2. Eligibility check as a pure function, not hardcoded inline in the
   handler: `IsEligible(user, balance) (bool, reason string)`, using
   configurable thresholds (env/config, not magic numbers in code).
3. `POST /loans/request` — creates a `"pending"` loan, no funds move yet.
4. Admin-only `POST /loans/{id}/approve` or `/reject` — only on approval
   does a `Transaction` credit the user's balance.
5. Repayment scheduler (same pattern as the accrual engine) that runs
   monthly interest/principal deductions for `"active"` loans.

### Test / verify
- Unit test: `IsEligible` correctly gates on the configured threshold.
- Manual: request a loan below threshold — rejected client-side with clear
  reason. Request above threshold — lands in `"pending"`, no balance change.
- Manual: admin approves — balance credited, transaction logged, status
  flips to `"active"`.
- Manual: run repayment scheduler manually, confirm balance debited per
  schedule and loan flips to `"paid"` after final installment.

### Definition of done
- [ ] No loan funds move without an explicit approval step.
- [ ] Every disbursement/repayment has a transaction log entry.
- [ ] Eligibility thresholds are configurable, not hardcoded.

---

## Phase 6 — Compliance Scaffolding

### Feature analysis
Before any of the above touches real money, these need to exist. Treat this
phase as a prerequisite gate for production, not a nice-to-have.

### Implementation
1. **Audit log** — append-only table (`audit_log`: actor, action, target,
   before/after values, timestamp) written on every balance-affecting
   operation. Do this as a shared helper called from every handler that
   touches `Transaction`, not bolted on per-feature.
2. **ToS/risk disclosure acceptance** — `terms_accepted_at`, `terms_version`
   on `User`, enforced at signup and re-enforced on version bump.
3. **KYC hook point** — even if not integrating a real provider yet, add a
   `kyc_status` field (`"unverified"`, `"pending"`, `"verified"`) and gate
   deposits/withdrawals above a configurable threshold on `"verified"`, so
   the integration point exists when a provider (Persona/Sumsub/Onfido) is
   wired in.
4. **Legal review flag** — this phase is a placeholder for a non-engineering
   task: before enabling real fund movement, get the plan/loan mechanics
   reviewed against money-transmitter/securities regulations in your
   operating jurisdiction. Note this explicitly in the PR description so
   it isn't silently skipped.

### Test / verify
- Unit test: audit log entry created for deposit, withdraw, transfer,
  invest, loan approval — one test per action type.
- Manual: attempt a withdrawal above the KYC threshold on an unverified
  account — expect rejection with a clear message.

### Definition of done
- [ ] Every money-movement action is represented in the audit log.
- [ ] KYC gate demonstrably blocks large transactions pre-verification.
- [ ] Legal review noted as outstanding/complete in project tracker.

---

## Phase 7 — Infra & Ops

### Feature analysis
Not user-facing, but required before this can run reliably outside a
developer's laptop.

### Implementation
1. Migrate SQLite → Postgres (`pgx` driver). Write `golang-migrate` or
   `goose` migration files for every table created above.
2. Dockerfile + `docker-compose.yml` (app + Postgres + Redis).
3. GitHub Actions: `go vet`, `go test ./...`, `golangci-lint run` on every
   PR.
4. Structured logging (`zerolog` or `zap`) replacing `log.Println`.
5. Metrics (Prometheus client) on: login attempts (success/fail), deposit/
   withdrawal volume, active investments, loan approval rate.

### Test / verify
- `docker-compose up` brings up app + DB + Redis with one command, app
  reachable on the configured port.
- CI pipeline fails on an intentionally broken test to confirm it's wired
  correctly, then passes once fixed.
- Metrics endpoint (`/metrics`) returns Prometheus-formatted output.

### Definition of done
- [ ] Fresh clone + `docker-compose up` gets a working stack with no manual
  steps beyond `.env` setup.
- [ ] CI blocks merge on failing tests/lint.
- [ ] Core metrics visible and dashboarded (Grafana or similar).

---

## Handover Notes for the Next Agent

- Work phases **in order** — Phase 2 depends on Phase 1's auth middleware,
  Phase 3 depends on Phase 2's models, Phase 4 is only needed if
  `market_linked` plans are actually in scope.
- Every phase above ends with a **Definition of Done** checklist — don't
  mark a phase complete in the tracker until every box is checked and the
  manual test steps have been run and observed, not just assumed to work.
- If you hit a feature request that would make return numbers not traceable
  to a real, disclosed calculation (see ground rule at the top), stop and
  raise it rather than implementing it silently.
- Existing code style: no framework, stdlib `net/http`, one handler per
  file, repository pattern for DB access. Keep new code consistent with
  that rather than introducing a new pattern mid-project.
