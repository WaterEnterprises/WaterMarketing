# Water Enterprises — Marketing & Investor Suite

Marketing operations, lead management CRM, for investor, sponsors, etc...

## Repo Structure

```
├── cmd/crm.go              # CRM binary (CLI + HTTP API)
├── internal/db/            # Database layer (SQLCipher-encrypted SQLite)
├── web/                    # CRM dashboard (Svelte + Skeleton + Tailwind)
├── databases/              # Encrypted leads.db, mail-credentials.db
├── WaterParty/             # Pitch decks, guides, email templates, investor docs
│   ├── pitch-deck-outline.md / .pt.md
│   ├── executive-summary.md
│   ├── financial-projections.md
│   ├── sponsorship-deck.md / .pt.md
│   ├── email-*.txt
│   └── ...
├── Makefile
└── README.md
```

## Quick Start

### Prerequisites

- [Go](https://go.dev/dl/) 1.25+
- [Bun](https://bun.sh) 1.3+

### Build everything

```bash
make
```

This builds `crm.exe` and the `web/` frontend.

### Run the CRM API server

```bash
make ARGS=serve run
# or
.\crm.exe serve
```

API listens on `http://localhost:8080`.

### Run the CRM dashboard (dev mode)

```bash
make web-dev
# or
cd web && bun run dev
```

Opens on `http://localhost:5173`. Proxies `/api/*` to the Go backend.

## CRM Commands

| Command | Description |
|---|---|
| `crm stats` | Dashboard overview |
| `crm list` | List all leads |
| `crm list --status cold --tier 1` | Filter by status / tier |
| `crm list --search "Monashees"` | Search |
| `crm add` | Interactive add lead |
| `crm view <id>` | Lead detail + activity log |
| `crm update <id>` | Interactive update |
| `crm delete <id>` | Delete lead |
| `crm status <id> <new_status>` | Quick status change |
| `crm log <id>` | Log activity (email/call/meeting/note) |
| `crm followups` | Show due follow-ups |
| `crm import --path file.csv` | Import from CSV |
| `crm export --path file.csv` | Export to CSV |
| `crm mail --emails "a@b.com" --subject "Hi" --body "Hello"` | Send BCC email via Gmail SMTP |
| `crm password` | Store Gmail app password (secure prompt) |
| `crm telegram` | Send campaign draft to Telegram |
| `crm campaign --send` | Run cold email campaign |
| `crm serve` | Start Fiber HTTP API |

## Environment

A `.env` file in the repo root with:

```env
EMAIL_DB_PASSWORD="your_sqlcipher_password"
```

Optionally for Telegram:

```env
TELEGRAM_BOT_TOKEN="your_bot_token"
TELEGRAM_CHAT_ID="your_chat_id"
```

## Stack

| Layer | Technology |
|---|---|
| Backend API | Go + Fiber |
| Frontend | Svelte 4 + Skeleton UI + Tailwind CSS |
| Runtime | Bun (dev server, package manager) |
| Database | SQLCipher-encrypted SQLite |

---

*Built by **Water Enterprises** (Stellarium Foundation)*
