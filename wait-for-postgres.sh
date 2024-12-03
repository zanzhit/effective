#!/bin/sh

set -e

host="$1"
port="$2"
shift 2
cmd="$@"

until PGPASSWORD="$DB_PASSWORD" psql -h "$host" -U "$DB_USER" -p "$port" -c '\l'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - executing command"
exec $cmd