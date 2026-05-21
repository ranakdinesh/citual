#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CITUAL_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
WORKSPACE_ROOT="${CITUAL_DEV_WORKSPACE:-$(cd "${CITUAL_ROOT}/.." && pwd)}"
STATE_DIR="${CITUAL_ROOT}/deployments/.deploy"
PREVIOUS_COMMITS="${STATE_DIR}/previous_commits.tsv"

if [ ! -f "$PREVIOUS_COMMITS" ]; then
  echo "✗ No previous deployment state found at ${PREVIOUS_COMMITS}" >&2
  exit 1
fi

while IFS=$'\t' read -r repo commit; do
  [ -n "$repo" ] || continue
  if [ ! -d "${WORKSPACE_ROOT}/${repo}/.git" ]; then
    echo "Skipping ${repo}; repository is not present in ${WORKSPACE_ROOT}."
    continue
  fi
  echo "→ Rolling back ${repo} to ${commit}"
  git -C "${WORKSPACE_ROOT}/${repo}" fetch origin --tags
  git -C "${WORKSPACE_ROOT}/${repo}" reset --hard "$commit"
done <"$PREVIOUS_COMMITS"

"${SCRIPT_DIR}/deploy-dev.sh"
