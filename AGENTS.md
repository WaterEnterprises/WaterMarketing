# Water Enterprises — Agent Instructions

## Project Overview
Marketing operations & lead management CRM for Water Enterprises (Stellarium Foundation). Go CLI/API backend + Svelte dashboard frontend.

## Build & Run

```bash
# Build everything (Go binary + web frontend)
make

# Build individually
make build        # go build -o crm.exe ./cmd
make web-build    # cd web && bun run build
make web-install  # cd web && bun install

# Run
make serve        # .\crm.exe serve  (port :8080)
make web-dev      # cd web && bun run dev  (port :5173)
make ARGS=serve run

# Clean
make clean
```

## Project Structure

```
├── cmd/crm.go               # Main binary: CLI subcommands + Fiber HTTP server
├── internal/db/db.go         # Database layer (SQLCipher-encrypted SQLite)
├── web/                      # SvelteKit + Skeleton UI + Tailwind dashboard
│   ├── src/lib/
│   │   ├── api.ts            # Typed API client
│   │   ├── store.ts          # Svelte stores (activeTab)
│   │   └── components/
│   │       ├── MainContent.svelte   # Main single-page UI (tabs, leads table, email modal, pagination)
│   │       ├── BottomNav.svelte      # Bottom navigation bar
│   │       └── StatsGrid.svelte      # Dashboard stats cards
│   └── build/                # Static output (gitignored)
├── databases/                # SQLCipher-encrypted DBs (leads.db, mail-credentials.db)
├── WaterParty/               # Pitch decks, email templates, investor docs
├── Makefile
├── AGENTS.md
└── README.md
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25+ |
| HTTP server | github.com/gofiber/fiber/v2 |
| Database | github.com/mutecomm/go-sqlcipher/v4 (SQLCipher SQLite) |
| Frontend | Svelte 4 + Skeleton UI (crimson theme) + Tailwind CSS |
| Runtime | Bun 1.3+ (dev server, package manager, builder) |

## Database Conventions

- **SQLCipher password** from `.env`: `EMAIL_DB_PASSWORD="Alpha789@"`
- **DB paths**: `databases/leads.db`, `databases/mail-credentials.db`
- **Leads schema**: `id TEXT PK`, `company TEXT NOT NULL UNIQUE`, `contact_name`, `email`, `phone`, `website`, `tier TEXT NOT NULL`, `type TEXT NOT NULL`, `vertical`, `check_size`, `pitch_angle`, `status`, `next_action`, `next_action_date`, `notes`, `source`, `created_at`, `updated_at`
- **Unique index**: `idx_leads_company` on `company`
- **Mail credentials**: `credentials` table with `email`, `app_password`

## API Endpoints (Fiber)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/stats` | Dashboard stats |
| GET | `/api/leads` | Paginated leads (`?page=&limit=&search=&status=&tier=`) |
| GET | `/api/leads/:id` | Single lead |
| POST | `/api/leads` | Create lead (validates company/tier/type required, 409 on duplicate) |
| PUT | `/api/leads/:id` | Update lead |
| DELETE | `/api/leads/:id` | Delete lead |
| PUT | `/api/leads/:id/status` | Update status only |
| GET | `/api/leads/:id/outreach` | Activity log for lead |
| POST | `/api/leads/:id/outreach` | Log activity (`activity_type`, `notes`, `outcome`) |
| GET | `/api/followups` | Due follow-ups |
| POST | `/api/send-email` | BCC email `{emails: string[], subject, body}` |
| POST | `/api/leads/export` | CSV export `{ids: string[]}` → file download |

## Frontend Conventions

- **Single-page app**: Dashboard + Leads tabs via BottomNav
- **SPA fallback**: Fiber serves `web/build/`; unknown routes serve `index.html`
- **Vite proxy**: `/api/*` → `http://localhost:8080` (in `vite.config.ts`)
- **Theme**: `<html class="dark">`, `<body data-theme="crimson">`
- **API client**: All calls in `web/src/lib/api.ts` — uses `fetch()`, returns typed promises
- **Checkbox + bulk actions**: Leads table has checkboxes, sticky action bar for Send Email / Export CSV
- **Email modal**: Subject + Body fields, validates at least one recipient with valid emails

## Go Code Conventions

- Package: `main` in `cmd/crm.go`, `db` in `internal/db/db.go`
- Module: `waterenterprises`
- CLI subcommands registered via `flag.NewFlagSet` in `main()` dispatch
- Fiber endpoints registered in `cmdServe()` under `/api` group
- Error handling: `os.Exit(1)` in CLI, `c.Status(NNN).JSON(...)` in API
- All DB errors bubble up as JSON error responses

## Security Notes

- `.env` is gitignored — store `EMAIL_DB_PASSWORD` there
- Gmail app password stored in `mail-credentials.db` (also SQLCipher-encrypted)
- Telegram bot token in `.env` (optional)
- Never log or expose DB passwords or API keys
- Never commit `databases/*.db` files
