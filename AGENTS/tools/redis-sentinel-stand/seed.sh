#!/bin/bash
# Seed the stand through node A: project "Sentinel QA", key, repo (/tmp/repo inside the
# containers), inventory and the bash templates "ok" and "slow". Idempotent by name; writes ids.env.
set -euo pipefail
HERE=$(cd "$(dirname "$0")" && pwd)
BASE=${BASE:-http://localhost:3110}
JAR=$HERE/cookies-seed.txt
api() { local m=$1 p=$2 d=${3:-}
  if [ -n "$d" ]; then curl -sf -b "$JAR" -c "$JAR" -X "$m" -H 'Content-Type: application/json' -d "$d" "$BASE/api$p"
  else curl -sf -b "$JAR" -c "$JAR" -X "$m" "$BASE/api$p"; fi; }
jid() { python3 -c 'import sys,json;print(json.load(sys.stdin)["id"])'; }
by_name() { python3 -c "import sys,json
ids=[t['id'] for t in json.load(sys.stdin) if t['name']==sys.argv[1]]; print(ids[0] if ids else '')" "$1"; }
rm -f "$JAR"; api POST /auth/login '{"auth":"admin","password":"admin123"}' >/dev/null
PID=$(api GET /projects | by_name "Sentinel QA"); [ -n "$PID" ] || PID=$(api POST /projects '{"name":"Sentinel QA"}' | jid)
KEY=$(api GET "/project/$PID/keys" | by_name none); [ -n "$KEY" ] || KEY=$(api POST "/project/$PID/keys" "{\"project_id\":$PID,\"name\":\"none\",\"type\":\"none\"}" | jid)
REPO=$(api GET "/project/$PID/repositories" | by_name scripts); [ -n "$REPO" ] || REPO=$(api POST "/project/$PID/repositories" "{\"project_id\":$PID,\"name\":\"scripts\",\"git_url\":\"/tmp/repo\",\"git_branch\":\"main\",\"ssh_key_id\":$KEY}" | jid)
INV=$(api GET "/project/$PID/inventory" | by_name localhost); [ -n "$INV" ] || INV=$(api POST "/project/$PID/inventory" "{\"project_id\":$PID,\"name\":\"localhost\",\"type\":\"static\",\"inventory\":\"localhost ansible_connection=local\",\"ssh_key_id\":$KEY}" | jid)
mk_tpl() { local id; id=$(api GET "/project/$PID/templates" | by_name "$1"); [ -n "$id" ] && { echo "$id"; return; }
  api POST "/project/$PID/templates" "{\"project_id\":$PID,\"name\":\"$1\",\"app\":\"bash\",\"playbook\":\"$2\",\"repository_id\":$REPO,\"inventory_id\":$INV,\"environment_ids\":[],\"type\":\"\",\"task_params\":{}}" | jid; }
OK=$(mk_tpl ok ok.sh); SLOW=$(mk_tpl slow slow.sh)
printf 'PID=%s\nTPL_OK=%s\nTPL_SLOW=%s\n' "$PID" "$OK" "$SLOW" | tee "$HERE/ids.env"
