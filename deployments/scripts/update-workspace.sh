#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CITUAL_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
WORKSPACE_ROOT="${CITUAL_DEV_WORKSPACE:-$(cd "${CITUAL_ROOT}/.." && pwd)}"
TARGET_BRANCH="${DEV_DEPLOY_BRANCH:-main}"
ALLOW_DIRTY="${DEV_DEPLOY_ALLOW_DIRTY:-false}"
REPOS="${DEV_DEPLOY_PULL_REPOS:-citual citual-web spur-identity spur-engage spur-messaging spur-storage spur-template}"

repo_url() {
  case "$1" in
    citual) echo "https://github.com/ranakdinesh/citual.git" ;;
    citual-web) echo "https://github.com/ranakdinesh/citual-web.git" ;;
    spur-identity) echo "https://github.com/ranakdinesh/spur-identity.git" ;;
    spur-engage) echo "https://github.com/ranakdinesh/spur-engage.git" ;;
    spur-messaging) echo "https://github.com/ranakdinesh/spur-messaging.git" ;;
    spur-storage) echo "https://github.com/ranakdinesh/spur-storage.git" ;;
    spur-template) echo "https://github.com/ranakdinesh/spur-template.git" ;;
    spur-web) echo "https://github.com/ranakdinesh/spur-web.git" ;;
    *) return 1 ;;
  esac
}

ensure_repo() {
  local name="$1"
  local dir="${WORKSPACE_ROOT}/${name}"
  local url
  url="$(repo_url "$name")"

  if [ ! -d "$dir/.git" ]; then
    echo "→ Cloning ${name}"
    git clone "$url" "$dir"
  fi

  cd "$dir"
  if [ "$ALLOW_DIRTY" != "true" ] && [ -n "$(git status --porcelain)" ]; then
    echo "✗ ${name} has uncommitted changes. Commit/stash them or set DEV_DEPLOY_ALLOW_DIRTY=true." >&2
    git status --short >&2
    exit 1
  fi

  echo "→ Updating ${name}"
  git fetch origin --tags
  git checkout "$TARGET_BRANCH"
  git reset --hard "origin/${TARGET_BRANCH}"
}

ensure_go_workspace() {
  local go_work="${WORKSPACE_ROOT}/go.work"
  local go_work_sum="${WORKSPACE_ROOT}/go.work.sum"

  if [ ! -f "$go_work" ]; then
    cat >"$go_work" <<'EOF'
go 1.26.2

use (
	./citual
	./spur-identity
	./spur-engage
	./spur-messaging
	./spur-storage
	./spur-template
	./spur-web
)

replace github.com/ranakdinesh/spur-identity v1.1.6 => ./spur-identity
replace github.com/ranakdinesh/spur-engage v0.1.0 => ./spur-engage
replace github.com/ranakdinesh/spur-messaging v1.0.5 => ./spur-messaging
replace github.com/ranakdinesh/spur-storage v0.1.1 => ./spur-storage
replace github.com/ranakdinesh/spur-template v0.1.1 => ./spur-template
EOF
  fi

  if [ ! -f "$go_work_sum" ]; then
    touch "$go_work_sum"
  fi
}

mkdir -p "$WORKSPACE_ROOT"

for repo in $REPOS; do
  ensure_repo "$repo"
done

ensure_go_workspace

echo "✓ Workspace updated at ${WORKSPACE_ROOT}"
