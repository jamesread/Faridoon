## Database Migrations

Faridoon applies sql-migrate upgrades every time the container starts (`sql-migrate up` in the entrypoint).

To run manually:

```bash
docker exec -it faridoon /bin/sh
cd /var/faridoon/database
sql-migrate up
```

Current required migration id: `9.cvars.sql` (configuration variables table; earlier migrations include quote submitter columns, header links, audit logs, webhooks and password column width).
