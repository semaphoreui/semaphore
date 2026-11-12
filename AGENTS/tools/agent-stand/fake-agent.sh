#!/bin/sh
# Stands in for claude / codex on the agent stand: shows what the agent would get, calls nothing.
# Started as Claude Code is (--output-format=stream-json) it then reports a made-up run in the
# events the real CLI prints — text in pieces, MCP and built-in tool calls, a failed call, a
# subagent working at the same time, two turns — so the task log can be checked in the browser.
# Started as Codex is (exec --json) it reports a made-up run in the events of codex exec instead
# (fake-codex-events.sh).
# FAKE_AGENT_QUIET (a variable of the template's Variable Group) leaves these lines out: a real
# CLI does not print them, and a screenshot for the docs must not show them.
if [ -n "${FAKE_AGENT_QUIET:-}" ]; then
  cat > /dev/null
else
  echo "fake agent args: $*"
  echo "prompt: $(cat)"
  echo "mcp token variables: $(env | grep -c '^SEMAPHORE_MCP_TOKEN_')"
  echo "first token: ${SEMAPHORE_MCP_TOKEN_1:-<none>}"
  echo "anthropic key set: $([ -n "${ANTHROPIC_API_KEY:-}" ] && echo yes || echo no)"
fi

case " $* " in
  *" --output-format=stream-json "*) ;;
  *" --json "*) exec "$(dirname "$0")/fake-codex-events.sh" ;;
  *) exit 0 ;;
esac

PAUSE=${FAKE_AGENT_PAUSE:-0.15}

ev() {
  printf '%s\n' "$1"
  sleep "$PAUSE"
}

delta() {
  ev '{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"'"$1"'"}},"parent_tool_use_id":null}'
}

INIT='{"type":"system","subtype":"init","cwd":"/repo","tools":["Bash","Read","Agent"],"model":"claude-sonnet-5-5","mcp_servers":[{"name":"github","status":"connected"},{"name":"docs","status":"failed"}]}'

ev '{"type":"system","subtype":"session_state_changed","state":"running"}'
ev "$INIT"
ev '{"type":"stream_event","event":{"type":"message_start","message":{"id":"msg_1"}},"parent_tool_use_id":null}'
delta 'I will look at the **open issues** first, and start a subagent for the flaky tests.\n'
ev '{"type":"stream_event","event":{"type":"content_block_stop","index":0},"parent_tool_use_id":null}'
ev '{"type":"assistant","message":{"id":"msg_1","content":[{"type":"text","text":"I will look at the **open issues** first, and start a subagent for the flaky tests.\n"}]},"parent_tool_use_id":null}'
ev '{"type":"assistant","message":{"id":"msg_1","content":[{"type":"tool_use","id":"toolu_1","name":"ToolSearch","input":{"query":"select:mcp__github__list_issues,mcp__github__get_issue","max_results":10}}]},"parent_tool_use_id":null}'
ev '{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_1","content":[{"type":"tool_reference","tool_name":"mcp__github__list_issues"},{"type":"tool_reference","tool_name":"mcp__github__get_issue"}]}]},"parent_tool_use_id":null}'
ev '{"type":"assistant","message":{"id":"msg_1","content":[{"type":"tool_use","id":"toolu_2","name":"mcp__github__list_issues","input":{"repo":"semaphoreui/semaphore","state":"open","labels":["bug"],"limit":3}}]},"parent_tool_use_id":null}'
ev '{"type":"assistant","message":{"id":"msg_1","content":[{"type":"tool_use","id":"toolu_3","name":"Agent","input":{"description":"Research area: flaky tests","prompt":"Find the tests that hang in pro_impl/services/tasks/k8s and say why.\nReport in five lines.","run_in_background":true}}]},"parent_tool_use_id":null}'
ev '{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_3","content":"Async agent launched successfully.\nagentId: a48438a722ad07eb4\nThe agent is working in the background."}]},"parent_tool_use_id":null}'
ev '{"type":"assistant","message":{"id":"msg_s1","content":[{"type":"tool_use","id":"toolu_4","name":"Read","input":{"file_path":"/repo/pro_impl/services/tasks/k8s/executor_test.go"}}]},"parent_tool_use_id":"toolu_3"}'
ev '{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_2","content":"{\"result\":[{\"number\":4410,\"title\":\"Agent app: run Claude Code and Codex\",\"labels\":[\"feature\",\"agents\"],\"assignee\":null},{\"number\":4431,\"title\":\"Remove process settings\",\"labels\":[],\"body\":\"The option is refused at startup.\\n\\nSee util/config_removed.go.\"}],\"total\":2}"}]},"parent_tool_use_id":null}'
ev '{"type":"stream_event","event":{"type":"message_start","message":{"id":"msg_2"}},"parent_tool_use_id":null}'
delta 'There are **2** open issues. The token of the server is '"${SEMAPHORE_MCP_TOKEN_1:-none}"', and `list_iss'
ev '{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_4","content":"func TestStreamPodLogsLive_FetchesInitContainerAndBuildContainer(t *testing.T) {\n\tcfg := newTestConfig()\n\t// ...\n}"}]},"parent_tool_use_id":"toolu_3"}'
delta 'ues` returned both.\n\n## Next\n\nNow the file that does not exist.\n'
ev '{"type":"stream_event","event":{"type":"content_block_stop","index":0},"parent_tool_use_id":null}'
ev '{"type":"assistant","message":{"id":"msg_s2","content":[{"type":"text","text":"The three TestStreamPodLogsLive tests hang in the fake clientset.\nNothing in the executor is at fault."}]},"parent_tool_use_id":"toolu_3"}'
ev '{"type":"assistant","message":{"id":"msg_2","content":[{"type":"tool_use","id":"toolu_5","name":"Bash","input":{"command":"cat /nope/missing","description":"Read a file that is not there"}}]},"parent_tool_use_id":null}'
ev '{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_5","content":"Exit code 1\ncat: /nope/missing: No such file or directory","is_error":true}]},"parent_tool_use_id":null}'
ev '{"type":"assistant","message":{"id":"msg_2","content":[{"type":"tool_use","id":"toolu_6","name":"mcp__github__get_issue","input":{"number":4410}}]},"parent_tool_use_id":null}'
ev '{"type":"system","subtype":"api_retry","attempt":1,"max_retries":10,"retry_delay_ms":500,"error":"overloaded"}'
ev '{"type":"result","subtype":"success","is_error":false,"result":"Now the file that does not exist.","num_turns":6,"duration_ms":19100,"total_cost_usd":1.5}'
ev "$INIT"
ev '{"type":"stream_event","event":{"type":"message_start","message":{"id":"msg_3"}},"parent_tool_use_id":null}'
delta 'The subagent is done: the hanging tests are not caused by the executor.\n'
ev '{"type":"stream_event","event":{"type":"content_block_stop","index":0},"parent_tool_use_id":null}'
ev '{"type":"result","subtype":"success","is_error":false,"result":"The subagent is done.","num_turns":2,"duration_ms":15500,"total_cost_usd":2.1784}'
ev '{"type":"system","subtype":"session_state_changed","state":"idle"}'
