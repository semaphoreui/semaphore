#!/bin/bash
# Seed the stand with the project "Demo Stand" and the workflow "Build and deploy" used for the
# screenshots in docs/docs/user-guide/workflows.md (§ Outputs and inputs). Idempotent by name.
# Scripts: /tmp/semaphore-stand/wf-repo/{provision.sh,build.sh,playbooks/deploy.yml}.
# Prints "project=<id> workflow=<id>"; shoot with shoot-docs.cjs after one successful run.
set -euo pipefail
BASE=${BASE:-http://localhost:3100}
JAR=/tmp/semaphore-stand/cookies-docs.txt
REPO_PATH=/tmp/semaphore-stand/wf-repo
PROJECT_NAME=${PROJECT_NAME:-"Demo Stand"}
rm -f "$JAR"

api() { local m=$1 p=$2 d=${3:-}
  if [ -n "$d" ]; then curl -sf -b "$JAR" -c "$JAR" -X "$m" -H 'Content-Type: application/json' -d "$d" "$BASE/api$p"
  else curl -sf -b "$JAR" -c "$JAR" -X "$m" "$BASE/api$p"; fi; }
jid() { python3 -c 'import sys,json;print(json.load(sys.stdin)["id"])'; }
find_by_name() { python3 -c "import sys,json
name=sys.argv[1]; ids=[t['id'] for t in json.load(sys.stdin) if t['name']==name]; print(ids[0] if ids else '')" "$1"; }

api POST /auth/login '{"auth":"admin","password":"admin123"}' > /dev/null
PID=$(api GET /projects | find_by_name "$PROJECT_NAME")
[ -n "$PID" ] || PID=$(api POST /projects "{\"name\":\"$PROJECT_NAME\"}" | jid)
KEY=$(api GET "/project/$PID/keys" | find_by_name "none")
[ -n "$KEY" ] || KEY=$(api POST "/project/$PID/keys" "{\"project_id\":$PID,\"name\":\"none\",\"type\":\"none\"}" | jid)
REPO=$(api GET "/project/$PID/repositories" | find_by_name "shop")
[ -n "$REPO" ] || REPO=$(api POST "/project/$PID/repositories" "{\"project_id\":$PID,\"name\":\"shop\",\"git_url\":\"$REPO_PATH\",\"git_branch\":\"main\",\"ssh_key_id\":$KEY}" | jid)
INV=$(api GET "/project/$PID/inventory" | find_by_name "localhost")
[ -n "$INV" ] || INV=$(api POST "/project/$PID/inventory" "{\"project_id\":$PID,\"name\":\"localhost\",\"type\":\"static\",\"inventory\":\"localhost ansible_connection=local ansible_python_interpreter=$(command -v python3)\",\"ssh_key_id\":$KEY}" | jid)

# mk_tpl NAME APP PLAYBOOK [SURVEY_VARS_JSON]
mk_tpl() { local id; id=$(api GET "/project/$PID/templates" | find_by_name "$1")
  [ -n "$id" ] && { echo "$id"; return; }
  api POST "/project/$PID/templates" "{\"project_id\":$PID,\"name\":\"$1\",\"app\":\"$2\",\"playbook\":\"$3\",\"repository_id\":$REPO,\"inventory_id\":$INV,\"environment_ids\":[],\"type\":\"\",\"task_params\":{},\"survey_vars\":${4:-[]}}" | jid; }

DEPLOY_VARS='[
 {"name":"image_tag","title":"Image tag","type":"","required":true},
 {"name":"replicas","title":"Replicas","type":"int","default_value":"1"},
 {"name":"subnets","title":"Subnets","type":"select","values":[{"name":"subnet-1","value":"subnet-1"},{"name":"subnet-2","value":"subnet-2"},{"name":"subnet-3","value":"subnet-3"}]}]'
T_PROV=$(mk_tpl "Provision network" bash provision.sh)
T_BUILD=$(mk_tpl "Build image" bash build.sh)
T_DEPLOY=$(mk_tpl "Deploy to cluster" ansible playbooks/deploy.yml "$DEPLOY_VARS")

node() { echo "{\"id\":$1,\"kind\":\"task\",\"template_id\":$2,\"task_params\":{},\"position_x\":$3,\"position_y\":${4:-200}}"; }
# One root: Provision -> Build (by name, nothing to deliver) -> Deploy (by name: image_tag, replicas);
# the direct Provision -> Deploy connection maps subnet_ids to "subnets" explicitly.
WF=$(api GET "/project/$PID/workflows" | find_by_name "Build and deploy")
[ -n "$WF" ] || WF=$(api POST "/project/$PID/workflows" "{\"project_id\":$PID,\"name\":\"Build and deploy\",\"nodes\":[
 $(node 1 $T_PROV 40 220), $(node 2 $T_BUILD 340 90), $(node 3 $T_DEPLOY 640 220)],
 \"edges\":[{\"source_node_id\":1,\"destination_node_id\":2,\"condition\":\"on_success\"},
            {\"source_node_id\":2,\"destination_node_id\":3,\"condition\":\"on_success\"},
            {\"source_node_id\":1,\"destination_node_id\":3,\"condition\":\"on_success\",\"input_mode\":\"explicit\",\"input_mappings\":[{\"var\":\"subnets\",\"key\":\"subnet_ids\"}]}]}" | jid)
echo "project=$PID workflow=$WF templates=$T_PROV,$T_BUILD,$T_DEPLOY"
