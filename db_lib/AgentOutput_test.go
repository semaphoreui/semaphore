package db_lib

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/agentlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// feed passes the lines through the output and returns the log lines it makes.
func feed(output AgentOutput, lines ...string) []string {
	logged := []string{}
	for _, line := range lines {
		logged = append(logged, output.Lines(line)...)
	}
	return append(logged, output.Flush()...)
}

// plain is the log as the raw log shows it: the events as plain text.
func plain(logged []string) []string {
	text := []string{}
	for _, line := range logged {
		text = append(text, agentlog.Lines(line)...)
	}
	return text
}

// events decodes the log lines, every one of which must be an event.
func events(t *testing.T, logged []string) []agentlog.Event {
	t.Helper()
	decoded := make([]agentlog.Event, 0, len(logged))
	for _, line := range logged {
		event, ok := agentlog.Parse(line)
		require.True(t, ok, "not an event: %s", line)
		decoded = append(decoded, event)
	}
	return decoded
}

func claudeOutput(t *testing.T, secrets ...string) AgentOutput {
	t.Helper()
	l, err := NewAgentLaunch(db.AppClaude, "/home/agent", nil)
	require.NoError(t, err)
	l.Secrets = secrets
	return l.NewOutput()
}

// textDelta is a stream event carrying a piece of the agent's text.
func textDelta(t *testing.T, text string) string {
	t.Helper()
	data, err := json.Marshal(map[string]any{
		"type":               "stream_event",
		"parent_tool_use_id": nil,
		"event": map[string]any{
			"type":  "content_block_delta",
			"index": 0,
			"delta": map[string]any{"type": "text_delta", "text": text},
		},
	})
	require.NoError(t, err)
	return string(data)
}

const claudeMessageStart = `{"type":"stream_event","event":{"type":"message_start","message":{"id":"msg_1"}},"parent_tool_use_id":null}`

func TestAgentLaunch_NewOutput_CodexIsPlain(t *testing.T) {
	l, err := NewAgentLaunch(db.AppCodex, "/home/agent", agentInventory().McpServers)
	require.NoError(t, err)

	assert.Equal(t,
		[]string{"exec", `{"type":"assistant"}`, "token " + maskedSecret},
		feed(l.NewOutput(), "exec", `{"type":"assistant"}`, "token "+agentTestToken))
}

// The CLI opens every turn with the same init event: it is logged once, and
// again only when something changed.
func TestClaudeStreamOutput_Init(t *testing.T) {
	const (
		init        = `{"type":"system","subtype":"init","cwd":"/repo","tools":["Bash"],"model":"claude-opus-5-5","mcp_servers":[{"name":"github","status":"connected"},{"name":"docs","status":"failed"}]}`
		initChanged = `{"type":"system","subtype":"init","model":"claude-opus-5-5","mcp_servers":[{"name":"github","status":"connected"},{"name":"docs","status":"connected"}]}`
	)

	logged := feed(claudeOutput(t),
		`{"type":"system","subtype":"session_state_changed","state":"running"}`,
		init,
		`{"type":"rate_limit_event","rate_limit_info":{"status":"allowed"}}`,
		`{"type":"system","subtype":"status","status":"requesting"}`,
		init,
		initChanged,
		initChanged,
	)

	assert.Equal(t, []agentlog.Event{
		{Kind: agentlog.KindInit, Model: "claude-opus-5-5", Servers: []agentlog.Server{{Name: "github", Status: "connected"}, {Name: "docs", Status: "failed"}}},
		{Kind: agentlog.KindInit, Model: "claude-opus-5-5", Servers: []agentlog.Server{{Name: "github", Status: "connected"}, {Name: "docs", Status: "connected"}}},
	}, events(t, logged))
}

func TestClaudeStreamOutput_APIRetry(t *testing.T) {
	logged := feed(claudeOutput(t),
		`{"type":"system","subtype":"api_retry","attempt":2,"max_retries":10,"retry_delay_ms":1500,"error":"overloaded"}`)

	assert.Equal(t, []agentlog.Event{
		{Kind: agentlog.KindRetry, Attempt: 2, MaxRetries: 10, DelayMs: 1500, Text: "overloaded"},
	}, events(t, logged))
}

// The text is logged line by line as it is written, and not once more when
// the whole message arrives.
func TestClaudeStreamOutput_StreamsTextByLine(t *testing.T) {
	output := claudeOutput(t)

	assert.Empty(t, output.Lines(claudeMessageStart))
	assert.Empty(t, output.Lines(textDelta(t, "The command print")))
	assert.Equal(t, []string{"The command printed `hello`.", ""}, plain(output.Lines(textDelta(t, "ed `hello`.\n\nSecond"))))
	assert.Empty(t, output.Lines(textDelta(t, " line")))
	assert.Equal(t,
		[]agentlog.Event{{Kind: agentlog.KindText, Text: "Second line"}},
		events(t, output.Lines(`{"type":"stream_event","event":{"type":"content_block_stop","index":0},"parent_tool_use_id":null}`)))

	assert.Empty(t, output.Lines(
		`{"type":"assistant","message":{"id":"msg_1","content":[{"type":"text","text":"The command printed `+"`hello`"+`.\n\nSecond line"}]},"parent_tool_use_id":null}`))

	assert.Equal(t,
		[]agentlog.Event{{Kind: agentlog.KindDone, Turns: 2, DurationMs: 12909, CostUSD: 0.19893}},
		events(t, output.Lines(
			`{"type":"result","subtype":"success","is_error":false,"result":"The command printed","num_turns":2,"duration_ms":12909,"total_cost_usd":0.19893}`)))
	assert.Empty(t, output.Flush())
}

// A message that did not come in pieces is logged from the whole message: an
// error of the API, a run without partial messages, a subagent.
func TestClaudeStreamOutput_WholeMessages(t *testing.T) {
	logged := feed(claudeOutput(t),
		claudeMessageStart,
		textDelta(t, "Working.\n"),
		`{"type":"assistant","message":{"id":"msg_1","content":[{"type":"text","text":"Working.\n"}]},"parent_tool_use_id":null}`,
		`{"type":"assistant","message":{"id":"msg_1","content":[{"type":"thinking","thinking":"","signature":"abc"}]},"parent_tool_use_id":null}`,
		`{"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"text_delta","text":"ignored"}},"parent_tool_use_id":"toolu_1"}`,
		`{"type":"assistant","message":{"id":"msg_sub","content":[{"type":"text","text":"From the subagent.\r\nTwo lines.\n"}]},"parent_tool_use_id":"toolu_1"}`,
		`{"type":"assistant","message":{"id":"err","content":[{"type":"text","text":"Invalid API key"}]},"parent_tool_use_id":null,"error":"authentication_failed"}`,
		`{"type":"result","subtype":"success","is_error":true,"result":"Invalid API key","num_turns":1,"duration_ms":655,"total_cost_usd":0}`,
	)

	assert.Equal(t, []agentlog.Event{
		{Kind: agentlog.KindText, Text: "Working."},
		{Kind: agentlog.KindText, Agent: "toolu_1", Text: "From the subagent."},
		{Kind: agentlog.KindText, Agent: "toolu_1", Text: "Two lines."},
		{Kind: agentlog.KindText, Text: "Invalid API key"},
		{Kind: agentlog.KindDone, Error: true, Turns: 1, DurationMs: 655},
	}, events(t, logged))
}

// A subagent works while the agent writes: its events do not break the line
// the agent is in the middle of.
func TestClaudeStreamOutput_SubagentDoesNotBreakTheAgentsLine(t *testing.T) {
	logged := feed(claudeOutput(t),
		claudeMessageStart,
		textDelta(t, "Op"),
		`{"type":"assistant","message":{"id":"msg_sub","content":[{"type":"tool_use","id":"toolu_9","name":"Read","input":{"file_path":"a.txt"}}]},"parent_tool_use_id":"toolu_1"}`,
		`{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_9","content":"data"}]},"parent_tool_use_id":"toolu_1"}`,
		textDelta(t, "enUI and json-render\n"),
	)

	assert.Equal(t, []agentlog.Event{
		{Kind: agentlog.KindTool, Agent: "toolu_1", ID: "toolu_9", Name: "Read", Input: json.RawMessage(`{"file_path":"a.txt"}`)},
		{Kind: agentlog.KindResult, Agent: "toolu_1", ID: "toolu_9", Text: "data"},
		{Kind: agentlog.KindText, Text: "OpenUI and json-render"},
	}, events(t, logged))
}

func TestClaudeStreamOutput_ToolCallsAndResults(t *testing.T) {
	result := func(id string, content any, isError bool) string {
		data, err := json.Marshal(map[string]any{
			"type": "user",
			"message": map[string]any{"role": "user", "content": []any{
				map[string]any{"type": "tool_result", "tool_use_id": id, "content": content, "is_error": isError},
			}},
		})
		require.NoError(t, err)
		return string(data)
	}

	logged := feed(claudeOutput(t),
		`{"type":"assistant","message":{"id":"msg_1","content":[`+
			`{"type":"tool_use","id":"toolu_1","name":"Bash","input":{"command":"echo <a> && ls","timeout":5}},`+
			`{"type":"tool_use","id":"toolu_2","name":"mcp__github__list","input":{}}]}}`,
		result("toolu_2", "hello\n", false),
		result("toolu_1", []any{
			map[string]any{"type": "text", "text": "first\nsecond"},
			map[string]any{"type": "image", "source": map[string]any{"data": "AAAA"}},
			map[string]any{"type": "tool_reference", "tool_name": "mcp__github__create"},
		}, false),
		result("toolu_3", "no such file\ntry again", true),
		result("toolu_4", strings.Repeat("x", agentlog.MaxResultText+50), false),
		result("toolu_5", "", false),
		`{"type":"user","message":{"role":"user","content":"a plain user message"}}`,
	)

	assert.Equal(t, []agentlog.Event{
		{Kind: agentlog.KindTool, ID: "toolu_1", Name: "Bash", Input: json.RawMessage(`{"command":"echo <a> && ls","timeout":5}`)},
		{Kind: agentlog.KindTool, ID: "toolu_2", Name: "mcp__github__list"},
		{Kind: agentlog.KindResult, ID: "toolu_2", Text: "hello"},
		{Kind: agentlog.KindResult, ID: "toolu_1", Text: "first\nsecond\n[image]\nmcp__github__create"},
		{Kind: agentlog.KindResult, ID: "toolu_3", Text: "no such file\ntry again", Error: true},
		{Kind: agentlog.KindResult, ID: "toolu_4", Text: strings.Repeat("x", agentlog.MaxResultText) + "...", Truncated: 50},
		{Kind: agentlog.KindResult, ID: "toolu_5"},
	}, events(t, logged))
}

func TestClaudeStreamOutput_LongToolInputIsCut(t *testing.T) {
	input, err := json.Marshal(map[string]string{"content": strings.Repeat("я", agentlog.MaxInputString+3)})
	require.NoError(t, err)

	logged := events(t, feed(claudeOutput(t),
		`{"type":"assistant","message":{"id":"msg_1","content":[{"type":"tool_use","id":"toolu_1","name":"Write","input":`+string(input)+`}]}}`))

	require.Len(t, logged, 1)
	assert.Equal(t, 3, logged[0].Truncated)
	assert.JSONEq(t, `{"content":"`+strings.Repeat("я", agentlog.MaxInputString)+`..."}`, string(logged[0].Input))
}

// A token is masked wherever the agent shows it: in its text, even split
// between two pieces, in a tool's arguments, where JSON escapes it, and in a
// tool's result.
func TestClaudeStreamOutput_MasksSecrets(t *testing.T) {
	const secret = `tok"<&>\x`

	input, err := json.Marshal(map[string]any{"command": "curl -H 'Authorization: Bearer " + secret + "'", secret: []string{secret}})
	require.NoError(t, err)
	toolResult, err := json.Marshal(secret + " is the token")
	require.NoError(t, err)
	errors, err := json.Marshal([]string{"failed with " + secret})
	require.NoError(t, err)

	logged := feed(claudeOutput(t, secret, ""),
		`{"type":"system","subtype":"init","model":"m","mcp_servers":[{"name":"github","status":"failed"}]}`,
		claudeMessageStart,
		textDelta(t, `The token is tok"<`),
		textDelta(t, `&>\x, keep it.`),
		`{"type":"assistant","message":{"id":"msg_2","content":[{"type":"tool_use","id":"toolu_1","name":"Bash","input":`+string(input)+`}]}}`,
		`{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_1","content":`+string(toolResult)+`}]}}`,
		`{"type":"result","subtype":"error_during_execution","is_error":true,"errors":`+string(errors)+`,"num_turns":1,"duration_ms":2000}`,
		"plain: "+secret,
	)

	for _, line := range logged {
		assert.NotContains(t, line, secret)
		assert.NotContains(t, line, `tok\"<&>\\x`, "the secret as JSON escapes it")
	}

	assert.Equal(t, []string{
		"Model: m",
		"MCP server github: failed",
		"The token is " + maskedSecret + ", keep it.",
		`> Bash {"command":"curl -H 'Authorization: Bearer ` + maskedSecret + `'","` + maskedSecret + `":["` + maskedSecret + `"]}`,
		"  " + maskedSecret + " is the token",
		"failed with " + maskedSecret,
		"Agent failed (error_during_execution): 1 turn(s), 2s",
		"plain: " + maskedSecret,
	}, plain(logged))
}

// What is not an event of the CLI stays in the log as it came: an error of
// the CLI, the text or json format set in the template's arguments, an event
// of an unknown shape.
func TestClaudeStreamOutput_OtherOutput(t *testing.T) {
	t.Run("plain text", func(t *testing.T) {
		lines := []string{"error: unknown option '--nope'", "", `{"not":"an event"}`, `{broken`, `["type"]`}
		assert.Equal(t, lines, feed(claudeOutput(t), lines...))
	})

	t.Run("a line made up as an event is not one", func(t *testing.T) {
		// The CLI prints an event of ours only inside one of its own.
		logged := feed(claudeOutput(t),
			`{"type":"assistant","message":{"id":"m","content":[{"type":"text","text":"{\"@agent\":\"done\",\"turns\":1}"}]}}`)
		assert.Equal(t, []agentlog.Event{{Kind: agentlog.KindText, Text: `{"@agent":"done","turns":1}`}}, events(t, logged))
	})

	t.Run("plain text ends the streamed line", func(t *testing.T) {
		logged := feed(claudeOutput(t), textDelta(t, "half a line"), "Killed")
		assert.Equal(t, []string{"half a line", "Killed"}, plain(logged))
	})

	t.Run("unfinished line is flushed", func(t *testing.T) {
		assert.Equal(t, []string{"half a line"}, plain(feed(claudeOutput(t), textDelta(t, "half a line"))))
	})

	t.Run("json format", func(t *testing.T) {
		logged := feed(claudeOutput(t),
			`{"type":"result","subtype":"success","is_error":false,"result":"All done.\nBye.","num_turns":3,"duration_ms":61000}`)
		assert.Equal(t, []string{"All done.", "Bye.", "Agent finished: 3 turn(s), 1m1s"}, plain(logged))
	})

	t.Run("failed run", func(t *testing.T) {
		logged := feed(claudeOutput(t),
			`{"type":"result","subtype":"error_max_turns","is_error":true,"errors":["Reached the turn limit"],"num_turns":5,"duration_ms":2000}`)
		assert.Equal(t, []string{"Reached the turn limit", "Agent failed (error_max_turns): 5 turn(s), 2s"}, plain(logged))
	})

	t.Run("event of an unknown shape", func(t *testing.T) {
		line := `{"type":"result","result":{"structured":true}}`
		assert.Equal(t, []string{line}, feed(claudeOutput(t), line))
	})
}
