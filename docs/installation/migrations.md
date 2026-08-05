## Database Migrations

Faridoon applies sql-migrate upgrades every time the container starts (`sql-migrate up` in the entrypoint for the active driver under `database/<driver>/`).

### Development

```bash
make migrate
# or: make -C database/mysql
```

### Container

```bash
docker exec -it faridoon /bin/sh
cd /var/faridoon/database/mysql
sql-migrate up
```

Set `DB_DRIVER` if needed (default `mysql`). sql-migrate expects `DB_HOST`, `DB_USER`, `DB_PASS`, and `DB_NAME` (the entrypoint maps Laravel-style aliases when present).

Current required migration id: `13.webhook-targets-events.sql` (splits legacy `webhooks` into `webhook_targets` + `webhook_events`; earlier migrations include quotes FULLTEXT search, cvar category/ordinal, cvar title/description, the cvars table, quote submitter columns, header links, audit logs, and password column width).
