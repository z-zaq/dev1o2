# AGENT.md — Handover Specification (v4)

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
  document didn't yet reflect. Always `view`/`cat` the real files —
  ideally from a fresh zip upload if one's available — before planning
  the next step.

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
  edit-profile, change-password, **admin_plan (plan creation form)**.
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
  `/admin` (users + transactions), `/plans` (plan list)
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
plans, admin_plan.

**Known minor cosmetic issue, not fixed:** `/admin`'s "User ID" table
header wraps to two lines at normal viewport widths. Low priority.

### 1.3 Investment Plans feature — STARTED, catalog only, NOT investable yet
New since the last full handover. Read the ground rule below before
extending this.

**What exists:**
- `models.Plan` — `Name`, `AssetClass`, `Duration` (int days),
  `RateStructure` (free-text string, e.g. "Fixed" or a raw number — **not
  yet structured/parseable**, see caveat below), `MinDeposit`,
  `MaxDeposit`.
- `repository.PlanRepository` — `CreateTable`, `CreatePlan`,
  `GetPlanByID`, `GetAllPlans`.
- `handlers.PlansHandler` (`GET /plans`, authenticated) — lists all plans
  for any logged-in user to browse.
- `handlers.AdminCreatePlanHandler` (`GET`/`POST /admin/plans`,
  admin-only) — form to create a new plan.
- `templates/plans.html` and `templates/admin_plan.html` — both styled,
  confirmed working end-to-end (plan created via the form shows up
  correctly in the listing).

**What does NOT exist yet — this is the important part:**
- No `Investment` model. A user cannot actually put money into a plan.
- No accrual/valuation engine. Nothing computes a "current value" for
  anything.
- No maturity/withdrawal-from-plan flow.

**This is exactly right for where it is.** `/plans` is a read-only
catalog right now — nobody's money is at risk, nothing fabricates a
return. The ground rule from earlier in this project still applies and
matters more from here on: **any feature that lets a user invest money
and see a return must have that return traceable to something real
(a disclosed fixed calculation or actual market data) and logged
auditably.** Do not implement an "Investment" model or accrual engine
that silently generates a "profit" number with no underlying source —
flag it and discuss instead of building it quietly.

**Before building the Investment/accrual layer**, `RateStructure` needs
to become a real, parseable value (e.g. split into a `RateType` enum —
`"fixed"` — and a `RateValue float64`) instead of freeform text like
`"20"` or `"Fixed"`. A human typing an arbitrary string into that field
right now has no defined meaning to any future code that would read it.

---

## 2. Immediate Next Task

Nothing is committed. The security/hardening punch list (§1.1) and the
full visual pass (§1.2) are both complete. Reasonable next steps, roughly
in order of what's needed before the next one:

1. **Structure `RateStructure`** (§1.3) into something parseable, if the
   Investment/accrual feature is going to be built next.
2. **Design and scope the `Investment` model + accrual engine** (§1.3) —
   this is a real design conversation, not just a coding task, because of
   the ground rule about traceable returns. Decide fixed-rate vs.
   market-linked before writing any code.
3. Minor: clean up dead `.nav-anon`/`.nav-authed` CSS (§1.2).
4. Minor: fix `/admin` "User ID" header wrapping (§1.2).

Ask the human which one before starting.

---

## 3. Handover Notes for the Next Agent

- Load §1.2's tokens/classes before writing any new template or CSS.
- Confirm actual file state before trusting this document, especially if
  time has passed since it was written — see the note at the top of §1.
- Keep the one-file-at-a-time, explicit-verification rhythm from §0.
- The ground rule in §1.3 is the most important thing in this document if
  the next task touches money movement or returns. Flag, don't silently
  build, anything that fabricates a return with no real source.
