#!/usr/bin/env bash
# srt-cli.sh — bash wrapper around /admin/subject-rights/tickets endpoints.
# Reuses a session cookie-jar across invocations. For staff use only.
#
# Subcommands:
#   login <email> <password>
#   create <kind> <requester_email> [notes]
#   list [--overdue] [--status <new|in_progress|closed|rejected>]
#   patch <id> <json_body>
#   close <id> <closed|rejected> [final_notes]
set -euo pipefail

BASE_URL="${PESA_BASE_URL:-https://localhost:8443}"
JAR="${PESA_COOKIE_JAR:-$HOME/.cache/pesa-srt-cookies.txt}"
mkdir -p "$(dirname "$JAR")"

usage() {
  cat <<USAGE
Usage: $0 <login|create|list|patch|close> [args]
  login  <email> <password>
  create <kind> <requester_email> [notes]
  list   [--overdue] [--status STATUS]
  patch  <id> <json_body>
  close  <id> <closed|rejected> [final_notes]
Env: PESA_BASE_URL (default $BASE_URL), PESA_COOKIE_JAR (default $JAR)
USAGE
}

curl_jar() {
  curl --silent --show-error --fail \
    --cookie "$JAR" --cookie-jar "$JAR" "$@"
}

case "${1:-}" in
  login)
    [ "$#" -eq 3 ] || { usage; exit 2; }
    curl_jar -X POST "$BASE_URL/api/v1/auth/login" \
      -H 'Content-Type: application/json' \
      -d "{\"email\":\"$2\",\"password\":\"$3\"}"
    echo "logged in; cookie jar at $JAR"
    ;;
  create)
    [ "$#" -ge 3 ] || { usage; exit 2; }
    NOTES="${4:-}"
    BODY=$(printf '{"kind":"%s","requester_email":"%s","notes":%s}' \
      "$2" "$3" "$(printf '%s' "$NOTES" | python3 -c 'import json,sys;print(json.dumps(sys.stdin.read()))')")
    curl_jar -X POST "$BASE_URL/admin/subject-rights/tickets" \
      -H 'Content-Type: application/json' -d "$BODY"
    ;;
  list)
    shift
    Q=""
    while [ "$#" -gt 0 ]; do
      case "$1" in
        --overdue) Q="${Q:+$Q&}overdue=true" ;;
        --status)  Q="${Q:+$Q&}status=$2"; shift ;;
        *) echo "unknown arg: $1" >&2; exit 2 ;;
      esac
      shift
    done
    URL="$BASE_URL/admin/subject-rights/tickets"
    [ -n "$Q" ] && URL="$URL?$Q"
    curl_jar -X GET "$URL"
    ;;
  patch)
    [ "$#" -eq 3 ] || { usage; exit 2; }
    curl_jar -X PATCH "$BASE_URL/admin/subject-rights/tickets/$2" \
      -H 'Content-Type: application/json' -d "$3"
    ;;
  close)
    [ "$#" -ge 3 ] || { usage; exit 2; }
    FINAL="${4:-}"
    BODY=$(printf '{"terminal":"%s","final_notes":%s}' "$3" \
      "$(printf '%s' "$FINAL" | python3 -c 'import json,sys;print(json.dumps(sys.stdin.read()))')")
    curl_jar -X POST "$BASE_URL/admin/subject-rights/tickets/$2/close" \
      -H 'Content-Type: application/json' -d "$BODY"
    ;;
  ""|-h|--help)
    usage ;;
  *)
    echo "unknown subcommand: $1" >&2
    usage; exit 2 ;;
esac
