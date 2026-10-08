#!/bin/bash
# Show what flowed through a workflow run on the stand: per task node, the outputs the producer
# stored (task.artifacts) and the "INPUT name=value" lines the consumer scripts print.
# usage: inputs.sh <workflow_id> [run_id]   (latest run of the workflow when run_id is omitted; PID=<project>)
set -euo pipefail
BASE=${BASE:-http://localhost:3100}
JAR=/tmp/semaphore-stand/cookies-inputs.txt
PID=${PID:-1}
WF=${1:?workflow id}; RUN=${2:-}
api() { curl -sf -b "$JAR" -c "$JAR" "$BASE/api$1"; }
rm -f "$JAR"; curl -sf -c "$JAR" -X POST -H 'Content-Type: application/json' -d '{"auth":"admin","password":"admin123"}' "$BASE/api/auth/login" > /dev/null
if [ -z "$RUN" ]; then
  RUN=$(api "/project/$PID/workflows/$WF/runs" | python3 -c 'import sys,json; r=json.load(sys.stdin); print(max(x["id"] for x in r) if r else "")')
  [ -n "$RUN" ] || { echo "workflow $WF has no runs" >&2; exit 1; }
fi
TPLS=$(api "/project/$PID/templates")
api "/project/$PID/workflows/$WF/runs/$RUN" | python3 -c '
import sys,json,re,subprocess
tpls={t["id"]:t["name"] for t in json.loads(sys.argv[1])}
base,pid,jar=sys.argv[2],sys.argv[3],sys.argv[4]
d=json.load(sys.stdin); r=d["run"]
print("run %d: %s (revision %s)"%(r["id"],r["status"],r.get("revision_id")))
ansi=re.compile(r"\x1b\[[0-9;]*m")
for n in d["nodes"]:
    node=n["node"]; t=n.get("task")
    if n.get("approval"): print("  node %d approval: %s"%(node["id"],n["approval"]["status"])); continue
    if not t: continue
    print("  node %d task %d [%s] %s runner=%s"%(node["id"],t["id"],t["status"],tpls.get(t["template_id"],"?"),t.get("used_runner_id") or "server"))
    a=t.get("artifacts")
    if a:
        doc=json.loads(a) if isinstance(a,str) else a
        print("    outputs:", json.dumps(doc.get("values",{}),ensure_ascii=False)[:300])
        if doc.get("skipped"): print("    skipped:", json.dumps(doc["skipped"]))
    out=subprocess.run(["curl","-sf","-b",jar,"%s/api/project/%s/tasks/%d/output"%(base,pid,t["id"])],capture_output=True,text=True).stdout
    for rec in json.loads(out or "[]"):
        line=ansi.sub("",rec["output"]).strip().strip(",").strip()
        if line.startswith("\"INPUT ") or line.startswith("INPUT "):
            # Ansible debug prints each message as a JSON string: drop its quotes and escapes.
            if line.startswith("\""):
                try: line=json.loads(line)
                except ValueError: line=line.strip("\"")
            print("    "+line)
' "$TPLS" "$BASE" "$PID" "$JAR"
