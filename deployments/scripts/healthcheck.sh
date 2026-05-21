#!/usr/bin/env bash
set -euo pipefail

BACKEND_HEALTH_URL="${BACKEND_HEALTH_URL:-http://127.0.0.1:9090/healthz}"
WEB_HEALTH_URL="${WEB_HEALTH_URL:-http://127.0.0.1:8085/api/health}"
ATTEMPTS="${HEALTHCHECK_ATTEMPTS:-30}"
SLEEP_SECONDS="${HEALTHCHECK_SLEEP_SECONDS:-3}"

check_url() {
  local name="$1"
  local url="$2"
  local attempt

  for attempt in $(seq 1 "$ATTEMPTS"); do
    if curl -fsS --max-time 5 "$url" >/dev/null; then
      echo "✓ ${name} healthy: ${url}"
      return 0
    fi
    echo "Waiting for ${name} (${attempt}/${ATTEMPTS})..."
    sleep "$SLEEP_SECONDS"
  done

  echo "✗ ${name} did not become healthy: ${url}" >&2
  return 1
}

check_url "backend" "$BACKEND_HEALTH_URL"
check_url "citual-web" "$WEB_HEALTH_URL"
