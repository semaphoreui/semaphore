package db_lib

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// feed passes the lines through the output and returns everything it logs.
func feed(output AgentOutput, lines ...string) []string {
	logged := []string{}
	for _, line := range lines {
		logged = append(logged, output.Lines(line)...)
	}
	return append(logged, output.Flush()...)
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

func TestAgentLaunch_NewOutput_CodexIsPlain(t *testing.T) {
	l, err := NewAgentLaunch(db.AppCodex, "/home/agent", agentInventory().McpServers)
	require.NoError(t, err)

	assert.Equal(t,
		[]string{"exec", `{"type":"assistant"}`, "token " + maskedSecret},
		feed(l.NewOutput(), "exec", `{"type":"assistant"}`, "token "+agentTestToken))
}

func TestClaudeStreamOutput_Init(t *testing.T) {
	logged := feed(claudeOutput(t),
		`{"type":"system","subtype":"session_state_changed","state":"running"}`,
		`{"type":"system","subtype":"init","cwd":"/repo","tools":["Bash"],"model":"claude-opus-5-5",`+
			`"mcp_servers":[{"name":"github","status":"connected"},{"name":"docs","status":"failed"}]}`,
		`{"type":"rate_limit_event","rate_limit_info":{"status":"allowed"}}`,
		`{"type":"system","subtype":"status","status":"requesting"}`,
	)

	assert.Equal(t, []string{
		"Model: claude-opus-5-5",
		"MCP server github: connected",
		"MCP server docs: failed",
	}, logged)
}

func TestClaudeStreamOutput_APIRetry(t *testing.T) {
	logged := feed(claudeOutput(t),
		`{"type":"system","subtype":"api_retry","attempt":2,"max_retries":10,"retry_delay_ms":1500,"error":"overloaded"}`)

	assert.Equal(t, []string{"API request failed, retry 2 of 10 in 1.5s: overloaded"}, logged)
}

// The text is logged line by line as it is written, and not once more when
// the whole message arrives.
func TestClaudeStreamOutput_StreamsTextByLine(t *testing.T) {
	output := claudeOutput(t)

	assert.Empty(t, output.Lines(`{"type":"stream_event","event":{"type":"message_start","message":{"id":"msg_1"}},"parent_tool_use_id":null}`))
	assert.Empty(t, output.Lines(textDelta(t, "The command print")))
	assert.Equal(t, []string{"The command printed `hello`.", ""}, output.Lines(textDelta(t, "ed `hello`.\n\nSecond")))
	assert.Empty(t, output.Lines(textDelta(t, " line")))
	assert.Equal(t, []string{"Second line"},
		output.Lines(`{"type":"stream_event","event":{"type":"content_block_stop","index":0},"parent_tool_use_id":null}`))

	assert.Empty(t, output.Lines(
		`{"type":"assistant","message":{"id":"msg_1","content":[{"type":"text","text":"The command printed `+"`hello`"+`.\n\nSecond line"}]},"parent_tool_use_id":null}`))

	assert.Equal(t, []string{"Agent finished: 2 turn(s), 12.9s, $0.1989"}, output.Lines(
		`{"type":"result","subtype":"success","is_error":false,"result":"The command printed","num_turns":2,"duration_ms":12909,"total_cost_usd":0.19893}`))
	assert.Empty(t, output.Flush())
}

// A message that did not come in pieces is logged from the whole message: an
// error of the API, a run without partial messages, a subagent.
func TestClaudeStreamOutput_WholeMessages(t *testing.T) {
	logged := feed(claudeOutput(t),
		`{"type":"stream_event","event":{"type":"message_start","message":{"id":"msg_1"}},"parent_tool_use_id":null}`,
		textDelta(t, "Working.\n"),
		`{"type":"assistant","message":{"id":"msg_1","content":[{"type":"text","text":"Working.\n"}]},"parent_tool_use_id":null}`,
		`{"type":"assistant","message":{"id":"msg_1","content":[{"type":"thinking","thinking":"","signature":"abc"}]},"parent_tool_use_id":null}`,
		`{"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"text_delta","text":"ignored"}},"parent_tool_use_id":"toolu_1"}`,
		`{"type":"assistant","message":{"id":"msg_sub","content":[{"type":"text","text":"From the subagent.\r\nTwo lines.\n"}]},"parent_tool_use_id":"toolu_1"}`,
		`{"type":"assistant","message":{"id":"err","content":[{"type":"text","text":"Invalid API key"}]},"parent_tool_use_id":null,"error":"authentication_failed"}`,
		`{"type":"result","subtype":"success","is_error":true,"result":"Invalid API key","num_turns":1,"duration_ms":655,"total_cost_usd":0}`,
	)

	assert.Equal(t, []string{
		"Working.",
		"From the subagent.",
		"Two lines.",
		"Invalid API key",
		"Agent failed: 1 turn(s), 700ms",
	}, logged)
}

func TestClaudeStreamOutput_ToolCallsAndResults(t *testing.T) {
	long := strings.Repeat("x", agentToolResultMaxLine+50)
	many := strings.TrimSuffix(strings.Repeat("row\n", agentToolResultMaxLines+5), "\n")

	result := func(content any, isError bool) string {
		data, err := json.Marshal(map[string]any{
			"type": "user",
			"message": map[string]any{"role": "user", "content": []any{
				map[string]any{"type": "tool_result", "tool_use_id": "toolu_1", "content": content, "is_error": isError},
			}},
		})
		require.NoError(t, err)
		return string(data)
	}

	logged := feed(claudeOutput(t),
		`{"type":"assistant","message":{"id":"msg_1","content":[`+
			`{"type":"tool_use","id":"toolu_1","name":"Bash","input":{"command":"echo <a> && ls","timeout":5}},`+
			`{"type":"tool_use","id":"toolu_2","name":"mcp__github__list","input":{}}]}}`,
		result("hello\n", false),
		result([]any{
			map[string]any{"type": "text", "text": "first\nsecond"},
			map[string]any{"type": "image", "source": map[string]any{"data": "AAAA"}},
		}, false),
		result("no such file\ntry again", true),
		result(long, false),
		result(many, false),
		result("", false),
		`{"type":"user","message":{"role":"user","content":"a plain user message"}}`,
	)

	expected := []string{
		`> Bash {"command":"echo <a> && ls","timeout":5}`,
		"> mcp__github__list",
		"  hello",
		"  first",
		"  second",
		"  [image]",
		"  ! no such file",
		"  ! try again",
		"  " + strings.Repeat("x", agentToolResultMaxLine) + "...",
	}
	for range agentToolResultMaxLines {
		expected = append(expected, "  row")
	}
	expected = append(expected, "  ... 5 more line(s)")

	assert.Equal(t, expected, logged)
}

func TestClaudeStreamOutput_LongToolInputIsCut(t *testing.T) {
	input, err := json.Marshal(map[string]string{"content": strings.Repeat("я", agentToolInputMax*2)})
	require.NoError(t, err)

	logged := feed(claudeOutput(t),
		`{"type":"assistant","message":{"id":"msg_1","content":[{"type":"tool_use","name":"Write","input":`+string(input)+`}]}}`)

	require.Len(t, logged, 1)
	assert.True(t, strings.HasPrefix(logged[0], `> Write {"content":"яяя`), logged[0][:40])
	assert.True(t, strings.HasSuffix(logged[0], "я..."))
	assert.Equal(t, len("> Write ")+agentToolInputMax+len("..."), len([]rune(logged[0])))
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

	logged := feed(claudeOutput(t, secret, ""),
		`{"type":"stream_event","event":{"type":"message_start","message":{"id":"msg_1"}},"parent_tool_use_id":null}`,
		textDelta(t, `The token is tok"<`),
		textDelta(t, `&>\x, keep it.`),
		`{"type":"assistant","message":{"id":"msg_2","content":[{"type":"tool_use","name":"Bash","input":`+string(input)+`}]}}`,
		`{"type":"user","message":{"content":[{"type":"tool_result","content":`+string(toolResult)+`}]}}`,
		"plain: "+secret,
	)

	assert.Equal(t, []string{
		"The token is " + maskedSecret + ", keep it.",
		`> Bash {"` + maskedSecret + `":["` + maskedSecret + `"],"command":"curl -H 'Authorization: Bearer ` + maskedSecret + `'"}`,
		"  " + maskedSecret + " is the token",
		"plain: " + maskedSecret,
	}, logged)
}

// What is not an event stays in the log: an error of the CLI, the text or
// json format set in the template's arguments, an event of an unknown shape.
func TestClaudeStreamOutput_OtherOutput(t *testing.T) {
	t.Run("plain text", func(t *testing.T) {
		logged := feed(claudeOutput(t),
			"error: unknown option '--nope'",
			"",
			`{"not":"an event"}`,
			`{broken`,
			`["type"]`,
		)
		assert.Equal(t, []string{"error: unknown option '--nope'", "", `{"not":"an event"}`, `{broken`, `["type"]`}, logged)
	})

	t.Run("plain text ends the streamed line", func(t *testing.T) {
		logged := feed(claudeOutput(t), textDelta(t, "half a line"), "Killed")
		assert.Equal(t, []string{"half a line", "Killed"}, logged)
	})

	t.Run("unfinished line is flushed", func(t *testing.T) {
		assert.Equal(t, []string{"half a line"}, feed(claudeOutput(t), textDelta(t, "half a line")))
	})

	t.Run("json format", func(t *testing.T) {
		logged := feed(claudeOutput(t),
			`{"type":"result","subtype":"success","is_error":false,"result":"All done.\nBye.","num_turns":3,"duration_ms":61000}`)
		assert.Equal(t, []string{"All done.", "Bye.", "Agent finished: 3 turn(s), 1m1s"}, logged)
	})

	t.Run("failed run", func(t *testing.T) {
		logged := feed(claudeOutput(t),
			`{"type":"result","subtype":"error_max_turns","is_error":true,"errors":["Reached the turn limit"],"num_turns":5,"duration_ms":2000}`)
		assert.Equal(t, []string{"Reached the turn limit", "Agent failed (error_max_turns): 5 turn(s), 2s"}, logged)
	})

	t.Run("event of an unknown shape", func(t *testing.T) {
		line := `{"type":"result","result":{"structured":true}}`
		assert.Equal(t, []string{line}, feed(claudeOutput(t), line))
	})
}

func TestTruncateChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{"shorter", "abc", 5, "abc"},
		{"exact", "abcde", 5, "abcde"},
		{"longer", "abcdef", 5, "abcde..."},
		{"multibyte", "привет", 3, "при..."},
		{"empty", "", 3, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, truncateChars(tt.input, tt.max))
		})
	}
}
