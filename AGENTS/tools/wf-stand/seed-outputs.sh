#!/bin/bash
# Seed the stand with a project "Artifacts QA", templates and workflows for testing workflow
# outputs (task.artifacts) by hand. Idempotent: reuses anything found by name.
# Scripts live in /tmp/semaphore-stand/wf-repo (see README). Prints the ids it created.
set -euo pipefail
BASE=${BASE:-http://localhost:3100}
JAR=/tmp/semaphore-stand/cookies.txt
REPO_PATH=/tmp/semaphore-stand/wf-repo
PROJECT_NAME=${PROJECT_NAME:-"Artifacts QA"}
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
REPO=$(api GET "/project/$PID/repositories" | find_by_name "wf-local")
[ -n "$REPO" ] || REPO=$(api POST "/project/$PID/repositories" "{\"project_id\":$PID,\"name\":\"wf-local\",\"git_url\":\"$REPO_PATH\",\"git_branch\":\"main\",\"ssh_key_id\":$KEY}" | jid)
INV=$(api GET "/project/$PID/inventory" | find_by_name "localhost")
[ -n "$INV" ] || INV=$(api POST "/project/$PID/inventory" "{\"project_id\":$PID,\"name\":\"localhost\",\"type\":\"static\",\"inventory\":\"localhost ansible_connection=local ansible_python_interpreter=$(command -v python3)\",\"ssh_key_id\":$KEY}" | jid)
echo "project $PID key $KEY repo $REPO inv $INV"

# mk_tpl NAME APP PLAYBOOK [SURVEY_VARS_JSON]
# Terraform-family templates get no inventory_id: the API creates a terraform-workspace inventory
# owned by the template (passing a shared one makes the API claim it and hide it from the list).
mk_tpl() { local id inv; id=$(api GET "/project/$PID/templates" | find_by_name "$1")
  [ -n "$id" ] && { echo "$id"; return; }
  # auto_approve: a terraform plan with changes otherwise parks the task in waiting_confirmation.
  case "$2" in terraform|tofu|terragrunt) inv=""; tp='{"auto_approve":true,"allow_auto_approve":true}';; *) inv="\"inventory_id\":$INV,"; tp='{}';; esac
  api POST "/project/$PID/templates" "{\"project_id\":$PID,\"name\":\"$1\",\"app\":\"$2\",\"playbook\":\"$3\",\"repository_id\":$REPO,${inv}\"environment_ids\":[],\"type\":\"\",\"task_params\":$tp,\"survey_vars\":${4:-[]}}" | jid; }

CONSUMER_VARS='[
 {"name":"image_tag","title":"Image tag","type":"","required":true},
 {"name":"replicas","title":"Replicas","type":"int","default_value":"1"},
 {"name":"enabled","title":"Enabled","type":"enum","values":[{"name":"yes","value":"true"},{"name":"no","value":"false"}]},
 {"name":"subnet_ids","title":"Subnets","type":"select","values":[{"name":"subnet-1","value":"subnet-1"},{"name":"subnet-2","value":"subnet-2"},{"name":"subnet-3","value":"subnet-3"}]},
 {"name":"config","title":"Config (text)","type":"text"},
 {"name":"env_name","title":"Env name (env target)","type":"","target":"env"},
 {"name":"token","title":"Secret (not mappable)","type":"secret"}]'

T_OK=$(mk_tpl "bash: ok" bash ok.sh)
T_FAIL=$(mk_tpl "bash: fail" bash fail.sh)
T_SLOW=$(mk_tpl "bash: slow 30s" bash slow.sh)
T_PROD=$(mk_tpl "bash: produce outputs" bash produce.sh '[{"name":"image_tag","title":"Image tag","type":"","default_value":"from-survey"}]')
T_EMPTY=$(mk_tpl "bash: produce empty" bash produce_empty.sh)
T_BADJSON=$(mk_tpl "bash: bad json (must fail)" bash produce_bad_json.sh)
T_BADNAME=$(mk_tpl "bash: bad name (must fail)" bash produce_bad_name.sh)
T_MANY=$(mk_tpl "bash: 101 outputs (must fail)" bash produce_too_many.sh)
T_LARGE=$(mk_tpl "bash: 40KB value (must fail)" bash produce_too_large.sh)
T_THENFAIL=$(mk_tpl "bash: outputs then exit 1" bash produce_then_fail.sh)
T_SYMLINK=$(mk_tpl "bash: symlink attack (must fail)" bash produce_symlink.sh)
T_CONS=$(mk_tpl "bash: consume" bash consume.sh "$CONSUMER_VARS")
T_TF=$(mk_tpl "terraform: outputs" terraform tf '[{"name":"image_tag","title":"Image tag","type":"","default_value":"tf-default"},{"name":"replicas","title":"Replicas","type":"int","default_value":"2"}]')
T_TOFU=$(mk_tpl "tofu: outputs" tofu tf)
T_ANS=$(mk_tpl "ansible: set_stats outputs" ansible playbooks/outputs.yml '[{"name":"image_tag","title":"Image tag","type":"","default_value":"ansible-default"}]')
T_ANS_CONS=$(mk_tpl "ansible: consume" ansible playbooks/consume.yml "$CONSUMER_VARS")
echo "templates ok=$T_OK fail=$T_FAIL slow=$T_SLOW produce=$T_PROD empty=$T_EMPTY badjson=$T_BADJSON badname=$T_BADNAME many=$T_MANY large=$T_LARGE thenfail=$T_THENFAIL symlink=$T_SYMLINK consume=$T_CONS tf=$T_TF tofu=$T_TOFU ansible=$T_ANS ansible_consume=$T_ANS_CONS"

mk_wf() { local id; id=$(api GET "/project/$PID/workflows" | find_by_name "$1")
  [ -n "$id" ] && { echo "$id"; return; }
  api POST "/project/$PID/workflows" "$2" | jid; }
node() { echo "{\"id\":$1,\"kind\":\"task\",\"template_id\":$2,\"task_params\":{${4:-}},\"position_x\":$3,\"position_y\":${5:-200}}"; }
edge() { echo "{\"source_node_id\":$1,\"destination_node_id\":$2,\"condition\":\"${3:-on_success}\"}"; }

# Static node values for the bash consumer (fallback when a mapped output is missing).
# Built outside the double-quoted body: bash mangles \" inside $( ) inside "...".
ENV_STATIC='"environment":"{\"image_tag\":\"static-from-node\",\"env_name\":\"static-env\"}"'

# 1. start -> three producers (bash, terraform, ansible) -> bash + ansible consumers.
#    A workflow must have exactly one root node, hence the "start" node.
W_MAIN=$(mk_wf "Outputs: producers -> consumers" "{\"project_id\":$PID,\"name\":\"Outputs: producers -> consumers\",\"nodes\":[
 $(node 9 $T_OK 40 '' 260),
 $(node 1 $T_PROD 360 '' 100), $(node 2 $T_TF 360 '' 260), $(node 3 $T_ANS 360 '' 420),
 $(node 4 $T_CONS 720 "$ENV_STATIC" 180),
 $(node 5 $T_ANS_CONS 720 '' 340)],
 \"edges\":[$(edge 9 1), $(edge 9 2), $(edge 9 3), $(edge 1 4), $(edge 2 4), $(edge 3 4), $(edge 1 5), $(edge 2 5), $(edge 3 5)]}")

# 2. Every invalid producer in parallel: each must FAIL at the producer; "empty" and "ok" succeed.
W_BAD=$(mk_wf "Outputs: invalid producers" "{\"project_id\":$PID,\"name\":\"Outputs: invalid producers\",\"nodes\":[
 $(node 1 $T_OK 80 '' 300),
 $(node 2 $T_EMPTY 420 '' 40), $(node 3 $T_BADJSON 420 '' 160), $(node 4 $T_BADNAME 420 '' 280),
 $(node 5 $T_MANY 420 '' 400), $(node 6 $T_LARGE 420 '' 520), $(node 7 $T_SYMLINK 420 '' 640)],
 \"edges\":[$(edge 1 2), $(edge 1 3), $(edge 1 4), $(edge 1 5), $(edge 1 6), $(edge 1 7)]}")

# 3. Failed producer: outputs written then exit 1 -> on_failure consumer must see no outputs (fallback).
W_FAILPROD=$(mk_wf "Outputs: failed producer" "{\"project_id\":$PID,\"name\":\"Outputs: failed producer\",\"nodes\":[
 $(node 1 $T_THENFAIL 80), $(node 2 $T_CONS 420 '' 100), $(node 3 $T_CONS 420 '' 300)],
 \"edges\":[$(edge 1 2 on_failure), $(edge 1 3 on_success)]}")

# 4. Tofu producer alone, then consumer (tofu binary path).
W_TOFU=$(mk_wf "Outputs: tofu -> consume" "{\"project_id\":$PID,\"name\":\"Outputs: tofu -> consume\",\"nodes\":[
 $(node 1 $T_TOFU 80), $(node 2 $T_CONS 420)],
 \"edges\":[$(edge 1 2)]}")

echo "workflows main=$W_MAIN invalid=$W_BAD failed_producer=$W_FAILPROD tofu=$W_TOFU"
echo "$PID $W_MAIN $W_BAD $W_FAILPROD $W_TOFU" > /tmp/semaphore-stand/wf-outputs-ids.txt
echo "UI: $BASE/project/$PID/workflows"
