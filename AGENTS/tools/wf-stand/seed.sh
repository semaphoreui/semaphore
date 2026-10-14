#!/bin/bash
# Seed the stand with a local bash repo, templates and two workflows whose outcome is known.
# Writes "<happy_id> <branches_id>" to /tmp/semaphore-stand/wf-exec-ids.txt. Idempotent.
set -euo pipefail
BASE=${BASE:-http://localhost:3100}
JAR=/tmp/semaphore-stand/cookies.txt
PID=${PID:-1}
REPO_PATH=/tmp/semaphore-stand/wf-repo
rm -f "$JAR"

api() { local m=$1 p=$2 d=${3:-}
  if [ -n "$d" ]; then curl -sf -b "$JAR" -c "$JAR" -X "$m" -H 'Content-Type: application/json' -d "$d" "$BASE/api$p"
  else curl -sf -b "$JAR" -c "$JAR" -X "$m" "$BASE/api$p"; fi; }
jid() { python3 -c 'import sys,json;print(json.load(sys.stdin)["id"])'; }
find_by_name() { python3 -c "import sys,json
name=sys.argv[1]; ids=[t['id'] for t in json.load(sys.stdin) if t['name']==name]; print(ids[0] if ids else '')" "$1"; }

api POST /auth/login '{"auth":"admin","password":"admin123"}' > /dev/null

# Local repository (path, no clone). Needs an access key of type none.
KEY=$(api GET "/project/$PID/keys" | find_by_name "wf-none")
[ -n "$KEY" ] || KEY=$(api POST "/project/$PID/keys" "{\"project_id\":$PID,\"name\":\"wf-none\",\"type\":\"none\"}" | jid)
REPO=$(api GET "/project/$PID/repositories" | find_by_name "wf-local")
[ -n "$REPO" ] || REPO=$(api POST "/project/$PID/repositories" "{\"project_id\":$PID,\"name\":\"wf-local\",\"git_url\":\"$REPO_PATH\",\"git_branch\":\"main\",\"ssh_key_id\":$KEY}" | jid)
INV=$(api GET "/project/$PID/inventory" | find_by_name "wf-local")
[ -n "$INV" ] || INV=$(api POST "/project/$PID/inventory" "{\"project_id\":$PID,\"name\":\"wf-local\",\"type\":\"static\",\"inventory\":\"localhost ansible_connection=local\",\"ssh_key_id\":$KEY}" | jid)

mk_tpl() { local id; id=$(api GET "/project/$PID/templates" | find_by_name "$1")
  [ -n "$id" ] && { echo "$id"; return; }
  api POST "/project/$PID/templates" "{\"project_id\":$PID,\"name\":\"$1\",\"app\":\"bash\",\"playbook\":\"$2\",\"repository_id\":$REPO,\"inventory_id\":$INV,\"environment_ids\":[],\"type\":\"\"}" | jid; }
T_OK=$(mk_tpl "wf ok" ok.sh)
T_FAIL=$(mk_tpl "wf fail" fail.sh)
T_SLOW=$(mk_tpl "wf slow" slow.sh)
echo "key $KEY repo $REPO inv $INV templates ok=$T_OK fail=$T_FAIL slow=$T_SLOW"

mk_wf() { local id; id=$(api GET "/project/$PID/workflows" | find_by_name "$1")
  [ -n "$id" ] && { echo "$id"; return; }
  api POST "/project/$PID/workflows" "$2" | jid; }

# Happy path: task -> approval -> delay 5s -> task. Expected: approval, then success.
W_HAPPY=$(mk_wf "Exec happy" "{\"project_id\":$PID,\"name\":\"Exec happy\",\"start_version\":\"0.1.0\",\"nodes\":[
 {\"id\":1,\"kind\":\"task\",\"template_id\":$T_OK,\"task_params\":{},\"position_x\":80,\"position_y\":200},
 {\"id\":2,\"kind\":\"approval\",\"approval_timeout\":600,\"approval_message\":\"Continue?\",\"position_x\":400,\"position_y\":200},
 {\"id\":3,\"kind\":\"delay\",\"delay_seconds\":5,\"position_x\":720,\"position_y\":200},
 {\"id\":4,\"kind\":\"task\",\"template_id\":$T_OK,\"task_params\":{},\"position_x\":1040,\"position_y\":200}],
 \"edges\":[{\"source_node_id\":1,\"destination_node_id\":2,\"condition\":\"on_success\"},
 {\"source_node_id\":2,\"destination_node_id\":3,\"condition\":\"on_success\"},
 {\"source_node_id\":3,\"destination_node_id\":4,\"condition\":\"on_success\"}]}")

# Branches: ok -> fail; fail --on_failure--> recover(ok), fail --on_success--> never(ok); ok --always--> slow.
# Expected: recover ran, never has no task, run ends failed (a failed task marks the run failed).
W_BR=$(mk_wf "Exec branches" "{\"project_id\":$PID,\"name\":\"Exec branches\",\"nodes\":[
 {\"id\":1,\"kind\":\"task\",\"template_id\":$T_OK,\"task_params\":{},\"position_x\":80,\"position_y\":200},
 {\"id\":2,\"kind\":\"task\",\"template_id\":$T_FAIL,\"task_params\":{},\"position_x\":400,\"position_y\":100},
 {\"id\":3,\"kind\":\"task\",\"template_id\":$T_OK,\"task_params\":{},\"position_x\":720,\"position_y\":40},
 {\"id\":4,\"kind\":\"task\",\"template_id\":$T_OK,\"task_params\":{},\"position_x\":720,\"position_y\":180},
 {\"id\":5,\"kind\":\"task\",\"template_id\":$T_SLOW,\"task_params\":{},\"position_x\":400,\"position_y\":340}],
 \"edges\":[{\"source_node_id\":1,\"destination_node_id\":2,\"condition\":\"on_success\"},
 {\"source_node_id\":2,\"destination_node_id\":3,\"condition\":\"on_failure\"},
 {\"source_node_id\":2,\"destination_node_id\":4,\"condition\":\"on_success\"},
 {\"source_node_id\":1,\"destination_node_id\":5,\"condition\":\"always\"}]}")
echo "workflows happy=$W_HAPPY branches=$W_BR"
echo "$W_HAPPY $W_BR" > /tmp/semaphore-stand/wf-exec-ids.txt
