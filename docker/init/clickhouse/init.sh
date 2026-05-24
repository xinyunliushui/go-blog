#!/bin/sh
set -eu

DB="${CLICKHOUSE_DB:?CLICKHOUSE_DB is required}"
HOST="${CLICKHOUSE_HOST:-clickhouse}"
USER="${CLICKHOUSE_USER:?CLICKHOUSE_USER is required}"
PASS="${CLICKHOUSE_PASSWORD:?CLICKHOUSE_PASSWORD is required}"
MAX_RETRY="${MAX_RETRY:-30}"

ch_query() {
  if ! clickhouse-client \
    --host "${HOST}" \
    --user "${USER}" \
    --password "${PASS}" \
    --query "$1"
  then
    echo "clickhouse query failed: $1"
    exit 1
  fi
}

retry=0
until ch_query "SELECT 1" >/dev/null 2>&1; do
  retry=$((retry + 1))
  if [ "${retry}" -ge "${MAX_RETRY}" ]; then
    echo "clickhouse not ready after ${MAX_RETRY} attempts"
    exit 1
  fi
  echo "waiting for clickhouse... (${retry}/${MAX_RETRY})"
  sleep 2
done

ch_query "CREATE DATABASE IF NOT EXISTS ${DB}"

ch_query "CREATE TABLE IF NOT EXISTS ${DB}.blogs (
    id String,
    title String,
    content String,
    summary String,
    cover_image String,
    category Nullable(String),
    tags Nullable(String),
    status UInt8,
    author String,
    published_at Nullable(DateTime),
    created_at DateTime,
    updated_at DateTime,
    deleted_at Nullable(DateTime)
) ENGINE = MergeTree() ORDER BY id"

echo "clickhouse database ${DB} ready"
