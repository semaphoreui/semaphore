#!/bin/bash
# Seed the stand with the workflow "Survey: workflow variables": a workflow survey (enum, required
# string, int with default) and task nodes bound to it through survey_inputs. Needs the templates
# of seed-outputs.sh ("bash: ok", "bash: consume", "bash: produce renamed outputs") in the same
# project — run that first. Idempotent: reuses the workflow if it exists. Prints the ids.
#
# Nodes: 1 "bash: ok" (start) -> 2 "bash: consume" bound image_tag<-image_tag, replicas<-replicas,
# env_name<-env (plus static image_tag, to see the workflow value replace it); 3 "bash: consume"
# without bindings (gets nothing from the workflow although its variables share the names);
# 4 "bash: produce renamed outputs" -> 5 "bash: consume" over an explicit edge replicas<-count and
# bound image_tag<-image_tag (an edge mapping and a binding on different variables of one node).
set -euo pipefail
BASE=${BASE:-http://localhost:3100}
JAR=/tmp/semaphore-stand/cookies.txt
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
[ -n "$PID" ] || { echo "project \"$PROJECT_NAME\" not found — run seed-outputs.sh first" >&2; exit 1; }
tpl() { local id; id=$(api GET "/project/$PID/templates" | find_by_name "$1")
  [ -n "$id" ] || { echo "template \"$1\" not found — run seed-outputs.sh first" >&2; exit 1; }; echo "$id"; }
T_OK=$(tpl "bash: ok")
T_CONS=$(tpl "bash: consume")
T_PROD_RENAMED=$(tpl "bash: produce renamed outputs")
echo "project $PID templates ok=$T_OK consume=$T_CONS renamed=$T_PROD_RENAMED"

mk_wf() { local id; id=$(api GET "/project/$PID/workflows" | find_by_name "$1")
  [ -n "$id" ] && { echo "$id"; return; }
  api POST "/project/$PID/workflows" "$2" | jid; }
# node ID TPL X [TASK_PARAMS_BODY] [Y] [SURVEY_INPUTS_JSON]
node() { local si=""; [ -n "${6:-}" ] && si=",\"survey_inputs\":$6"
  echo "{\"id\":$1,\"kind\":\"task\",\"template_id\":$2,\"task_params\":{${4:-}},\"position_x\":$3,\"position_y\":${5:-200}$si}"; }
edge() { echo "{\"source_node_id\":$1,\"destination_node_id\":$2,\"condition\":\"${3:-on_success}\"}"; }
edgex() { local m="" p; for p in ${3:-}; do m="$m${m:+,}{\"var\":\"${p%%:*}\",\"key\":\"${p#*:}\"}"; done
  echo "{\"source_node_id\":$1,\"destination_node_id\":$2,\"condition\":\"on_success\",\"input_mode\":\"explicit\",\"input_mappings\":[$m]}"; }

ENV_STATIC='"environment":"{\"image_tag\":\"static-from-node\",\"env_name\":\"static-env\"}"'
SURVEY='[
 {"name":"env","title":"Environment","type":"enum","values":[{"name":"Production","value":"prod"},{"name":"Staging","value":"staging"}],"default_value":"staging"},
 {"name":"image_tag","title":"Image tag","type":"","required":true,"description":"Tag of the image to deploy"},
 {"name":"replicas","title":"Replicas","type":"int","default_value":"2"}]'
BIND_FULL='[{"var":"image_tag","workflow_var":"image_tag"},{"var":"replicas","workflow_var":"replicas"},{"var":"env_name","workflow_var":"env"}]'
BIND_TAG='[{"var":"image_tag","workflow_var":"image_tag"}]'

W_SURVEY=$(mk_wf "Survey: workflow variables" "{\"project_id\":$PID,\"name\":\"Survey: workflow variables\",\"start_version\":\"1.0.0\",\"survey_vars\":$SURVEY,\"nodes\":[
 $(node 1 $T_OK 40 '' 300),
 $(node 2 $T_CONS 420 "$ENV_STATIC" 100 "$BIND_FULL"),
 $(node 3 $T_CONS 420 "$ENV_STATIC" 300),
 $(node 4 $T_PROD_RENAMED 420 '' 500), $(node 5 $T_CONS 800 '' 500 "$BIND_TAG")],
 \"edges\":[$(edge 1 2), $(edge 1 3), $(edge 1 4), $(edgex 4 5 "replicas:count")]}")

echo "workflow survey=$W_SURVEY"
echo "$PID $W_SURVEY" > /tmp/semaphore-stand/wf-survey-ids.txt
echo "UI: $BASE/project/$PID/workflows/$W_SURVEY/edit"
echo "run: curl -b $JAR -X POST -H 'Content-Type: application/json' -d '{\"environment\":{\"image_tag\":\"1.2.3\",\"env\":\"prod\"}}' $BASE/api/project/$PID/workflows/$W_SURVEY/runs"
