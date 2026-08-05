#!/bin/sh
set -e
if [ -f /config/config.yaml ]; then
  export FARIDOON_CONFIG_FILE=/config/config.yaml
fi
# Support legacy env names for sql-migrate
if [ -n "${DB_DATABASE:-}" ] && [ -z "${DB_NAME:-}" ]; then
  export DB_NAME="$DB_DATABASE"
fi
if [ -n "${DB_USERNAME:-}" ] && [ -z "${DB_USER:-}" ]; then
  export DB_USER="$DB_USERNAME"
fi
if [ -n "${DB_PASSWORD:-}" ] && [ -z "${DB_PASS:-}" ]; then
  export DB_PASS="$DB_PASSWORD"
fi

DRIVER="${DB_DRIVER:-mysql}"
case "$DRIVER" in
  mysql|pgsql|sqlite) ;;
  *)
    echo "unsupported DB_DRIVER: $DRIVER" >&2
    exit 1
    ;;
esac

cd "/var/faridoon/database/${DRIVER}" && sql-migrate up
exec /usr/bin/faridoon-service "$@"
