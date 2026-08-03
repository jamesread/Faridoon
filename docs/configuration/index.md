# Configuration

Faridoon loads a YAML config file, then applies a small set of environment overrides for database connectivity (and a few bootstrap values). Runtime feature toggles live in the database as **configuration variables (cvars)**, edited under **Account → Settings**.

## Config file location

Search order:

1. `--configdir <dir>/config.yaml` (if the flag is set)
2. `./config.yaml`
3. `./config/config.yaml`
4. `$FARIDOON_CONFIG_FILE` (if set and the path exists)
5. `/config/config.yaml`

Docker images typically set `FARIDOON_CONFIG_FILE=/config/config.yaml`.

## Database

These keys may be set in YAML or overridden by environment variables:

- `database.host` / `DB_HOST`
- `database.port` (default `3306`; YAML only)
- `database.user` / `DB_USER` or `DB_USERNAME`
- `database.password` / `DB_PASS` or `DB_PASSWORD`
- `database.name` / `DB_NAME` or `DB_DATABASE`

sql-migrate (on container start) expects `DB_HOST`, `DB_USER`, `DB_PASS`, and `DB_NAME`. The entrypoint maps Laravel-style names to these when needed.

## Application (YAML / env)

- `siteTitle` / `SITE_TITLE`: initial site title used to **seed** the `site_title` cvar on first startup. After that, the live header and document title come from Settings.
- `requiredMigration` / `REQUIRED_MIGRATION`: expected sql-migrate id (default `9.cvars.sql`)
- `listen`: HTTP listen address (default `:8080`)
- `PORT`: if set, overrides `listen` (bare port like `8080` or a full address)

Feature flags are **not** environment variables and are **not** set under a YAML `features:` block.

## Settings (configuration variables)

Admins manage cvars at **Account → Settings** (`/admin/settings`). Missing defaults are inserted on startup if they do not already exist.

| Key | Type | Default | Effect |
|-----|------|---------|--------|
| `site_title` | string | from `siteTitle` / `SITE_TITLE` | Header and document (`<title>`) site title |
| `quotes_per_page` | int | 5 | Quotes shown per page on listing (1–127) |
| `enable_voting` | bool | off | Show vote controls; allow `VoteQuote` |
| `enable_registration` | bool | on | Allow `/register` and `Register` |
| `enable_guest_add` | bool | on | Allow logged-out users to submit quotes |
| `enable_syntax_highlighting` | bool | off | Show syntax highlighting field when adding/editing quotes |

Changing Settings reloads Init so the UI picks up new values without a process restart.

## Privileges

Default groups: **Admins** (id 1) have `SUPERUSER`; **Users** (id 2) have no privileges until granted.

- `SUPERUSER` — admin access (settings, users, webhooks, header links, diagnostics, audit logs, delete quotes). Implies all other privileges.
- `APPROVE_QUOTES` — approvals queue, approve/reject pending quotes, edit quotes.
- `BYPASS_APPROVAL` — new quotes are published without approval.

The first registered user joins Admins. Later registrations join Users. Rejecting a pending quote permanently deletes it. The last remaining admin cannot be demoted, deleted, or stripped of `SUPERUSER`.

## Auth (`auth`)

httpauthshim session settings. Sample keys:

- `localSessionCookieName` (e.g. `faridoon-sid`)
- `baseDir` (session storage directory)

User accounts and passwords live in the MySQL `users` table. Legacy sha1/bcrypt hashes are accepted and upgraded to argon2id on login. Authorization always uses privileges loaded from the database for the current request.
