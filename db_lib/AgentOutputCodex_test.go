package db_lib

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/agentlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// codexOutput is the reader of a Codex run with a clock that moves a second
// forward every time it is read.
func codexOutput(secrets ...string) AgentOutput {
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	return &codexJSONOutput{
		mask: newSecretMasker(secrets),
		now: func() time.Time {
			now = now.Add(time.Second)
			return now
		},
	}
}

func TestAgentLaunch_NewOutput_Codex(t *testing.T) {
	l, err := NewAgentLaunch(db.AppCodex, "/home/agent", agentInventory().McpServers)
	require.NoError(t, err)

	logged := feed(l.NewOutput(),
		`{"type":"item.completed","item":{"id":"item_0","type":"agent_message","text":"The token is `+agentTestToken+`"}}`)

	assert.Equal(t, []agentlog.Event{{Kind: agentlog.KindText, Text: "The token is " + maskedSecret}}, events(t, logged))
}

// The events of a real run (codex-cli 0.167.0): a message, a command that
// takes time, a command that fails, the end of the turn.
func TestCodexJSONOutput_MessagesAndCommands(t *testing.T) {
	logged := feed(codexOutput(),
		`{"type":"thread.started","thread_id":"01a1cc9f-c80c-7e00-9186-4ff99469e3fe"}`,
		`{"type":"turn.started"}`,
		`{"type":"item.completed","item":{"id":"item_0","type":"agent_message","text":"I’ll follow those steps.\n\nFirst the commands.\n"}}`,
		`{"type":"item.started","item":{"id":"item_1","type":"command_execution","command":"/bin/zsh -lc 'echo one; sleep 2; echo two'","aggregated_output":"","exit_code":null,"status":"in_progress"}}`,
		`{"type":"item.completed","item":{"id":"item_1","type":"command_execution","command":"/bin/zsh -lc 'echo one; sleep 2; echo two'","aggregated_output":"one\ntwo\n","exit_code":0,"status":"completed"}}`,
		`{"type":"item.started","item":{"id":"item_2","type":"command_execution","command":"/bin/zsh -lc 'cat /nope/missing'","aggregated_output":"","exit_code":null,"status":"in_progress"}}`,
		`{"type":"item.completed","item":{"id":"item_2","type":"command_execution","command":"/bin/zsh -lc 'cat /nope/missing'","aggregated_output":"cat: /nope/missing: No such file or directory\n","exit_code":1,"status":"failed"}}`,
		`{"type":"item.completed","item":{"id":"item_3","type":"agent_message","text":"Done."}}`,
		`{"type":"turn.completed","usage":{"input_tokens":91709,"cached_input_tokens":85888,"cache_write_input_tokens":0,"output_tokens":191,"reasoning_output_tokens":0}}`,
	)

	assert.Equal(t, []agentlog.Event{
		{Kind: agentlog.KindText, Text: "I’ll follow those steps."},
		{Kind: agentlog.KindText},
		{Kind: agentlog.KindText, Text: "First the commands."},
		{Kind: agentlog.KindTool, ID: "item_1", Name: "exec", Input: json.RawMessage(`{"command":"/bin/zsh -lc 'echo one; sleep 2; echo two'"}`)},
		{Kind: agentlog.KindResult, ID: "item_1", Text: "one\ntwo"},
		{Kind: agentlog.KindTool, ID: "item_2", Name: "exec", Input: json.RawMessage(`{"command":"/bin/zsh -lc 'cat /nope/missing'"}`)},
		{Kind: agentlog.KindResult, ID: "item_2", Text: "Exit code 1\ncat: /nope/missing: No such file or directory", Error: true},
		{Kind: agentlog.KindText, Text: "Done."},
		// The clock was read at the start and at the end of the turn.
		{Kind: agentlog.KindDone, DurationMs: 1000, InputTokens: 91709, OutputTokens: 191},
	}, events(t, logged))

	assert.Equal(t, []string{
		"I’ll follow those steps.",
		"",
		"First the commands.",
		`> exec {"command":"/bin/zsh -lc 'echo one; sleep 2; echo two'"}`,
		"  one",
		"  two",
		`> exec {"command":"/bin/zsh -lc 'cat /nope/missing'"}`,
		"  ! Exit code 1",
		"  ! cat: /nope/missing: No such file or directory",
		"Done.",
		"Agent finished: 1s, 91709 input / 191 output tokens",
	}, plain(logged))
}

// An MCP tool is named as Claude Code names it; its result is the text of
// the content, a failure the error or the content of the failed call.
func TestCodexJSONOutput_McpToolCalls(t *testing.T) {
	logged := feed(codexOutput(),
		`{"type":"item.started","item":{"id":"item_1","type":"mcp_tool_call","server":"tiny","tool":"list_issues","arguments":{"repo":"semaphoreui/semaphore","limit":3},"result":null,"error":null,"status":"in_progress"}}`,
		`{"type":"item.completed","item":{"id":"item_1","type":"mcp_tool_call","server":"tiny","tool":"list_issues","arguments":{"repo":"semaphoreui/semaphore","limit":3},"result":{"content":[{"type":"text","text":"{\"result\": [{\"number\": 4410, \"title\": \"Agent app\"}], \"total\": 1}"}],"structured_content":null},"error":null,"status":"completed"}}`,
		`{"type":"item.started","item":{"id":"item_2","type":"mcp_tool_call","server":"tiny","tool":"fail","arguments":{},"result":null,"error":null,"status":"in_progress"}}`,
		`{"type":"item.completed","item":{"id":"item_2","type":"mcp_tool_call","server":"tiny","tool":"fail","arguments":{},"result":{"content":[{"type":"text","text":"boom: the tool failed"},{"type":"image","data":"AAAA"}],"structured_content":null},"error":null,"status":"failed"}}`,
		`{"type":"item.completed","item":{"id":"item_3","type":"mcp_tool_call","server":"tiny","tool":"list_issues","arguments":{"repo":"x"},"result":null,"error":{"message":"MCP tool call requires approval, but approval policy is never"},"status":"failed"}}`,
		`{"type":"item.completed","item":{"id":"item_4","type":"mcp_tool_call","server":"tiny","tool":"count","result":{"content":[],"structured_content":{"total":2}},"error":null,"status":"completed"}}`,
	)

	assert.Equal(t, []agentlog.Event{
		{Kind: agentlog.KindTool, ID: "item_1", Name: "mcp__tiny__list_issues", Input: json.RawMessage(`{"repo":"semaphoreui/semaphore","limit":3}`)},
		{Kind: agentlog.KindResult, ID: "item_1", Text: `{"result":[{"number":4410,"title":"Agent app"}],"total":1}`},
		{Kind: agentlog.KindTool, ID: "item_2", Name: "mcp__tiny__fail"},
		{Kind: agentlog.KindResult, ID: "item_2", Text: "boom: the tool failed\n[image]", Error: true},
		// A call reported only when completed is logged then, with its result.
		{Kind: agentlog.KindTool, ID: "item_3", Name: "mcp__tiny__list_issues", Input: json.RawMessage(`{"repo":"x"}`)},
		{Kind: agentlog.KindResult, ID: "item_3", Text: "MCP tool call requires approval, but approval policy is never", Error: true},
		{Kind: agentlog.KindTool, ID: "item_4", Name: "mcp__tiny__count"},
		{Kind: agentlog.KindResult, ID: "item_4", Text: `{"total":2}`},
	}, events(t, logged))
}

// The other items of the CLI's schema (codex-rs exec_events.rs): a change of
// files, a web search, a plan, the reasoning, a call between agents.
func TestCodexJSONOutput_OtherItems(t *testing.T) {
	logged := feed(codexOutput(),
		`{"type":"item.completed","item":{"id":"item_1","type":"reasoning","text":"**Checking the tests**\nThey hang."}}`,
		`{"type":"item.started","item":{"id":"item_2","type":"todo_list","items":[{"text":"Read the test","completed":false},{"text":"Fix it","completed":false}]}}`,
		`{"type":"item.updated","item":{"id":"item_2","type":"todo_list","items":[{"text":"Read the test","completed":true},{"text":"Fix it","completed":false}]}}`,
		`{"type":"item.started","item":{"id":"item_3","type":"file_change","changes":[{"path":"a.txt","kind":"update"},{"path":"b.txt","kind":"add"}],"status":"in_progress"}}`,
		`{"type":"item.completed","item":{"id":"item_3","type":"file_change","changes":[{"path":"a.txt","kind":"update"},{"path":"b.txt","kind":"add"}],"status":"completed"}}`,
		`{"type":"item.completed","item":{"id":"item_4","type":"file_change","changes":[{"path":"c.txt","kind":"delete"}],"status":"failed"}}`,
		`{"type":"item.completed","item":{"id":"item_5","type":"web_search","id":"ws_1","query":"codex exec json","action":{"type":"search"}}}`,
		`{"type":"item.completed","item":{"id":"item_6","type":"collab_tool_call","tool":"spawn_agent","sender_thread_id":"t1","receiver_thread_ids":["t2"],"prompt":"Review it","agents_states":{},"status":"completed"}}`,
		`{"type":"item.completed","item":{"id":"item_2","type":"todo_list","items":[{"text":"Read the test","completed":true},{"text":"Fix it","completed":true}]}}`,
		`{"type":"item.completed","item":{"id":"item_7","type":"something_new","text":"?"}}`,
		`{"type":"item.completed"}`,
		`{"type":"something.new"}`,
	)

	assert.Equal(t, []agentlog.Event{
		{Kind: agentlog.KindThinking, Text: "**Checking the tests**"},
		{Kind: agentlog.KindThinking, Text: "They hang."},
		{Kind: agentlog.KindTool, ID: "item_2", Name: "todo_list", Input: json.RawMessage(`{"progress":"0/2","steps":["[ ] Read the test","[ ] Fix it"]}`)},
		// The plan is logged again as the agent ticks it off: a client keeps the last.
		{Kind: agentlog.KindTool, ID: "item_2", Name: "todo_list", Input: json.RawMessage(`{"progress":"1/2","steps":["[x] Read the test","[ ] Fix it"]}`)},
		{Kind: agentlog.KindTool, ID: "item_3", Name: "file_change", Input: json.RawMessage(`{"changes":[{"path":"a.txt","kind":"update"},{"path":"b.txt","kind":"add"}]}`)},
		{Kind: agentlog.KindResult, ID: "item_3"},
		{Kind: agentlog.KindTool, ID: "item_4", Name: "file_change", Input: json.RawMessage(`{"changes":[{"path":"c.txt","kind":"delete"}]}`)},
		{Kind: agentlog.KindResult, ID: "item_4", Text: "The change was not applied", Error: true},
		{Kind: agentlog.KindTool, ID: "ws_1", Name: "web_search", Input: json.RawMessage(`{"query":"codex exec json"}`)},
		{Kind: agentlog.KindResult, ID: "ws_1"},
		{Kind: agentlog.KindTool, ID: "item_6", Name: "spawn_agent", Input: json.RawMessage(`{"prompt":"Review it"}`)},
		{Kind: agentlog.KindResult, ID: "item_6"},
		{Kind: agentlog.KindTool, ID: "item_2", Name: "todo_list", Input: json.RawMessage(`{"progress":"2/2","steps":["[x] Read the test","[x] Fix it"]}`)},
		{Kind: agentlog.KindResult, ID: "item_2"},
	}, events(t, logged))
}

// An error is a plain line of the log; the failed turn does not repeat it.
func TestCodexJSONOutput_Errors(t *testing.T) {
	t.Run("an error that fails the turn", func(t *testing.T) {
		logged := feed(codexOutput(),
			`{"type":"turn.started"}`,
			`{"type":"error","message":"Reconnecting... 1/5"}`,
			`{"type":"item.completed","item":{"id":"item_1","type":"error","message":"The model is not available"}}`,
			`{"type":"error","message":"unexpected status 401 Unauthorized"}`,
			`{"type":"turn.failed","error":{"message":"unexpected status 401 Unauthorized"}}`,
		)

		assert.Equal(t, []string{
			"Error: Reconnecting... 1/5",
			"Error: The model is not available",
			"Error: unexpected status 401 Unauthorized",
			"Agent failed: 1s",
		}, plain(logged))
		assert.Equal(t, []agentlog.Event{{Kind: agentlog.KindDone, Error: true, DurationMs: 1000}}, events(t, logged[3:]))
	})

	t.Run("a turn that fails with an error not logged yet", func(t *testing.T) {
		logged := feed(codexOutput(), `{"type":"turn.failed","error":{"message":"stream closed"}}`)
		assert.Equal(t, []agentlog.Event{{Kind: agentlog.KindDone, Error: true, Errors: []string{"stream closed"}}}, events(t, logged))
	})
}

// A token is masked in the agent's text, in a command, in the arguments and
// the result of a tool, in an error and in the CLI's own log.
func TestCodexJSONOutput_MasksSecrets(t *testing.T) {
	const secret = `tok"<&>\x`

	quoted, err := json.Marshal(secret)
	require.NoError(t, err)
	inner, err := json.Marshal(map[string]string{"env": "TOKEN=" + secret})
	require.NoError(t, err)
	result, err := json.Marshal(string(inner))
	require.NoError(t, err)
	q := string(quoted)

	logged := feed(codexOutput(secret),
		`{"type":"item.completed","item":{"id":"item_0","type":"agent_message","text":`+q+`}}`,
		`{"type":"item.completed","item":{"id":"item_1","type":"reasoning","text":`+q+`}}`,
		`{"type":"item.completed","item":{"id":"item_2","type":"command_execution","command":`+q+`,"aggregated_output":`+q+`,"exit_code":0,"status":"completed"}}`,
		`{"type":"item.completed","item":{"id":"item_3","type":"mcp_tool_call","server":"s","tool":"t","arguments":{"auth":`+q+`},"result":{"content":[{"type":"text","text":`+string(result)+`}]},"status":"completed"}}`,
		`{"type":"error","message":`+q+`}`,
		`{"type":"turn.failed","error":{"message":"another `+secret[:3]+`\"<&>\\x"}}`,
		"2026-11-11T22:00:34Z ERROR rmcp: "+secret,
	)

	for _, line := range logged {
		assert.NotContains(t, line, secret)
		assert.NotContains(t, line, `tok\"<&>\\x`, "the secret as JSON escapes it")
	}

	assert.Equal(t, []string{
		maskedSecret,
		maskedSecret,
		`> exec {"command":"` + maskedSecret + `"}`,
		"  " + maskedSecret,
		`> mcp__s__t {"auth":"` + maskedSecret + `"}`,
		`  {"env":"TOKEN=` + maskedSecret + `"}`,
		"Error: " + maskedSecret,
		"another " + maskedSecret,
		"Agent failed",
		"2026-11-11T22:00:34Z ERROR rmcp: " + maskedSecret,
	}, plain(logged))
}

// What is not an event of the CLI stays in the log as it came.
func TestCodexJSONOutput_OtherOutput(t *testing.T) {
	lines := []string{
		"2026-11-11T22:00:34.119648Z ERROR rmcp::transport::worker: worker quit with fatal",
		"",
		`{"not":"an event"}`,
		`{broken`,
		"OpenAI Codex v0.167.0",
	}
	assert.Equal(t, lines, feed(codexOutput(), lines...))
}
