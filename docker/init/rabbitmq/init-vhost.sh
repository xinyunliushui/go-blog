#!/bin/sh
set -eu

VHOST="${RABBITMQ_VHOST:?RABBITMQ_VHOST is required}"
USER="${RABBITMQ_USER:?RABBITMQ_USER is required}"
PASS="${RABBITMQ_PASSWORD:?RABBITMQ_PASSWORD is required}"
MAX_RETRY="${MAX_RETRY:-60}"

retry=0
until curl -sf -u "${USER}:${PASS}" "http://rabbitmq:15672/api/overview" >/dev/null; do
  retry=$((retry + 1))
  if [ "${retry}" -ge "${MAX_RETRY}" ]; then
    echo "rabbitmq management api not ready after ${MAX_RETRY} attempts"
    exit 1
  fi
  echo "waiting for rabbitmq management api... (${retry}/${MAX_RETRY})"
  sleep 2
done

curl -sf -u "${USER}:${PASS}" -X PUT "http://rabbitmq:15672/api/vhosts/${VHOST}"
curl -sf -u "${USER}:${PASS}" \
  -H "content-type: application/json" \
  -X PUT "http://rabbitmq:15672/api/permissions/${VHOST}/${USER}" \
  -d '{"configure":".*","write":".*","read":".*"}'

echo "rabbitmq vhost ${VHOST} ready"
