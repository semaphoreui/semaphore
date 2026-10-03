#!/bin/bash
# Seeds workflows on the screenshot stand through the REST API.
set -euo pipefail
BASE=http://localhost:3100
JAR=/tmp/semaphore-stand/cookies.txt
PID=1
rm -f "$JAR"

api() { # method path [json]
  local m=$1 p=$2 d=${3:-}
  if [ -n "$d" ]; then
    curl -s -b "$JAR" -c "$JAR" -X "$m" -H 'Content-Type: application/json' -d "$d" "$BASE/api$p"
  else
    curl -s -b "$JAR" -c "$JAR" -X "$m" "$BASE/api$p"
  fi
}
jid() { python3 -c 'import sys,json;print(json.load(sys.stdin)["id"])'; }

api POST /auth/login '{"auth":"admin","password":"admin123"}' > /dev/null

REPO=1
INV=1

find_tpl() { # name -> id or empty
  api GET "/project/$PID/templates" | python3 -c "import sys,json
name=sys.argv[1]
ids=[t['id'] for t in json.load(sys.stdin) if t['name']==name]
print(ids[0] if ids else '')" "$1"
}
mk_tpl() { # name app playbook (idempotent)
  local existing
  existing=$(find_tpl "$1")
  if [ -n "$existing" ]; then echo "$existing"; return; fi
  api POST "/project/$PID/templates" "{\"project_id\":$PID,\"name\":\"$1\",\"playbook\":\"$3\",\"app\":\"$2\",\"repository_id\":$REPO,\"inventory_id\":$INV,\"environment_ids\":[],\"type\":\"\"}" | jid
}

T_DEPLOY=$(mk_tpl "Deploy website" ansible site.yml)
T_TF=$(mk_tpl "Apply infrastructure changes for prod" terraform main.tf)
T_NOTIFY=$(mk_tpl "Notify on-call" bash notify.sh)
T_SMOKE=$(mk_tpl "Smoke tests" python smoke.py)
T_XSS=$(mk_tpl "<img src=x onerror=alert(1)>" ansible xss.yml)
echo "templates $T_DEPLOY $T_TF $T_NOTIFY $T_SMOKE $T_XSS"

# 1. Release workflow with positions (diamond + failure branch + note)
W1=$(api POST "/project/$PID/workflows" "{
  \"project_id\": $PID, \"name\": \"Release\", \"start_version\": \"1.0.0\",
  \"nodes\": [
    {\"id\": 1, \"kind\": \"task\", \"template_id\": $T_DEPLOY, \"convergence_mode\": \"all\", \"task_params\": {}, \"position_x\": 80, \"position_y\": 200},
    {\"id\": 2, \"kind\": \"approval\", \"convergence_mode\": \"all\", \"approval_timeout\": 3600, \"approval_message\": \"Deploy to prod?\", \"position_x\": 400, \"position_y\": 100},
    {\"id\": 3, \"kind\": \"delay\", \"convergence_mode\": \"all\", \"delay_seconds\": 120, \"position_x\": 720, \"position_y\": 100},
    {\"id\": 4, \"kind\": \"task\", \"template_id\": $T_TF, \"convergence_mode\": \"any\", \"task_params\": {}, \"position_x\": 1040, \"position_y\": 200},
    {\"id\": 5, \"kind\": \"task\", \"template_id\": $T_NOTIFY, \"convergence_mode\": \"all\", \"task_params\": {}, \"position_x\": 720, \"position_y\": 380},
    {\"id\": 6, \"kind\": \"note\", \"note\": \"Prod apply needs sign-off from the on-call lead. Delay gives the CDN time to purge.\", \"position_x\": 1040, \"position_y\": 340}
  ],
  \"edges\": [
    {\"source_node_id\": 1, \"destination_node_id\": 2, \"condition\": \"on_success\"},
    {\"source_node_id\": 2, \"destination_node_id\": 3, \"condition\": \"on_success\"},
    {\"source_node_id\": 3, \"destination_node_id\": 4, \"condition\": \"always\"},
    {\"source_node_id\": 1, \"destination_node_id\": 5, \"condition\": \"on_failure\"}
  ]
}" | jid)
echo "workflow Release $W1"

# 2. Legacy workflow without positions (auto layout) + hostile template name
W2=$(api POST "/project/$PID/workflows" "{
  \"project_id\": $PID, \"name\": \"Nightly checks\",
  \"nodes\": [
    {\"id\": 1, \"kind\": \"task\", \"template_id\": $T_SMOKE, \"convergence_mode\": \"all\", \"task_params\": {}},
    {\"id\": 2, \"kind\": \"task\", \"template_id\": $T_XSS, \"convergence_mode\": \"all\", \"task_params\": {}},
    {\"id\": 3, \"kind\": \"task\", \"template_id\": $T_NOTIFY, \"convergence_mode\": \"all\", \"task_params\": {}},
    {\"id\": 4, \"kind\": \"note\", \"note\": \"<script>alert(1)</script>\"}
  ],
  \"edges\": [
    {\"source_node_id\": 1, \"destination_node_id\": 2, \"condition\": \"on_success\"},
    {\"source_node_id\": 1, \"destination_node_id\": 3, \"condition\": \"on_failure\"}
  ]
}" | jid)
echo "workflow Nightly $W2"

# 3. Start a run of Release so the run view has live statuses.
RUN=$(api POST "/project/$PID/workflows/$W1/runs" '{}' )
echo "run: $RUN"
echo "$W1 $W2" > /tmp/semaphore-stand/wf-ids.txt
