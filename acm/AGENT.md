# AGENT.md — Handover Specification (v5)

**Purpose:** handover note for the next LLM agent or engineer picking up
this codebase mid-stream. Read fully before touching anything.

---

## 0. Working Method (the human is strict about this — keep it)

- **One file at a time.** One `cat > path << 'EOF' ... EOF` heredoc per
  turn, unless explicitly asked to batch.
- **The human applies and tests every change themselves**, then reports
  back before the next file is given.
- **"No build error" is NOT sufficient confirmation for template/CSS/static
  file changes.** A whole batch of template files silently failed to save
  in an earlier session despite the human reporting no errors, because
  Go's build only checks `.go` files. Always verify non-Go file writes
  with an explicit command (`grep -c 'something-unique' path/to/file`)
  and require the actual number back, not just "it worked." For
  visual/CSS changes, ask for a screenshot or specific description of
  what rendered.
- Full-file rewrites (`cat >`) are the default; `cat >>` only for
  genuinely additive changes (e.g. appending CSS to the end of the
  stylesheet).
- No Go toolchain available to the agent in this sandboxed environment —
  the human is the entire build/test loop.
- **Re-verify actual file contents before assuming this document is
  current.** More than once this project has had work done between
  sessions (by the human directly, or by another agent) that this
  document didn't yet reflect — most recently the entire Investment
  model/repo/service/handler layer existed fully built when v4 of this
  doc still described it as not started. Always `view`/`cat` the real
  files — ideally from a fresh zip upload if one's available — before
  planning the next step.

---

## 1. Current State

**Stack:** Go (stdlib `net/http`, `html/template`), SQLite, bcrypt,
in-memory session store with expiry (`internal/auth/session.go`). No
frontend framework — server-rendered templates.

### 1.1 Auth & security — COMPLETE
- `Role string` (`"user"`/`"admin"`) on `models.User`, with `IsAdmin()`
  as a derived method (not a settable field).
- **Hardcoded `admin@acm.com` backdoor is REMOVED.** `RegisterHandler`
  always creates new users with `Role: "user"` — no special-cased email.
- **Promote/demote flow is built and working.** `/admin` (POST) takes
  `user_id` + `role` and calls `UserRepo.UpdateUserRole`. Protected by a
  last-admin check: `UserRepo.CountAdmins()` is consulted before any
  demotion, and demoting the sole remaining admin is rejected with a 400.
  `templates/admin.html` has Promote/Demote buttons per user row, wired
  to this endpoint.
- **Auth middleware** (`internal/middleware/auth.go`) — `RequireAuth`
  and `RequireRole` wrap handlers, resolve the session into a
  `*models.User`, and attach it to the request context. Handlers read it
  via `middleware.UserFromContext(r)` instead of each repeating a
  cookie/session/user lookup. `cmd/server/main.go` wraps every
  authenticated route with one of these two.
- **Session hardening** — `internal/auth/session.go` uses a
  mutex-guarded map with 24h TTL (`auth.SessionTTL`), lazy expiry on
  lookup. Access only via `CreateSession`/`GetSessionEmail`/
  `DeleteSession` — no exported raw map. Login cookie sets
  `HttpOnly: true`, `Secure: true`, `SameSite: Lax`.
  ⚠️ `Secure: true` means the cookie is HTTPS-only — works on the dev
  tunnel URL, but will silently break login on plain `http://localhost`
  with no TLS proxy in front. Don't mistake that for a new bug if it
  comes up.
- **Nav is server-rendered**, not JS-based. `views.RenderTemplate` takes
  `r *http.Request`, resolves login state server-side, and exposes a
  `LoggedIn` template function used in `base.html` via
  `{{if LoggedIn}}...{{else}}...{{end}}`. This is why `RenderTemplate`'s
  signature is `(w, r, file, data)` — every handler's call site reflects
  this.
- `internal/auth/current_user.go`'s `GetCurrentUser` was fixed early on
  (a stale/missing session used to return `(nil, nil)` — no error —
  causing a nil-pointer panic in callers that only checked `err != nil`).
  It's likely unused now that handlers use `middleware.UserFromContext`
  instead, but is still correct if anything calls it.

**All of §1.1 is confirmed working** via direct testing: non-admin denied
`/admin` (403), unauthenticated denied protected routes (redirect to
`/login`), promote/demote tested end-to-end including the last-admin
guard, cookie flags confirmed via DevTools.

### 1.2 Frontend design system — reuse, don't reinvent

**Brand:** "Avalon Capital Miner" — ledger-book aesthetic: numbered rows,
hairline dividers, scrolling ticker tape, mono figures for all numbers.

**Tokens** (`:root` in `static/css/style.css`):
```css
--ink: #0E1420;  --ink-surface: #161D2C;
--parchment: #F3EFE4;  --parchment-dim: #EAE4D4;
--brass: #B8933F;  --brass-dim: #8C7333;
--ledger-green: #3A6B57;  /* positive amounts */
--slate: #8A93A3;  --slate-dark: #4C5567;
--hairline: rgba(184, 147, 63, 0.35);
```
Danger color used ad hoc, not yet tokenized: `#A6483A` (hover `#8A3A2F`)
— used in `.btn--danger`, `.auth__card--danger`, `.action-tile--danger`.

Fonts: `Fraunces` (display/headings), `IBM Plex Sans` (body),
`IBM Plex Mono` (numbers/labels/tickers). Loaded via Google Fonts link in
`base.html`.

**Reusable classes:**
- `.hero` / `.hero--page` — dark intro band (full on homepage, short on
  About/Contact)
- `.ticker` / `.ticker__track` / `.ticker__item` — scrolling tape,
  homepage only, respects `prefers-reduced-motion`
- `.ledger` / `.ledger__row` / `.ledger__index` / `.ledger__body` —
  numbered sequential content
- `.cta-band` — light CTA band, homepage only
- `.auth` / `.auth__card` / `.auth__form` / `.field` / `.auth__switch` —
  dark card form layout: login, register, deposit, withdraw, transfer,
  edit-profile, change-password, admin_plan (plan creation form).
  `.entry__balance` adds a balance readout line above deposit/withdraw/
  transfer forms specifically.
- `.auth__card--danger` / `.auth__eyebrow--danger` / `.btn--danger` —
  destructive variant, `delete_account.html` only
- `.statement` / `.statement__balance` / `.statement__actions` /
  `.action-tile` / `.action-tile--danger` — dashboard/profile/admin/plans
  account-view layout
- `.profile-stats` / `.profile-stats__item` — stat grid, `/profile` and
  `/admin`
- `.ledger-table` — data table: dashboard recent-activity, `/history`,
  `/admin` (users + transactions), `/plans` (plan list), `/investments`
  (portfolio table)
- `.btn` / `.btn--primary` / `.btn--ghost` / `.btn--block` /
  `.btn--danger` / `.btn--small` — button variants. `.btn--small` (new
  this pass) has no color of its own by design, same pattern as
  `.btn--block` — pair it with `.btn--primary` or another color modifier.
- `.ledger-table__inline-form` / `.ledger-table__note` — new this pass,
  `/investments` only: wraps the "Claim" button and the "Awaiting
  maturity" fallback text in the Status column.
- `.role-badge` / `.role-badge--admin` — role pill in the admin users
  table
- `.site-header` / `.site-nav` / `.site-footer` — global chrome,
  `base.html`
- `.nav-anon` / `.nav-authed` — **dead CSS**, no longer referenced by any
  template since the nav went server-rendered (§1.1). Harmless but could
  be deleted next time `style.css` is touched.

**Pages fully restyled and confirmed:** homepage, base.html, login,
register, dashboard, deposit, withdraw, transfer, history, about,
contact, profile, edit-profile, change-password, delete-account, admin,
plans, admin_plan, invest, investments.

**Known minor cosmetic issue, not fixed:** `/admin`'s "User ID" table
header wraps to two lines at normal viewport widths. Low priority.

### 1.3 Investment Plans + Investments feature — COMPLETE end-to-end

Previously (v4 of this doc) this section described a read-only catalog
with no way to actually invest. That is no longer accurate — the full
lifecycle is now built and confirmed working. The ground rule that drove
this work still applies to anything that touches it going forward:
**any feature that lets a user invest money and see a return must have
that return traceable to something real (a disclosed fixed calculation
or actual market data) and logged auditably.** Nothing here fabricates a
number with no source — see the valuation formula below.

**Plan catalog (unchanged from v4, still correct):**
- `models.Plan` — `Name`, `AssetClass`, `Duration` (int days), plus the
  now-structured rate fields described next.
- `repository.PlanRepository` — `CreateTable`, `CreatePlan`,
  `GetPlanByID`, `GetAllPlans`.
- `handlers.PlansHandler` (`GET /plans`, authenticated) — browse-only
  catalog listing.
- `handlers.AdminCreatePlanHandler` (`GET`/`POST /admin/plans`,
  admin-only) — form to create a new plan.
- `templates/plans.html` and `templates/admin_plan.html` — styled,
  confirmed working.

**`RateStructure` has been structured** (this was v4's §1.3 prerequisite,
now done): `Plan.RateType` (currently only `"fixed_compounding"` is
supported/validated) and `Plan.RateValue float64` (total % return over
the full `Duration`, compounded daily to land exactly on that total at
maturity). See the doc comment on `models.Plan` and the formula in
`internal/services/valuation.go` for the exact math.

**Investment model, repo, and valuation engine — built:**
- `models.Investment` — `ID`, `UserID`, `PlanID`, `Principal`,
  `StartedAt`, `EndsAt`, `Status` (`"active"` / `"matured"` /
  `"withdrawn"`).
- `internal/services/valuation.go` — `CalculateInvestmentValue(investment,
  plan, now)` returns `InvestmentValuation{Principal, CurrentValue,
  Profit, Progress, DaysElapsed, DaysRemaining, Matured}`. Pure function,
  deterministic, unit-tested (`valuation_test.go`, 283 lines covering
  start/mid/maturity/edge cases). This is the single source of truth for
  "what is this investment worth right now" — every place that shows or
  pays out a value calls through this, nothing computes its own number.
- `repository.InvestmentRepository` — `CreateTable`,
  `CreateInvestment`, `GetInvestmentByID`, `GetInvestmentsByUserID`,
  `GetAllInvestments`, `UpdateStatus`, plus two transactional methods:
  - `CreateInvestmentWithFunding` — atomically inserts the investment row
    and a `"investment"` transaction row (the debit) in one DB
    transaction.
  - `MatureInvestment` — atomically flips status to `"matured"` and
    inserts an `"investment_return"` transaction row (principal +
    profit, the payout). **Hardened this pass**: the status-update
    `UPDATE ... WHERE id=? AND user_id=? AND status='active'` now checks
    `RowsAffected()` before the payout insert runs — if it matches zero
    rows (wrong owner, wrong ID, already matured/withdrawn), the whole
    transaction rolls back and the function returns `sql.ErrNoRows`
    instead of silently crediting a payout for a status change that
    never happened. This was a real gap in the original version — flag
    it as fixed, don't reintroduce it.

**Invest flow — `handlers.InvestHandler`, `GET`/`POST /invest`:**
Validates plan exists, amount within `[MinDeposit, MaxDeposit]`, amount
`<=` live balance (`TransactionRepo.GetBalanceByUserID`), then computes
`EndsAt = now + plan.Duration days` and calls
`CreateInvestmentWithFunding`. Redirects to `/investments` on success.

**Portfolio view — `handlers.InvestmentsHandler`, `GET /investments`:**
Loads the user's investments, joins each with its plan, runs each
through `CalculateInvestmentValue(investment, plan, time.Now())`, and
renders `templates/investments.html` — a table with Principal, Current
value, Profit, Rate, Progress %, Started/Matures dates, and Status.

**Claim/maturity flow — NEW this pass, `handlers.MatureInvestmentHandler`,
`POST /investments/mature`:**
- Looks up the investment by the submitted `investment_id`, checks
  `investment.UserID == user.ID` (403 if not — the "belongs to someone
  else" case is a clean 403, not a 500 or a silent no-op).
  - Checks `investment.Status == "active"` (409 if not).
  - Checks `!now.Before(investment.EndsAt)` — i.e. maturity date has
    actually passed (400 if claimed early). This is what makes maturity
    a real gate, not just a button that always works.
  - Recomputes the payout via `services.CalculateInvestmentValue` at
    claim time — **the payout amount is never read from the client**,
    only `investment_id` is form input. This is deliberate: it closes
    off any path where a submitted form field could set an arbitrary
    profit number.
  - Calls the hardened `MatureInvestment`; maps its `sql.ErrNoRows` case
    to a 409 ("no longer available to claim" — covers races like a
    double-submit).
- `templates/investments.html` shows a "Claim" button
  (`.btn.btn--primary.btn--small`, POSTs to `/investments/mature`) in the
  Status column only when `.Valuation.Matured` is true for an `"active"`
  investment; otherwise shows "Awaiting maturity" text. The template
  reads `.Valuation.Matured` (computed server-side) rather than
  re-deriving a date comparison in the template itself — one source of
  truth.
- Route registered in `cmd/server/main.go` under `RequireAuth`, right
  after `/investments`.

**End-to-end status:** invest → view portfolio with live valuation →
claim at maturity → balance credited, status flips to `"matured"` is
fully wired. **Not yet done by the human as of this doc**: a real
end-to-end test run (create a short-duration investment, let it pass
`EndsAt`, confirm Claim appears and correctly credits principal+profit).
Worth doing before considering this fully closed out.

**Still not built / explicitly out of scope so far:**
- No `"withdrawn"` path — nothing currently transitions an investment to
  `"withdrawn"` or lets a user exit early. If early-exit/penalty
  withdrawal is wanted, that's a new design conversation (does an early
  exit forfeit accrued profit? partial? penalty on principal?) — same
  "flag, don't silently build" rule applies as it did for the original
  Investment/accrual design.
- No admin-side view of all investments across users (`GetAllInvestments`
  exists in the repo but nothing calls it from a handler yet).

---

## 2. Immediate Next Task

Nothing is committed. Reasonable next steps, roughly in order:

1. **Full end-to-end test pass** on the claim flow (see above) — hasn't
   been done yet as of this doc and should happen before anything else
   builds on top of it.
2. Decide whether an early-withdrawal-from-investment path is wanted; if
   so, scope it as a design conversation first (see forfeiture/penalty
   question above) before writing code.
3. Minor: clean up dead `.nav-anon`/`.nav-authed` CSS (§1.2).
4. Minor: fix `/admin` "User ID" header wrapping (§1.2).
5. Minor: admin-facing view of all investments, using the already-built
   `GetAllInvestments`.

Ask the human which one before starting.

---

## 3. Handover Notes for the Next Agent

- Load §1.2's tokens/classes before writing any new template or CSS.
- Confirm actual file state before trusting this document, especially if
  time has passed since it was written — see the note at the top of §1.
- Keep the one-file-at-a-time, explicit-verification rhythm from §0.
- The ground rule in §1.3 is the most important thing in this document if
  the next task touches money movement or returns. Flag, don't silently
  build, anything that fabricates a return with no real source. The
  claim/maturity flow built this pass is a worked example of following
  it: server-computed payout, no client-supplied amount, atomic DB
  transaction, rows-affected check before crediting money.
