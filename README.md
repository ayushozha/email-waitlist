![Email Waitlist banner](docs/assets/banner.svg)

# Email Waitlist

**Multi-tenant email collection microservice** — collect waitlist signups from any frontend, manage subscribers per project, and optionally send branded confirmation emails via Resend.

[![CI](https://github.com/ayushozha/email-waitlist/actions/workflows/ci.yml/badge.svg)](https://github.com/ayushozha/email-waitlist/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white)](go.mod)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-4169E1?style=flat-square&logo=postgresql&logoColor=white)](internal/database/db.go)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=flat-square&logo=docker&logoColor=white)](Dockerfile)
[![API](https://img.shields.io/badge/API-REST_v1-60a5fa?style=flat-square)](/docs)
[![License](https://img.shields.io/badge/License-not_specified-9ca3af?style=flat-square)](#license)

> **Live:** [emailwaitlist.ayushojha.com](https://emailwaitlist.ayushojha.com) · **API docs:** [/docs](https://emailwaitlist.ayushojha.com/docs) · **Health:** [/health](https://emailwaitlist.ayushojha.com/health)

**Topics:** waitlist · email capture · multi-tenant API · PostgreSQL · Resend · CORS · rate limiting · referral tracking · CSV export

---

## Table of Contents

- [Overview](#overview)
- [Status at a Glance](#status-at-a-glance)
- [Architecture](#architecture)
- [Repository Map](#repository-map)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Authentication](#authentication)
- [API Reference](#api-reference)
- [Frontend Integration](#frontend-integration)
- [Email Confirmations](#email-confirmations)
- [Referral and Position Tracking](#referral-and-position-tracking)
- [Database Schema](#database-schema)
- [Security and Privacy](#security-and-privacy)
- [Deployment](#deployment)
- [Verification](#verification)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

Email Waitlist is a small, production-oriented Go service that lets you add email collection to any website without building signup infrastructure from scratch. Each **project** gets its own API key, CORS policy, subscriber list, stats dashboard data, and optional confirmation email template.

**Why use it instead of rolling your own?**

| Capability | Built in |
|------------|----------|
| Multi-tenant projects with isolated subscriber lists | Yes |
| Per-project publishable (`wl_pub_…`) + secret (`wl_sec_…`) keys | Yes |
| Secret keys stored as SHA-256 hashes | Yes |
| Browser-safe CORS with per-project `allowed_origins` | Yes |
| Subscribe endpoint rate limiting (per IP) | Yes |
| Subscriber metadata (JSON, up to 4 KB) | Yes |
| Paginated list, CSV export, unsubscribe | Yes |
| Signup stats (today / week / month / by day) | Yes |
| Waitlist position + referral attribution | Yes |
| Custom HTML confirmation emails via Resend | Yes (optional) |
| Embedded landing page and interactive API docs | Yes |

The service uses the standard library `net/http` router (Go 1.22+ path patterns), `pgx` for PostgreSQL, and `resend-go` for transactional email. Schema migrations run automatically on startup — no separate migration CLI required.

---

## Status at a Glance

| Area | Status | Notes |
|------|--------|-------|
| Project CRUD (admin) | Implemented | Create + list projects |
| Subscribe / list / delete / export | Implemented | Project-scoped via API key |
| Stats endpoint | Implemented | Totals + 30-day daily breakdown |
| Email templates (per project) | Implemented | GET / PUT / DELETE |
| Resend confirmation emails | Implemented | Disabled without `RESEND_API_KEY` |
| Referral + position tracking | Implemented | Optional fields on subscribe |
| Rate limiting | Implemented | Subscribe only, default 30/min/IP |
| CORS | Implemented | Per-project origin allowlist |
| Homepage + `/docs` | Implemented | Served from embedded HTML |
| Automated tests | Implemented | Unit tests for validation, auth, CORS, rate limiting, templates; `go test ./...` |
| CI | Implemented | GitHub Actions: gofmt, vet, build, test |
| OpenAPI / Swagger spec | Roadmap | Interactive HTML docs at `/docs` today |
| Webhook notifications | Roadmap | e.g. Slack/Discord on new signup |
| Double opt-in | Roadmap | Currently single-step subscribe |

---

## Architecture

### Diagram (static)

![Architecture diagram](docs/assets/architecture.svg)

### Diagram (Mermaid)

```mermaid
flowchart TB
    subgraph Clients
        FE[Browser / Frontend]
        ADM[Admin / curl / scripts]
    end

    subgraph Service["Go HTTP Server :8090"]
        direction TB
        PUB[Public routes<br/>GET / · /docs · /health]
        MW[Middleware chain<br/>Rate limit → Auth → CORS]
        H[Handlers<br/>subscribe · subscribers · stats · projects · email-template]
        PUB --> MW
        MW --> H
    end

    subgraph Data
        PG[(PostgreSQL<br/>projects · subscribers · email_templates)]
    end

    subgraph External
        RS[Resend API<br/>optional confirmation emails]
    end

    FE -->|X-API-Key| MW
    ADM -->|X-Admin-Key| MW
    H -->|pgx pool| PG
    H -.->|async goroutine| RS
```

### Request lifecycle (subscribe)

1. Browser sends `POST /api/v1/subscribe` with `X-API-Key` (publishable key) and JSON body.
2. **Rate limiter** enforces `RATE_LIMIT` requests per minute per client IP (`X-Forwarded-For` is only honoured when `TRUST_PROXY=true`).
3. **API key auth** resolves the project and injects it into request context.
4. **CORS** middleware checks `Origin` against the project's `allowed_origins` and rejects disallowed origins with `403` (auth must run first — the allowlist lives on the project).
5. Handler validates and normalizes the email, inserts subscriber (with position + optional referral), returns `201`.
6. If `RESEND_API_KEY` is configured, a background goroutine renders and sends the confirmation email.

### Middleware and auth layers

| Layer | Applies to | Behaviour |
|-------|------------|-----------|
| Rate limit | `POST /api/v1/subscribe` only | In-memory per-IP counter, configurable window |
| `X-API-Key` | Project-scoped endpoints | Publishable key → subscribe only; secret key (matched by SHA-256 hash) → everything |
| `X-Admin-Key` | `POST/GET /api/v1/projects` | Constant-time compare against server `ADMIN_KEY` env var |
| CORS | All `/api/` routes (runs after auth) | Per-project origin allowlist; handles `OPTIONS` preflight; `Vary: Origin` |

---

## Repository Map

```
email-waitlist/
├── cmd/server/main.go          # Entry point, routing, graceful shutdown
├── internal/
│   ├── config/config.go        # Env-based configuration
│   ├── database/db.go          # pgx pool + schema migrations
│   ├── email/service.go        # Resend client + template rendering
│   ├── handler/                # HTTP handlers (subscribe, stats, docs, …)
│   ├── middleware/             # CORS, rate limit, API/admin auth
│   └── model/                  # Domain types + database queries
├── docs/assets/                # README banner and architecture SVGs
├── Dockerfile                  # Multi-stage Alpine build
├── .env.example                # Required and optional env vars
├── go.mod / go.sum             # Go 1.25, pgx v5, resend-go v3
└── README.md                   # This file
```

---

## Quick Start

### Prerequisites

- **Go 1.25+**
- **PostgreSQL 14+** (17 recommended; `pgcrypto` extension used for UUIDs)
- Optional: **Resend** account and API key for confirmation emails

### 1. Clone and configure

```powershell
git clone git@github.com:ayushozha/email-waitlist.git
cd email-waitlist
Copy-Item .env.example .env
```

Edit `.env` with your database URL and admin key:

```env
DATABASE_URL=postgres://admin:YOUR_PASSWORD@localhost:5432/email_waitlist?sslmode=require
ADMIN_KEY=your-secret-admin-key-here
PORT=8090
```

### 2. Create the database (if needed)

```powershell
# Example: create database on shared Postgres instance
psql "host=postgresql.example.com port=5432 user=admin dbname=postgres sslmode=require" -c "CREATE DATABASE email_waitlist;"
```

### 3. Run locally

```powershell
# Load env vars for this session (PowerShell)
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
    [Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), 'Process')
  }
}

go run ./cmd/server/
```

The server starts on `http://localhost:8090`. Migrations apply automatically on boot.

### 4. Create your first project

```powershell
curl.exe -X POST http://localhost:8090/api/v1/projects `
  -H "Content-Type: application/json" `
  -H "X-Admin-Key: your-secret-admin-key-here" `
  -d '{\"name\":\"My App\",\"slug\":\"my-app\",\"allowed_origins\":[\"http://localhost:3000\"]}'
```

Save the secret `api_key` (`wl_sec_…`) from the response — it is only returned at creation time (the server stores a hash). Use the `public_key` (`wl_pub_…`) in your frontend; it can be retrieved again later via `GET /api/v1/projects`.

### 5. Subscribe a test email

```powershell
curl.exe -X POST http://localhost:8090/api/v1/subscribe `
  -H "Content-Type: application/json" `
  -H "X-API-Key: wl_pub_your_publishable_key" `
  -d '{\"email\":\"user@example.com\",\"metadata\":{\"source\":\"readme-test\"}}'
```

### Docker

```powershell
docker build -t email-waitlist .
docker run --rm -p 8090:8090 `
  -e DATABASE_URL="postgres://..." `
  -e ADMIN_KEY="your-secret-admin-key" `
  email-waitlist
```

---

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | **Yes** | — | PostgreSQL connection string (`postgres://user:pass@host:port/db?sslmode=…`) |
| `ADMIN_KEY` | **Yes** | — | Secret for admin endpoints (`X-Admin-Key` header) |
| `PORT` | No | `8090` | HTTP listen port |
| `RATE_LIMIT` | No | `30` | Max subscribe requests per minute per IP |
| `TRUST_PROXY` | No | `false` | Set `true` only behind a reverse proxy that appends the client IP to `X-Forwarded-For`; otherwise the header is spoofable and ignored |
| `RESEND_API_KEY` | No | — | Enables confirmation emails; service runs without it |
| `DEFAULT_FROM_EMAIL` | No | `Waitlist <waitlist@ayushojha.com>` | Fallback sender when template has no `from_email` |

See [.env.example](.env.example) for a copy-paste template.

---

## Authentication

Each project has **two keys**, plus a server-wide admin key.

| Header | Value format | Used for |
|--------|--------------|----------|
| `X-API-Key` | Publishable key `wl_pub_` + 32 hex chars | `POST /subscribe` only — safe to embed in frontend code |
| `X-API-Key` | Secret key `wl_sec_` + 64 hex chars | All project-scoped endpoints — server-side only |
| `X-Admin-Key` | Server `ADMIN_KEY` | Project creation and listing |

**Why two keys?** The subscribe key necessarily ships in public page source. If that same key could also list and export subscribers, anyone could view-source your landing page and dump your email list. Publishable keys are rejected (`403`) on management endpoints.

**Secret keys** are generated with `crypto/rand` and stored as SHA-256 hashes (`projects.api_key_hash`) — a database leak does not expose usable keys. The plaintext secret is returned exactly once, at project creation. Legacy `wl_…` keys created before the split still work as secret keys.

**Admin key** is a single server-wide secret, compared in constant time. Protect it like a root credential — never embed it in frontend code.

---

## API Reference

Base URL (production): `https://emailwaitlist.ayushojha.com`

Interactive documentation with examples: **[GET /docs](https://emailwaitlist.ayushojha.com/docs)**

### Public (no auth)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Marketing landing page |
| `GET` | `/docs` | Interactive API documentation |
| `GET` | `/health` | Health check — `{"status":"ok"}` |

### Project-scoped (`X-API-Key`)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/subscribe` | Add email to waitlist |
| `GET` | `/api/v1/subscribers` | List subscribers (paginated) |
| `GET` | `/api/v1/subscribers/export` | Download CSV |
| `DELETE` | `/api/v1/subscribers/{email}` | Remove subscriber |
| `GET` | `/api/v1/stats` | Dashboard statistics |
| `GET` | `/api/v1/email-template` | Get custom email template |
| `PUT` | `/api/v1/email-template` | Create or update template |
| `DELETE` | `/api/v1/email-template` | Remove template (revert to default) |

### Admin (`X-Admin-Key`)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/projects` | Create project + API key |
| `GET` | `/api/v1/projects` | List all projects |

---

### `POST /api/v1/subscribe`

Collect an email address. Rate-limited. Returns `201` on success.

**Request body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | Yes | Valid email, max 320 chars, normalized to lowercase |
| `metadata` | object | No | Arbitrary JSON, max 4 KB |
| `referral_code` | string | No | Caller-provided referral slug for this subscriber |
| `referred_by_code` | string | No | Referral code of the referring subscriber |

**Success `201`:**

```json
{
  "message": "Successfully joined the waitlist!",
  "subscriber": {
    "id": "uuid",
    "project_id": "uuid",
    "email": "user@example.com",
    "metadata": {"source": "landing-page"},
    "subscribed_at": "2026-07-04T12:00:00Z",
    "position": 41,
    "referral_code": null,
    "referred_by_id": null,
    "referral_count": 0
  }
}
```

**Errors:**

| Status | Meaning |
|--------|---------|
| `400` | Invalid email, empty body, or metadata too large |
| `401` | Missing or invalid API key |
| `403` | Request `Origin` not in the project's `allowed_origins` |
| `409` | Email already subscribed, or referral code already taken |
| `429` | Rate limit exceeded |

---

### `GET /api/v1/subscribers`

**Query parameters:** `limit` (default 50, max 500), `offset` (default 0)

```json
{
  "subscribers": [ /* … */ ],
  "total": 142,
  "limit": 50,
  "offset": 0
}
```

---

### `GET /api/v1/subscribers/export`

Returns `text/csv` attachment: `email`, `metadata`, `subscribed_at`.

Filename: `{project-slug}-subscribers.csv`

---

### `DELETE /api/v1/subscribers/{email}`

Removes subscriber by email address (URL path). Returns `404` if not found.

---

### `GET /api/v1/stats`

```json
{
  "total": 142,
  "today": 8,
  "this_week": 34,
  "this_month": 89,
  "by_day": [
    {"date": "2026-07-02", "count": 12},
    {"date": "2026-07-03", "count": 14},
    {"date": "2026-07-04", "count": 8}
  ]
}
```

`by_day` covers the last 30 days.

---

### `POST /api/v1/projects` (admin)

**Request body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Display name |
| `slug` | string | Yes | Lowercase alphanumeric + hyphens (e.g. `my-app`) |
| `allowed_origins` | string[] | No | CORS origins; empty = allow all; `["*"]` = wildcard |

**Success `201`:** returns the project including the secret `api_key` (`wl_sec_…`, shown only this once — the server keeps a hash) and the `public_key` (`wl_pub_…`) for frontend embedding. **Save the secret key immediately.**

`GET /api/v1/projects` returns `public_key` for every project but never secret keys.

---

### Email template endpoints

Manage per-project confirmation email content. Template variables: `{{.ProjectName}}`, `{{.Email}}`.

**`PUT /api/v1/email-template` body (all fields optional on update):**

| Field | Type | Description |
|-------|------|-------------|
| `subject` | string | Email subject line |
| `html_body` | string | HTML body (Go `text/template` syntax) |
| `from_name` | string | Display name for sender |
| `from_email` | string | Sender email address |
| `reply_to` | string | Reply-to address |
| `enabled` | boolean | Set `false` to suppress emails for this project |

If no template exists, a built-in default HTML template is used when Resend is enabled.

---

## Frontend Integration

### React / Next.js

```javascript
async function subscribe(email, metadata = {}) {
  const res = await fetch('https://emailwaitlist.ayushojha.com/api/v1/subscribe', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': process.env.NEXT_PUBLIC_WAITLIST_PUBLIC_KEY, // wl_pub_...
    },
    body: JSON.stringify({ email, metadata }),
  });

  const data = await res.json();

  if (res.ok) return { success: true, message: data.message };
  if (res.status === 409) return { success: false, message: "You're already on the list!" };
  if (res.status === 429) return { success: false, message: 'Too many requests. Try again shortly.' };
  return { success: false, message: data.error || 'Something went wrong.' };
}
```

> Use the **publishable key** (`wl_pub_…`) in browser code — it only allows subscribing and cannot read your subscriber list. Keep the secret key (`wl_sec_…`) server-side. For high-abuse surfaces, additionally set `allowed_origins` and consider proxying subscribe calls through your own backend.

### Plain HTML

```html
<form id="waitlist-form">
  <input type="email" id="wl-email" placeholder="you@example.com" required />
  <button type="submit">Join Waitlist</button>
  <p id="wl-msg"></p>
</form>

<script>
document.getElementById('waitlist-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const email = document.getElementById('wl-email').value;
  const msg = document.getElementById('wl-msg');

  const res = await fetch('https://emailwaitlist.ayushojha.com/api/v1/subscribe', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': 'wl_pub_your_publishable_key',
    },
    body: JSON.stringify({ email }),
  });

  const data = await res.json();
  msg.textContent = res.ok ? data.message : (data.error || data.message);
});
</script>
```

### Response handling cheatsheet

| Status | Meaning | User-facing message |
|--------|---------|---------------------|
| `201` | Subscribed | Success message |
| `400` | Invalid input | "Please enter a valid email" |
| `401` | Bad API key | Check your integration config |
| `403` | Origin not allowed | Add your domain to the project's `allowed_origins` |
| `409` | Duplicate | "You're already on the waitlist" |
| `429` | Rate limited | "Please try again in a minute" |

---

## Email Confirmations

Confirmation emails are **optional**. When `RESEND_API_KEY` is unset, subscribers are stored but no email is sent.

When enabled:

1. On successful subscribe, `email.Service.SendConfirmation` runs in a goroutine (bounded by a 30-second timeout).
2. The service loads the project's `email_templates` row (if any). If the lookup *fails* (as opposed to no template existing), the send is skipped — defaults must not override a project that disabled emails.
3. The subject is rendered with Go `text/template`; the HTML body with `html/template`, so subscriber-controlled values like `{{.Email}}` are escaped while your template's own HTML passes through.
4. Email is sent via [Resend](https://resend.com) using per-project or default `from` address.

Failures are logged server-side and do not affect the HTTP response — the subscriber is already persisted.

---

## Referral and Position Tracking

Each subscriber receives a **0-indexed position** per project, assigned at insert time using `MAX(position) + 1`. A unique index guards against concurrent-insert races; the losing insert is retried automatically (up to 3 attempts) rather than surfacing an error.

**Referral fields:**

| Field | Set by | Purpose |
|-------|--------|---------|
| `referral_code` | Client (optional) | Unique slug per subscriber within a project |
| `referred_by_code` | Client (optional) | Looks up referrer; silently ignored if invalid |
| `referred_by_id` | Server | Resolved referrer subscriber ID |
| `referral_count` | Server | Incremented on the referrer when attribution succeeds |

Invalid referral codes do not fail the subscribe request — attribution is best-effort.

---

## Database Schema

Migrations run from `internal/database/db.go` on every startup.

### `projects`

| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID | Primary key, `gen_random_uuid()` |
| `name` | VARCHAR(255) | Display name |
| `slug` | VARCHAR(100) | Unique URL-safe identifier |
| `api_key_hash` | CHAR(64) | SHA-256 of the secret key; unique, indexed |
| `public_key` | VARCHAR(128) | Publishable key (`wl_pub_…`); unique, indexed |
| `allowed_origins` | TEXT[] | CORS allowlist |
| `created_at` | TIMESTAMPTZ | Default `NOW()` |

Legacy databases upgrade in place on startup: plaintext `api_key` values are hashed into `api_key_hash`, the plaintext column is dropped, and missing `public_key` values are generated.

### `subscribers`

| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID | Primary key |
| `project_id` | UUID | FK → `projects`, cascade delete |
| `email` | VARCHAR(320) | Unique per project |
| `metadata` | JSONB | Default `{}` |
| `subscribed_at` | TIMESTAMPTZ | Default `NOW()` |
| `position` | BIGINT | 0-indexed signup rank per project |
| `referral_code` | TEXT | Optional, unique per project when set |
| `referred_by_id` | UUID | FK → `subscribers`, set null on delete |
| `referral_count` | INT | Default 0 |

**Indexes:** `project_id`, `subscribed_at`, `(project_id, position)`, `(project_id, referral_code)`, `referred_by_id`

### `email_templates`

| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID | Primary key |
| `project_id` | UUID | FK → `projects`, one template per project |
| `subject` | VARCHAR(500) | Default "You're on the waitlist!" |
| `html_body` | TEXT | Custom HTML |
| `from_name` | VARCHAR(255) | Optional |
| `from_email` | VARCHAR(320) | Optional |
| `reply_to` | VARCHAR(320) | Optional |
| `enabled` | BOOLEAN | Default `true` |
| `created_at` / `updated_at` | TIMESTAMPTZ | Auto-managed |

---

## Security and Privacy

| Topic | Implementation |
|-------|----------------|
| **Key separation** | Publishable key (subscribe only, browser-safe) vs secret key (management, server-only) |
| **Key storage** | Secret keys stored as SHA-256 hashes; plaintext shown once at creation |
| **Tenant isolation** | All subscriber queries filter by `project_id` from authenticated API key |
| **Admin separation** | Admin key compared in constant time; never exposed to browsers |
| **Rate limiting** | Subscribe endpoint only; `X-Forwarded-For` trusted only with `TRUST_PROXY=true` |
| **Input limits** | Body caps: 10 KB subscribe, 16 KB projects, 256 KB templates; 4 KB metadata; email validated and normalized via `net/mail` |
| **CORS / origin check** | Per-project allowlist enforced server-side (`403` on mismatch); empty list allows all origins |
| **CSV export** | Cells starting with formula characters are neutralized against spreadsheet injection |
| **Email rendering** | Body rendered with `html/template` — subscriber data is escaped |
| **Secrets** | `ADMIN_KEY`, `DATABASE_URL`, `RESEND_API_KEY` via environment only |
| **Email privacy** | Subscriber emails stored in your PostgreSQL instance; export via authenticated API |
| **HTTPS** | Required in production (enforced at reverse proxy / hosting layer) |

**Operational guidance:**

- Rotate `ADMIN_KEY` if compromised; API keys are per-project and can be recreated by making a new project.
- Use explicit `allowed_origins` in production instead of leaving the list empty.
- Behind a reverse proxy, set `TRUST_PROXY=true` so rate limiting keys on the real client IP; without a proxy, leave it `false` (the default) — otherwise the limit is trivially bypassable.

---

## Deployment

### Production instance

- **URL:** https://emailwaitlist.ayushojha.com
- **Port:** 8090 (container / process)
- **Database:** PostgreSQL (`email_waitlist` database on shared infrastructure)

### Docker image

The [Dockerfile](Dockerfile) uses a multi-stage build:

1. `golang:1.25-alpine` — compile static binary (`CGO_ENABLED=0`, dependencies verified against `go.sum`)
2. `alpine:3.21` — minimal runtime with CA certificates, runs as non-root user

```powershell
docker build -t email-waitlist .
docker run -d -p 8090:8090 --env-file .env email-waitlist
```

### Graceful shutdown

The server handles `SIGINT` and `SIGTERM`, allowing up to 10 seconds for in-flight requests to complete before exit.

---

## Verification

CI runs gofmt, vet, build, and tests on every PR ([.github/workflows/ci.yml](.github/workflows/ci.yml)). Locally:

```powershell
# Compile
go build ./...

# Static analysis
go vet ./...

# Unit tests (validation, auth, CORS, rate limiting, template rendering)
go test ./... -count=1

# Health check (with server running)
curl.exe http://localhost:8090/health
# Expected: {"status":"ok"}
```

---

## Roadmap

- [ ] OpenAPI 3 spec generated from handlers
- [x] `go test` coverage for validation, middleware, rendering, and DB migrations (migration tests are opt-in via `TEST_DATABASE_URL` pointing at a throwaway database)
- [ ] Webhook on new subscriber (Slack, Discord, custom URL)
- [ ] Double opt-in flow with confirmation tokens
- [ ] API key rotation without recreating projects
- [ ] Admin dashboard UI (today: API + `/docs` only)

---

## Contributing

When contributing, preserve these invariants:

1. **Multi-tenant isolation** — never query subscribers without `project_id` from authenticated context.
2. **Env-driven config** — no hardcoded secrets, ports, or provider keys in source.
3. **Backward-compatible migrations** — use `IF NOT EXISTS` / `ADD COLUMN IF NOT EXISTS` patterns in `db.go`.
4. **PR workflow** — branch from `main`, open a PR; do not push directly to `main`.
5. **Minimal dependencies** — prefer stdlib; justify new packages in the PR description.

---

## License

No license file is currently checked in to this repository. All rights reserved by the repository owner unless a `LICENSE` file is added later.

---

<p align="center">
  Built by <a href="https://ayushojha.com">Ayush Ojha</a> ·
  <a href="https://emailwaitlist.ayushojha.com">Live service</a> ·
  <a href="https://emailwaitlist.ayushojha.com/docs">API docs</a>
</p>