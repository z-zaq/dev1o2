# AGENT.md — Handover Specification (v3)

**Purpose:** handover note for the next LLM agent or engineer picking up
this codebase mid-stream. Read fully before touching anything.

---

## 0. Working Method (the human is strict about this — keep it)

- **One file at a time.** One `cat > path << 'EOF' ... EOF` heredoc per
  turn, unless explicitly asked to batch.
- **The human applies and tests every change themselves**, then reports
  back before the next file is given.
- **"No build error" is NOT sufficient confirmation for template/CSS/static
  file changes.** See §1.4 — a whole batch of files silently failed to
  save in a prior session despite the human reporting no errors, because
  Go's build only checks `.go` files, not templates or CSS. From now on,
  verify every non-Go file write with an explicit command
  (`grep -c 'something-unique' path/to/file`) and require the human to
  paste the actual number back, not just say "it worked." For visual/CSS
  changes, ask for a screenshot or a specific description of what
  rendered, not just "no error."
- Full-file rewrites (`cat >`) are the default; `cat >>` only for genuinely
  additive changes (e.g. appending CSS to the end of the stylesheet).
- No Go toolchain available to the agent in this sandboxed environment —
  the human is the entire build/test loop.

---

## 1. Current State

**Stack:** Go (stdlib `net/http`, `html/template`), SQLite, bcrypt,
in-memory session map (`acm/internal/auth.Sessions`). No frontend
framework — server-rendered templates.

### 1.1 Backend changes made and CONFIRMED working
- `internal/models/user.go` — `Role string` (`"user"`/`"admin"`) replaces
  the old `IsAdmin bool` field; `IsAdmin() bool` is now a derived method.
- `internal/repository/user.go` — `CreateTable()` migrates existing DBs
  (adds `role` column if missing, backfills from old `is_admin`). CRUD
  methods updated accordingly.
- `internal/handlers/auth.go` — sets `Role: "user"` by default,
  `"admin"` only for the hardcoded `admin@acm.com` email (**this backdoor
  still exists** — see §3).
- `internal/handlers/admin.go` — `/admin` now actually requires a valid
  session AND `user.IsAdmin()`. Previously had zero auth check.
- `internal/handlers/dashboard.go` — passes `Recent` (last 5 transactions)
  alongside `User`/`Balance`.
- `internal/handlers/deposit.go`, `withdraw.go`, `transfer.go` — GET
  branches pass `Balance` for display context.
- `internal/auth/current_user.go` — **bug fix.** Previously, a session
  cookie with no matching entry in `Sessions` (e.g. after a server
  restart, which wipes the in-memory map) caused `GetCurrentUser` to
  return `(nil, nil)` — a nil user with no error. Callers that only check
  `err != nil` (like `ProfileHandler`) would then panic dereferencing the
  nil pointer, causing a full server crash on that request (symptom:
  blank browser page + `curl` reporting "empty reply from server"). Fixed
  to return a proper error when the session isn't found. **Confirmed
  fixed and tested** by the human.

### 1.2 Frontend design system (established, confirmed working, reuse it)

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
Danger/destructive color used ad hoc (not yet tokenized): `#A6483A`
(hover `#8A3A2F`) — used in `.btn--danger`, `.auth__card--danger`,
`.action-tile--danger`. **Consider promoting this to a `--danger` CSS
variable next time style.css is touched, for consistency.**

Fonts: `Fraunces` (display/headings), `IBM Plex Sans` (body),
`IBM Plex Mono` (all numbers/labels/tickers). Loaded via Google Fonts link
in `base.html`.

**Reusable classes** — do not reinvent, reuse:
- `.hero` / `.hero--page` — dark intro band (full on homepage, short
  variant on About/Contact)
- `.ticker` / `.ticker__track` / `.ticker__item` — scrolling tape,
  homepage only, respects `prefers-reduced-motion`
- `.ledger` / `.ledger__row` / `.ledger__index` / `.ledger__body` —
  numbered sequential content
- `.cta-band` — light CTA band, homepage only
- `.auth` / `.auth__card` / `.auth__form` / `.field` / `.auth__switch` —
  dark card form layout: login, register, deposit, withdraw, transfer,
  edit-profile, change-password. `.entry__balance` adds a balance readout
  line above the form (deposit/withdraw/transfer only).
- `.auth__card--danger` / `.auth__eyebrow--danger` / `.btn--danger` —
  destructive variant, used only on `delete_account.html`
- `.statement` / `.statement__balance` / `.statement__actions` /
  `.action-tile` / `.action-tile--danger` — dashboard/profile account-view
  layout
- `.profile-stats` / `.profile-stats__item` — 3-stat grid on `/profile`
- `.ledger-table` — data table, dashboard recent-activity + `/history`
  (green/red via `.ledger-table__amount--{{.Type}}`)
- `.site-header` / `.site-nav` / `.site-footer` — global chrome in
  `base.html`
- `.nav-anon` / `.nav-authed` — auth-state visibility toggle, see §1.3

**Pages fully restyled and visually confirmed by the human:** homepage,
`base.html`, login, register, dashboard, deposit, withdraw, transfer,
history, about, contact, profile, edit-profile, change-password,
delete-account.

**Pages NOT yet restyled:** `admin.html` only — still plain HTML tables.
Auth is fixed (§1.1) but the visual pass hasn't happened.

### 1.3 Nav auth-state toggle — STOPGAP, read this before touching auth
`base.html` now shows/hides "Log in"/"Open an account" vs.
"Dashboard"/"Profile"/"Log out" based on a **client-side JS check** for
the presence of the `session` cookie:
```js
document.documentElement.classList.add(
    document.cookie.indexOf('session=') !== -1 ? 'is-authed' : 'is-anon'
);
```
CSS classes `.nav-anon`/`.nav-authed` (hidden by default, shown via
`html.is-anon`/`html.is-authed`) do the actual toggling.

**This only works because the session cookie currently has no `HttpOnly`
flag.** The moment session hardening happens (§3 item 3 — adding
`HttpOnly`/`Secure`/`SameSite`), this script can no longer read the
cookie and the nav will silently revert to always showing the logged-out
state. **When you do session hardening, you must also convert this to a
server-rendered check** — easiest path is adding a `CurrentUser` (or just
`LoggedIn bool`) field that every handler passes into its template data,
or wrapping `views.RenderTemplate` to inject it automatically. Don't
forget this — it's an easy thing to miss since the hardening ticket
doesn't mention the nav at all.

**⚠️ NOT YET VISUALLY CONFIRMED.** The human applied this change and the
`grep` counts matched, but they moved on before confirming in the browser
that the nav actually flips correctly between logged-in and logged-out
states. **Verify this first**, in both states, before assuming it works or
building anything else on top of it.

### 1.4 Known failure mode — verify file writes explicitly
Earlier in this session, five template files (`profile.html`,
`edit_profile.html`, `change_password.html`, `delete_account.html`, and
associated CSS) were reported by the human as "no error" / working, but a
fresh zip upload later showed they had never actually been written — the
old unstyled versions were still in place. Root cause wasn't fully
diagnosed (possibly a heredoc that silently failed, wrong working
directory at that moment, or similar). No proof this can't happen again.
**Mitigation going forward:** always ask for a `grep -c` (or similar)
confirmation of new, distinctive content after every file write, not just
"did the build succeed."

---

## 2. Immediate Next Task

Nothing is queued/committed — the human said "not sure yet" for what's
next. Reasonable candidates, in rough order of what unblocks the most:

1. **Verify the nav toggle (§1.3) actually renders correctly** — do this
   first regardless of what else is picked, since it's unconfirmed.
2. **Style `/admin`** — last unstyled page. Data shape: `.Users` (slice of
   `{ID, Name, Email, Role}`), `.Transactions` (slice of full
   `Transaction` structs). Reuse `.ledger-table` for both.
3. **Auth middleware refactor** — every handler repeats the same
   cookie→session→user lookup block. Extract to
   `internal/middleware/auth.go` (`RequireAuth`/`RequireRole`). Touches
   every handler file — do one at a time.
4. **Session hardening** (expiry + `HttpOnly`/`Secure`/`SameSite`) — see
   §1.3 for the nav dependency this creates.
5. **Remove the `admin@acm.com` hardcoded backdoor** (§3).

Ask the human which one before starting, rather than assuming.

---

## 3. Outstanding Items (carried forward, still valid)

1. Hardcoded `admin@acm.com` admin backdoor — still the only way to get
   admin access.
2. Auth middleware refactor (see §2.3).
3. Session hardening — unbounded in-memory session map, no expiry, no
   `HttpOnly`/`Secure`/`SameSite` on the cookie. **Now also blocks on
   updating the nav toggle first (§1.3)** if this is picked up.
4. Investment plans/portfolios, loans, market data, compliance
   scaffolding, infra — see AGENT.md v1 Phases 2–7, none started. Ground
   rule still applies: any feature promising a return on deposited funds
   must be traceable to something real and logged auditably — flag rather
   than silently build anything that fabricates "profit."

---

## 4. Handover Notes for the Next Agent

- Load §1.2's tokens/classes before writing any new template or CSS.
- Re-verify actual file contents before editing — don't trust this
  document's description of "done" over what's actually in the repo; §1.4
  is exactly why.
- Keep the one-file-at-a-time, explicit-verification rhythm from §0.
- Flag, don't silently build, any feature that fabricates financial
  returns with no real backing (§3 item 4).
