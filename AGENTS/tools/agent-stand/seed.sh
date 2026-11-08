#!/bin/bash
# Seeds the agent stand: project "Agents QA", a local git repository, a Variable Group with a fake
# ANTHROPIC_API_KEY, the mcp inventory "Agent MCP" (fake tokens) and the templates "Fix flaky
# tests" (claude, task prompt allowed) and "Review the last commit" (codex). Ids -> ids.txt.
set -euo pipefail
ROOT=$(cd "$(dirname "$0")/../../.." && pwd)
DIR=${AGENT_STAND_DIR:-$ROOT/AGENTS/tmp/agent-stand}
BASE=http://localhost:3102
JAR=$DIR/cookies.txt
mkdir -p "$DIR"
if [ ! -d "$DIR/repo/.git" ]; then
  mkdir -p "$DIR/repo/services/api"
  echo "# demo service" > "$DIR/repo/services/api/README.md"
  echo "print('hi')" > "$DIR/repo/services/api/app.py"
  git -C "$DIR/repo" init -q -b main
  git -C "$DIR/repo" add -A
  git -C "$DIR/repo" -c user.email=stand@example.com -c user.name=stand commit -q -m init
fi
api() { curl -sf -b "$JAR" -c "$JAR" -H 'Content-Type: application/json' "$@"; }
id() { python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])'; }
api -o /dev/null -X POST "$BASE/api/auth/login" -d '{"auth":"admin","password":"admin123"}'
P=$(api -X POST "$BASE/api/projects" -d '{"name":"Agents QA"}' | id)
K=$(api -X POST "$BASE/api/project/$P/keys" -d "{\"name\":\"none\",\"type\":\"none\",\"project_id\":$P}" | id)
R=$(api -X POST "$BASE/api/project/$P/repositories" -d "{\"name\":\"demo-service\",\"git_url\":\"$DIR/repo\",\"git_branch\":\"main\",\"ssh_key_id\":$K,\"project_id\":$P}" | id)
E=$(api -X POST "$BASE/api/project/$P/environment" -d "{\"name\":\"agent keys\",\"project_id\":$P,\"json\":\"{}\",\"env\":\"{}\",\"secrets\":[{\"name\":\"ANTHROPIC_API_KEY\",\"secret\":\"fake-anthropic-key\",\"type\":\"env\",\"operation\":\"create\"}]}" | id)
I=$(api -X POST "$BASE/api/project/$P/inventory" -d "{\"name\":\"Agent MCP\",\"type\":\"mcp\",\"project_id\":$P,\"mcp_servers\":[{\"name\":\"github\",\"url\":\"https://api.githubcopilot.com/mcp/\",\"enabled\":true,\"token\":{\"operation\":\"create\",\"secret\":\"fake-token-github\"}},{\"name\":\"docs\",\"url\":\"https://mcp.example.com/docs\",\"enabled\":true,\"token\":{\"operation\":\"create\",\"secret\":\"fake-token-docs\"}},{\"name\":\"search\",\"url\":\"https://mcp.example.com/search\",\"enabled\":false}]}" | id)
TC=$(api -X POST "$BASE/api/project/$P/templates" -d "{\"name\":\"Fix flaky tests\",\"app\":\"claude\",\"project_id\":$P,\"repository_id\":$R,\"inventory_id\":$I,\"environment_ids\":[$E],\"playbook\":\"services/api\",\"task_params\":{\"prompt\":\"Find the flaky tests in services/api, fix them and open a pull request.\",\"allow_override_prompt\":true}}" | id)
TX=$(api -X POST "$BASE/api/project/$P/templates" -d "{\"name\":\"Review the last commit\",\"app\":\"codex\",\"project_id\":$P,\"repository_id\":$R,\"inventory_id\":$I,\"environment_ids\":[$E],\"playbook\":\"\",\"task_params\":{\"prompt\":\"Review the last commit and list risky changes.\"}}" | id)
echo "project=$P key=$K repo=$R env=$E inventory=$I claude_template=$TC codex_template=$TX" | tee "$DIR/ids.txt"
