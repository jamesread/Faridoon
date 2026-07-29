## Database Migrations

Faridoon applies sql-migrate upgrades every time the container starts (`sql-migrate up` in the entrypoint).

To run manually:

```bash
docker exec -it faridoon /bin/sh
cd /var/faridoon/database
sql-migrate up
```

Current required migration id: `6.audit-logs.sql` (audit log table; earlier migrations include webhooks and password column width).
