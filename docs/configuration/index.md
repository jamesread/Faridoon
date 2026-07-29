# Configuration

Faridoon reads `/config/config.yaml` (or `FARIDOON_CONFIG_FILE`). Environment variables override database settings for Docker Compose compatibility.

## Database

- `database.host` / `DB_HOST`
- `database.user` / `DB_USER` or `DB_USERNAME`
- `database.password` / `DB_PASS` or `DB_PASSWORD`
- `database.name` / `DB_NAME` or `DB_DATABASE`

sql-migrate (on container start) expects `DB_HOST`, `DB_USER`, `DB_PASS`, and `DB_NAME`. The entrypoint maps Laravel-style names to these when needed.

## Application

- `siteTitle` / `SITE_TITLE`: Header title
- `requiredMigration` / `REQUIRED_MIGRATION`: Expected sql-migrate id (default `6.audit-logs.sql`)
- `listen`: HTTP listen address (default `:8080`)

## Feature flags (`features`)

- `enableVoting` (default true)
- `enableSyntaxHighlighting` (default false)
- `disableRegistration` (default false)
- `guestsDisableAdd` (default false)

## Auth (`auth`)

httpauthshim session settings (cookie name, session file directory). User accounts and passwords live in the MySQL `users` table. Legacy sha1/bcrypt hashes are accepted and upgraded to argon2id on login.
