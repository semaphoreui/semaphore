#!/usr/bin/env bash
# Runs the real codex/claude CLI the way AgentApp does (isolated home, MCP
# config with the token taken from an env var, prompt on stdin) against a
# local fake MCP server. Keys come from the git-ignored .env next to this file.
set -euo pipefail

app="${1:-}"
prompt="${2:-Reply with exactly one word: OK}"
port="${PROBE_PORT:-8765}"
dir="$(cd "$(dirname "$0")" && pwd)"

[[ "$app" == codex || "$app" == claude ]] || { echo "usage: $0 codex|claude [prompt]" >&2; exit 2; }
[[ -f "$dir/.env" ]] || { echo "missing $dir/.env (copy .env.example)" >&2; exit 2; }
set -a; source "$dir/.env"; set +a

work="$(mktemp -d -t agent-cli-probe)"
trap 'kill "${srv:-}" 2>/dev/null || true; rm -rf "$work"' EXIT
mkdir -p "$work/home" "$work/repo"

python3 -I "$dir/mock_mcp.py" "$port" "$work/hits.log" & srv=$!; disown
sleep 0.5

export SEMAPHORE_MCP_TOKEN_0="probe-token-$RANDOM"
status=0
cd "$work/repo"

if [[ "$app" == codex ]]; then
  [[ -n "${CODEX_API_KEY:-}" ]] || { echo "CODEX_API_KEY is empty in .env" >&2; exit 2; }
  cat > "$work/home/config.toml" <<TOML
[mcp_servers.probe]
url = "http://127.0.0.1:$port/mcp"
bearer_token_env_var = "SEMAPHORE_MCP_TOKEN_0"
default_tools_approval_mode = "auto"
TOML
  printf '%s' "$prompt" | env -u OPENAI_API_KEY CODEX_HOME="$work/home" HOME="$work/home" \
    codex exec --skip-git-repo-check - || status=$?
else
  [[ -n "${ANTHROPIC_API_KEY:-}" ]] || { echo "ANTHROPIC_API_KEY is empty in .env" >&2; exit 2; }
  cat > "$work/mcp.json" <<JSON
{"mcpServers":{"probe":{"type":"http","url":"http://127.0.0.1:$port/mcp","headers":{"Authorization":"Bearer \${SEMAPHORE_MCP_TOKEN_0}"}}}}
JSON
  printf '%s' "$prompt" | CLAUDE_CONFIG_DIR="$work/home" \
    claude -p --mcp-config="$work/mcp.json" --strict-mcp-config --no-session-persistence \
      --allowedTools='mcp__probe__*' || status=$?
fi

echo "--- exit: $status"
echo "--- MCP requests:"
if [[ -s "$work/hits.log" ]]; then
  sed "s/$SEMAPHORE_MCP_TOKEN_0/<token ok>/" "$work/hits.log"
else
  echo "(none)"
fi
exit "$status"
