## Database Migrations

Faridoon applies sql-migrate upgrades every time the container starts (`sql-migrate up` in the entrypoint).

To run manually:

```bash
docker exec -it faridoon /bin/sh
cd /var/faridoon/database
sql-migrate up
```

Current required migration id: `12.quotes-content-fulltext.sql` (FULLTEXT index on `quotes.content` for search; earlier migrations include cvar category/ordinal, cvar title/description, the cvars table, quote submitter columns, header links, audit logs, webhooks and password column width).
