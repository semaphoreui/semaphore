#!/bin/sh
# The made-up run fake-agent.sh reports when started as Codex is (codex exec --json): the events
# of the real CLI (codex-rs exec_events.rs) — the reasoning, a plan ticked off, an MCP tool call
# with a JSON result, a command that fails, a change of files, a line of the CLI's own log, an
# error, the end of the turn with its tokens. The token of the first MCP server is printed to
# check that the log masks it.
PAUSE=${FAKE_AGENT_PAUSE:-0.15}

ev() {
  printf '%s\n' "$1"
  sleep "$PAUSE"
}

ev '{"type":"thread.started","thread_id":"01a1cc9f-c80c-7e00-9186-4ff99469e3fe"}'
ev '{"type":"turn.started"}'
ev '{"type":"item.completed","item":{"id":"item_0","type":"reasoning","text":"**Planning the review**\nThe open issues first, then the commit."}}'
ev '{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"I will review the **last commit** and check the open issues with `list_issues`.\n"}}'
ev '{"type":"item.started","item":{"id":"item_2","type":"todo_list","items":[{"text":"List the open issues","completed":false},{"text":"Read the last commit","completed":false},{"text":"Note the risky changes","completed":false}]}}'
ev '{"type":"item.started","item":{"id":"item_3","type":"mcp_tool_call","server":"github","tool":"list_issues","arguments":{"repo":"semaphoreui/semaphore","state":"open","limit":3},"result":null,"error":null,"status":"in_progress"}}'
ev '2026-11-12T10:00:34.119648Z ERROR rmcp::transport::worker: worker quit with fatal: Transport channel closed (https://mcp.example.com/docs)'
ev '{"type":"item.completed","item":{"id":"item_3","type":"mcp_tool_call","server":"github","tool":"list_issues","arguments":{"repo":"semaphoreui/semaphore","state":"open","limit":3},"result":{"content":[{"type":"text","text":"{\"result\":[{\"number\":4410,\"title\":\"Agent app: run Claude Code and Codex\",\"labels\":[\"feature\",\"agents\"]},{\"number\":4431,\"title\":\"Remove process settings\",\"body\":\"The option is refused at startup.\\n\\nSee util/config_removed.go.\"}],\"total\":2}"}],"structured_content":null},"error":null,"status":"completed"}}'
ev '{"type":"item.updated","item":{"id":"item_2","type":"todo_list","items":[{"text":"List the open issues","completed":true},{"text":"Read the last commit","completed":false},{"text":"Note the risky changes","completed":false}]}}'
ev '{"type":"item.started","item":{"id":"item_4","type":"command_execution","command":"/bin/sh -lc '"'"'git log -1 --stat'"'"'","aggregated_output":"","exit_code":null,"status":"in_progress"}}'
ev '{"type":"item.completed","item":{"id":"item_4","type":"command_execution","command":"/bin/sh -lc '"'"'git log -1 --stat'"'"'","aggregated_output":"commit 7e50a373\nAuthor: stand\n\n    init\n\n services/api/app.py | 1 +\n 1 file changed, 1 insertion(+)\n","exit_code":0,"status":"completed"}}'
ev '{"type":"item.started","item":{"id":"item_5","type":"command_execution","command":"/bin/sh -lc '"'"'cat /nope/missing'"'"'","aggregated_output":"","exit_code":null,"status":"in_progress"}}'
ev '{"type":"item.completed","item":{"id":"item_5","type":"command_execution","command":"/bin/sh -lc '"'"'cat /nope/missing'"'"'","aggregated_output":"cat: /nope/missing: No such file or directory\n","exit_code":1,"status":"failed"}}'
ev '{"type":"item.started","item":{"id":"item_6","type":"file_change","changes":[{"path":"services/api/REVIEW.md","kind":"add"},{"path":"services/api/app.py","kind":"update"}],"status":"in_progress"}}'
ev '{"type":"item.completed","item":{"id":"item_6","type":"file_change","changes":[{"path":"services/api/REVIEW.md","kind":"add"},{"path":"services/api/app.py","kind":"update"}],"status":"completed"}}'
ev '{"type":"error","message":"Reconnecting... 1/5"}'
ev '{"type":"item.completed","item":{"id":"item_2","type":"todo_list","items":[{"text":"List the open issues","completed":true},{"text":"Read the last commit","completed":true},{"text":"Note the risky changes","completed":true}]}}'
ev '{"type":"item.started","item":{"id":"item_7","type":"mcp_tool_call","server":"github","tool":"get_issue","arguments":{"number":4410},"result":null,"error":null,"status":"in_progress"}}'
ev '{"type":"item.completed","item":{"id":"item_8","type":"agent_message","text":"There are **2** open issues; the commit is small. The token of the server is '"${SEMAPHORE_MCP_TOKEN_1:-none}"'.\n\n## Risky changes\n\nNone: only `app.py` changed."}}'
ev '{"type":"turn.completed","usage":{"input_tokens":91709,"cached_input_tokens":85888,"cache_write_input_tokens":0,"output_tokens":191,"reasoning_output_tokens":0}}'
