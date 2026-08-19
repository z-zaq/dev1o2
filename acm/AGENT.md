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
- `.nav-anon` / `.nav-authed` — legacy classes, no longer used by
  `base.html` as of §1.3, may still exist as dead CSS

**Pages fully restyled and visually confirmed by the human:** homepage,
`base.html`, login, register, dashboard, deposit, withdraw, transfer,
history, about, contact, profile, edit-profile, change-password,
delete-account, admin.

**All pages are now styled.** The visual pass described in this document
is complete — every page uses the design system in §1.2. Admin page uses
`.profile-stats` for the summary counts and `.ledger-table` for both the
users and transactions lists, plus a new `.role-badge`/`.role-badge--admin`
class for the role column.

**Known minor cosmetic issue (not fixed):** on `/admin`, the "User ID"
table header wraps to two lines and crowds the adjacent "ID" column at
normal viewport widths. Low priority, flagged but not addressed.

### 1.3 Nav auth-state — NOW SERVER-RENDERED (no longer a stopgap)
The nav previously used a client-side JS check for the `session` cookie's
presence (see git history / AGENT.md v3 if you need the old approach).
That broke once the cookie became `HttpOnly` (§1.6), as predicted.

**Current implementation:** `views.RenderTemplate` (in
`internal/views/render.go`) now takes `r *http.Request` as its second
argument, checks the session server-side via `auth.GetSessionEmail`, and
exposes the result to templates as a `LoggedIn` template function (via
`template.FuncMap`). `base.html` uses `{{if LoggedIn}} ... {{else}} ...
{{end}}` in both the header nav and footer to show the correct links.
No JS, no reliance on a readable cookie. The old `.nav-anon`/`.nav-authed`
CSS classes are no longer used by `base.html` but may still exist in
`style.css` as dead rules — harmless, but could be cleaned up next time
that file is touched.

**This required updating every single handler** that calls
`RenderTemplate`, since the function signature changed (`w, "file.html",
data` → `w, r, "file.html", data"`). All 12 handler files plus
`render.go` were updated as part of this — see §1.6.

**✅ VISUALLY CONFIRMED.** Human confirmed the nav correctly shows the
logged-in state after reload, post-refactor.

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

### 1.5 Auth middleware refactor — COMPLETE and CONFIRMED
Added `internal/middleware/auth.go` with `RequireAuth(userRepo, handler)`
and `RequireRole(userRepo, role, handler)`, both wrapping a handler and
attaching the resolved `*models.User` to the request context via
`context.WithValue`. Handlers retrieve it with
`middleware.UserFromContext(r)` instead of each repeating the cookie →
session → user lookup.

`cmd/server/main.go` now wraps every route that needs a session:
`/dashboard`, `/deposit`, `/withdraw`, `/history`, `/profile`,
`/transfer`, `/profile/edit`, `/delete-account`, `/change-password` all
use `middleware.RequireAuth`. `/admin` uses `middleware.RequireRole(...,
"admin", ...)`. Public routes (`/`, `/about`, `/contact`, `/login`,
`/register`, `/logout`) are untouched.

Every one of the 10 previously-duplicated handlers
(`dashboard.go`, `deposit.go`, `withdraw.go`, `transfer.go`, `history.go`,
`profile.go`, `profile_edit.go`, `delete.go`, `change_password.go`,
`admin.go`) was rewritten to use `middleware.UserFromContext(r)` instead
of its own cookie/session lookup. `admin.go` no longer needs
`UserFromContext` at all, since it doesn't use the current user for
anything — the route-level `RequireRole` wrapper handles both auth and
the admin check before the handler runs.

`delete.go` is the one exception: it still does a direct
`r.Cookie("session")` read, but only to invalidate that specific session
on account deletion — something the middleware doesn't do and shouldn't,
since that's a delete-account-specific side effect, not a general auth
concern.

**Confirmed by the human**, after a full regression pass: all authenticated
pages load correctly, deposit/withdraw/transfer still work, edit
profile/change password still work, logged-out access to `/dashboard`
correctly redirects to `/login`, non-admin access to `/admin` still gets
403, and admin access to `/admin` still works. (One false alarm during
testing — a stale session cookie from before the server restart mid-session
caused a temporary redirect-to-login on `/admin`; resolved by re-logging
in, not a bug in the refactor.)

### 1.6 Session hardening — COMPLETE and CONFIRMED
- `internal/auth/session.go` — replaced the exported `Sessions` map with
  an unexported `sessions` map guarded by a `sync.Mutex` (the old version
  had no concurrency protection at all — a real data race under
  concurrent requests, now fixed as part of this). Sessions now carry an
  `ExpiresAt` (24h TTL, `auth.SessionTTL`) and are lazily deleted on
  lookup if expired. Access is via `CreateSession(email) string`,
  `GetSessionEmail(id) (string, bool)`, `DeleteSession(id)` — nothing
  outside the `auth` package touches the map directly anymore.
- `internal/handlers/auth.go` — `LoginHandler` uses `auth.CreateSession`
  and sets the cookie with `HttpOnly: true`, `Secure: true`,
  `SameSite: http.SameSiteLaxMode`, `MaxAge` matching `SessionTTL`.
  `LogoutHandler` uses `auth.DeleteSession` and clears the cookie with the
  same flags.
- `internal/handlers/delete.go` — same pattern, invalidates the session on
  account deletion.
- `internal/middleware/auth.go` and `internal/auth/current_user.go` —
  updated to call `auth.GetSessionEmail` instead of the removed map.

**⚠️ Important operational caveat, confirmed correct but worth
repeating:** `Secure: true` means the cookie is **only sent over HTTPS**.
This works fine on the dev tunnel URL (`https://opulent-eureka-...
app.github.dev`) but will silently break login if ever run as plain
`http://localhost` without a TLS-terminating proxy in front — the cookie
gets set but the browser won't send it back on the next request, making
every page look logged-out. Don't spend time debugging that as a "bug"
if it happens — it's this flag working as intended in the wrong context.

**Confirmed by the human** via DevTools: the `session` cookie shows
HttpOnly ✓, Secure ✓, SameSite=Lax after logging in on the dev tunnel URL.

---

## 2. Immediate Next Task

The full visual pass (§1.2), auth middleware refactor (§1.5), and session
hardening + server-rendered nav (§1.3, §1.6) are all complete and
confirmed. Nothing is queued/committed beyond that. Remaining candidates:

1. **Remove the `admin@acm.com` hardcoded backdoor** (§3) — replace with
   a real admin-promotion flow now that `Role` exists. This is the last
   item from the original security punch list.
2. Minor: fix the `/admin` "User ID" header wrapping (§1.2, cosmetic).
3. Minor: remove now-dead `.nav-anon`/`.nav-authed` CSS rules from
   `style.css` if they're still there (§1.3).
4. Beyond that: the actual product features (investment plans, loans,
   etc.) — see §3 item 4 and AGENT.md v1 Phases 2–7. Nothing
   backend-hygiene-related is blocking that work anymore.

Ask the human which one before starting, rather than assuming.

---

## 3. Outstanding Items (carried forward, still valid)

1. Hardcoded `admin@acm.com` admin backdoor — still the only way to get
   admin access. **Last remaining security punch-list item.**
2. ~~Auth middleware refactor~~ — DONE, see §1.5.
3. ~~Session hardening~~ — DONE, see §1.6. Nav conversion also DONE, §1.3.
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
