#!/usr/bin/env bash
# render-placeholders.sh — substitute {{PLACEHOLDER}} tokens across legal + root markdown.
# Usage:
#   scripts/render-placeholders.sh --dry-run            # list all unresolved placeholders
#   scripts/render-placeholders.sh --check              # exit 1 if any placeholder unresolved
#   scripts/render-placeholders.sh --from .env.render   # substitute in place from env file
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TARGETS=(
  "$ROOT/MARKETING-RULES.md"
  "$ROOT/README.md"
  "$ROOT/SECURITY.md"
  "$ROOT/CONTRIBUTING.md"
  "$ROOT/web/public/.well-known/security.txt"
)

mode="${1:---dry-run}"

list_placeholders() {
  local existing=()
  for t in "${TARGETS[@]}"; do [ -e "$t" ] && existing+=("$t"); done
  [ "${#existing[@]}" -gt 0 ] || return 0
  grep -rhoE '\{\{[A-Z_][A-Z0-9_]*\}\}' "${existing[@]}" 2>/dev/null | sort -u
}

case "$mode" in
  --dry-run|--check)
    found="$(list_placeholders || true)"
    if [ -z "$found" ]; then
      echo "No unresolved placeholders."
      exit 0
    fi
    echo "Unresolved placeholders:"
    echo "$found"
    [ "$mode" = "--check" ] && exit 1 || exit 0
    ;;
  --from)
    env_file="${2:?usage: --from <env-file>}"
    # shellcheck disable=SC1090
    set -a; . "$env_file"; set +a
    for ph in $(list_placeholders); do
      name="${ph#\{\{}"; name="${name%\}\}}"
      val="${!name:-}"
      if [ -z "$val" ]; then
        echo "WARNING: $ph not in env file, skipping"; continue
      fi
      existing=()
      for t in "${TARGETS[@]}"; do [ -e "$t" ] && existing+=("$t"); done
      find "${existing[@]}" -type f \( -name '*.md' -o -name 'security.txt' \) -print0 \
        | xargs -0 sed -i.bak "s|{{${name}}}|${val//|/\\|}|g"
      find "${existing[@]}" -type f -name '*.bak' -delete
    done
    ;;
  *) echo "usage: $0 [--dry-run|--check|--from <env-file>]"; exit 2 ;;
esac
