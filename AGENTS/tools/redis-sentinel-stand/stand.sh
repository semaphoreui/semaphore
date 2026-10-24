#!/bin/bash
# Bring up / inspect / break the Redis Sentinel HA stand (docker compose in this folder).
# usage: stand.sh {build|up|down|status|master|failover|restore|logs [service] [n]}
#   SEMAPHORE_HA_ENGINE=redis SEMAPHORE_HA_REDIS_ADDR=redis-master:6379 stand.sh up   # compatibility run
set -euo pipefail
HERE=$(cd "$(dirname "$0")" && pwd)
ROOT=$(cd "$HERE/../../.." && pwd)
ENV_FILE=$HERE/.env
A=http://localhost:3110; B=http://localhost:3111
dc() { docker compose --project-directory "$HERE" "$@"; }
login() { # login HOST JAR
  curl -sf -c "$2" -X POST -H 'Content-Type: application/json' -d '{"auth":"admin","password":"admin123"}' "$1/api/auth/login" >/dev/null; }
cluster() { # cluster HOST
  local jar; jar=$(mktemp); login "$1" "$jar" || { echo "  $1: login failed"; rm -f "$jar"; return; }
  curl -sf -b "$jar" "$1/api/cluster" | python3 -c '
import sys,json; d=json.load(sys.stdin)
r=d.get("redis") or {}
print("  nodes:", ", ".join("%s%s(%s)"%(n["node_id"],"*" if n.get("is_self") else "","alive" if n.get("alive") else "dead") for n in d.get("nodes",[])))
print("  redis:", r.get("addr"), "connected=%s"%r.get("connected"), "version=%s"%r.get("version"))' 2>/dev/null || echo "  $1: /api/cluster failed"
  rm -f "$jar"; }
master() { dc exec -T sentinel redis-cli -p 26379 sentinel get-master-addr-by-name mymaster | paste -sd: -; }
ensure_env() {
  touch "$ENV_FILE"
  for k in SEMAPHORE_COOKIE_HASH SEMAPHORE_COOKIE_ENCRYPTION SEMAPHORE_ACCESS_KEY_ENCRYPTION; do
    grep -q "^$k=.\+" "$ENV_FILE" || { sed -i '' "/^$k=/d" "$ENV_FILE"; echo "$k=$(head -c32 /dev/urandom | base64)" >> "$ENV_FILE"; }
  done
  grep -q "^SEMAPHORE_SUBSCRIPTION_KEY=.\+" "$ENV_FILE" || echo "warning: SEMAPHORE_SUBSCRIPTION_KEY is empty in $ENV_FILE (Enterprise features off)" >&2
}
case "${1:-status}" in
  build)
    echo "cross-compiling the Pro binary (go.work) for linux/$(uname -m | sed 's/x86_64/amd64/')..."
    (cd "$ROOT" && GOOS=linux GOARCH=$(uname -m | sed 's/x86_64/amd64/') CGO_ENABLED=0 go build -tags netgo -o "$HERE/bin/semaphore" ./cli)
    dc build --quiet semaphore-a
    echo "built $HERE/bin/semaphore and image semaphore-sentinel-stand:local" ;;
  up)
    ensure_env
    [ -x "$HERE/bin/semaphore" ] || "$0" build
    dc up -d --wait --wait-timeout 180
    echo "stand: A=$A B=$B master=$(master)"
    "$0" status ;;
  down) dc down -v --remove-orphans; echo "stand: down" ;;
  status)
    echo "sentinel master: $(master 2>/dev/null || echo '?')"
    for h in "$A" "$B"; do echo "$h"; cluster "$h"; done ;;
  master) master ;;
  failover)
    before=$(master); echo "$(date -u +%T) master before: $before"
    dc kill redis-master >/dev/null; echo "$(date -u +%T) redis-master killed"
    for _ in $(seq 1 60); do
      now=$(master 2>/dev/null || true)
      if [ -n "$now" ] && [ "$now" != "$before" ]; then echo "$(date -u +%T) master after:  $now"; exit 0; fi
      sleep 0.5
    done
    echo "no failover within 30 s" >&2; exit 1 ;;
  restore) dc start redis-master; sleep 3; echo "redis-master started; role: $(dc exec -T redis-master redis-cli role | head -1); master=$(master)" ;;
  logs)
    if [ -n "${2:-}" ]; then
      dc logs --no-color --tail "${3:-50}" "$2"
    else
      dc logs --no-color --tail "${3:-50}"
    fi ;;
  *) echo "usage: $0 {build|up|down|status|master|failover|restore|logs [service] [n]}" >&2; exit 2 ;;
esac
