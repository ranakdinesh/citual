#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CITUAL_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
WORKSPACE_ROOT="${CITUAL_DEV_WORKSPACE:-$(cd "${CITUAL_ROOT}/.." && pwd)}"
DEPLOY_DIR="${CITUAL_ROOT}/deployments"
STATE_DIR="${DEPLOY_DIR}/.deploy"
COMPOSE_FILE="${DEPLOY_DIR}/docker-compose.yml"
ENV_FILE="${DEPLOY_DIR}/.env"
COMPOSE=(docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE")

record_commits() {
  local output="$1"
  : >"$output"
  for repo in citual citual-web spur-identity spur-engage spur-messaging spur-storage spur-template; do
    if [ -d "${WORKSPACE_ROOT}/${repo}/.git" ]; then
      git -C "${WORKSPACE_ROOT}/${repo}" rev-parse HEAD | awk -v repo="$repo" '{print repo "\t" $1}' >>"$output"
    fi
  done
}

print_logs_on_failure() {
  echo ""
  echo "Deployment failed. Recent backend and web logs follow." >&2
  "${COMPOSE[@]}" logs --tail=160 backend citual-web >&2 || true
}

require_file() {
  if [ ! -f "$1" ]; then
    echo "✗ Required file missing: $1" >&2
    exit 1
  fi
}

main() {
  cd "$CITUAL_ROOT"
  require_file "$ENV_FILE"
  require_file "$COMPOSE_FILE"

  mkdir -p "$STATE_DIR"
  if [ -f "${STATE_DIR}/current_commits.tsv" ]; then
    cp "${STATE_DIR}/current_commits.tsv" "${STATE_DIR}/previous_commits.tsv"
  fi
  record_commits "${STATE_DIR}/deploying_commits.tsv"

  if [ ! -f "${CITUAL_ROOT}/keys/private.pem" ]; then
    echo "→ Generating development signing key"
    make keys
  fi

  echo "→ Validating compose file"
  "${COMPOSE[@]}" config -q

  echo "→ Building and starting Citual dev stack"
  trap print_logs_on_failure ERR
  "${COMPOSE[@]}" up -d --build --remove-orphans

  echo "→ Running health checks"
  "${SCRIPT_DIR}/healthcheck.sh"
  trap - ERR

  cp "${STATE_DIR}/deploying_commits.tsv" "${STATE_DIR}/current_commits.tsv"
  date -u +"%Y-%m-%dT%H:%M:%SZ" >"${STATE_DIR}/last_successful_deploy_at"
  git -C "$CITUAL_ROOT" rev-parse HEAD >"${STATE_DIR}/last_successful_citual_commit"

  echo "✓ Dev deployment complete"
}

main "$@"
