package agentlog

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvent_LineAndParse(t *testing.T) {
	event := Event{
		Kind:  KindTool,
		Agent: "toolu_0",
		ID:    "toolu_1",
		Name:  "Bash",
		Input: json.RawMessage(`{"command":"echo <a> && ls\nexit"}`),
	}

	line := event.Line()
	assert.Equal(t, `{"@agent":"tool","agent":"toolu_0","id":"toolu_1","name":"Bash","input":{"command":"echo <a> && ls\nexit"}}`, line)
	assert.NotContains(t, line, "\n", "an event is one log line")

	parsed, ok := Parse(line)
	require.True(t, ok)
	assert.Equal(t, event, parsed)
}

func TestEvent_LineOfEmptyText(t *testing.T) {
	line := Event{Kind: KindText}.Line()
	assert.Equal(t, `{"@agent":"text"}`, line)

	parsed, ok := Parse(line)
	require.True(t, ok)
	assert.Equal(t, []string{""}, parsed.Plain())
}

func TestParse_NotAnEvent(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{"plain text", "PLAY [all]"},
		{"empty", ""},
		{"other json", `{"type":"assistant"}`},
		{"key is not first", `{"text":"hi","@agent":"text"}`},
		{"leading space", ` {"@agent":"text"}`},
		{"broken json", `{"@agent":"text","text":`},
		{"unknown kind", `{"@agent":"shout","text":"hi"}`},
		{"kind is not a string", `{"@agent":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := Parse(tt.line)
			assert.False(t, ok)
			assert.Equal(t, []string{tt.line}, Lines(tt.line), "the line stays as it is")
		})
	}
}

func TestLines_EventsAsPlainText(t *testing.T) {
	manyRows := strings.TrimSuffix(strings.Repeat("row\n", textResultMaxLines+5), "\n")

	tests := []struct {
		name     string
		event    Event
		expected []string
	}{
		{
			"init",
			Event{Kind: KindInit, Model: "claude-opus-5-5", Servers: []Server{{"github", "connected"}, {"docs", "failed"}}},
			[]string{"Model: claude-opus-5-5", "MCP server github: connected", "MCP server docs: failed"},
		},
		{"text", Event{Kind: KindText, Text: "Two open issues"}, []string{"Two open issues"}},
		{"tool without input", Event{Kind: KindTool, Name: "mcp__github__list"}, []string{"> mcp__github__list"}},
		{
			"tool",
			Event{Kind: KindTool, Name: "Bash", Input: json.RawMessage(`{"command":"ls"}`)},
			[]string{`> Bash {"command":"ls"}`},
		},
		{
			"long tool input",
			Event{Kind: KindTool, Name: "Write", Input: json.RawMessage(`"` + strings.Repeat("я", textInputMax*2) + `"`)},
			[]string{`> Write "` + strings.Repeat("я", textInputMax-1) + "..."},
		},
		{"result", Event{Kind: KindResult, Text: "first\nsecond"}, []string{"  first", "  second"}},
		{"failed result", Event{Kind: KindResult, Text: "no such file\ntry again", Error: true}, []string{"  ! no such file", "  ! try again"}},
		{"empty result", Event{Kind: KindResult}, nil},
		{
			"long result line",
			Event{Kind: KindResult, Text: strings.Repeat("x", textResultMaxLine+50)},
			[]string{"  " + strings.Repeat("x", textResultMaxLine) + "..."},
		},
		{
			"result of many lines",
			Event{Kind: KindResult, Text: manyRows},
			append(strings.Split(strings.TrimSuffix(strings.Repeat("  row\n", textResultMaxLines), "\n"), "\n"), "  ... 5 more line(s)"),
		},
		{
			"retry",
			Event{Kind: KindRetry, Attempt: 2, MaxRetries: 10, DelayMs: 1500, Text: "overloaded"},
			[]string{"API request failed, retry 2 of 10 in 1.5s: overloaded"},
		},
		{
			"done",
			Event{Kind: KindDone, Turns: 2, DurationMs: 12909, CostUSD: 0.19893},
			[]string{"Agent finished: 2 turn(s), 12.9s, $0.1989"},
		},
		{
			"failed turn",
			Event{Kind: KindDone, Error: true, Subtype: "error_max_turns", Errors: []string{"Reached the turn limit"}, Turns: 5, DurationMs: 61000},
			[]string{"Reached the turn limit", "Agent failed (error_max_turns): 5 turn(s), 1m1s"},
		},
		{"turn failed by the API", Event{Kind: KindDone, Error: true, Turns: 1, DurationMs: 655}, []string{"Agent failed: 1 turn(s), 700ms"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Lines(tt.event.Line()))
		})
	}
}

func noMask(s string) string { return s }

func TestNewTool(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		for _, input := range []string{"", "{}", " {}\n", "null"} {
			assert.Equal(t, Event{Kind: KindTool, ID: "t1", Name: "List"}, NewTool("t1", "List", []byte(input), noMask), input)
		}
	})

	t.Run("arguments are kept as they are, in their order", func(t *testing.T) {
		const input = `{"prompt":"echo <a>","description":"d","timeout":5.0,"env":["A=1",true,null],"n":{"z":1,"a":2}}`
		event := NewTool("t1", "Bash", []byte(" "+input+"\n"), noMask)
		assert.Equal(t, input, string(event.Input))
		assert.Zero(t, event.Truncated)
	})

	t.Run("arguments that are not JSON are left out", func(t *testing.T) {
		assert.Equal(t, Event{Kind: KindTool, ID: "t1", Name: "Bash"}, NewTool("t1", "Bash", []byte(`{"a":`), noMask))
		assert.Equal(t, Event{Kind: KindTool, ID: "t1", Name: "Bash"}, NewTool("t1", "Bash", []byte(`{"a":1} {"b":2}`), noMask))
	})

	t.Run("a long string is cut", func(t *testing.T) {
		input, err := json.Marshal(map[string]any{
			"path":    "a.txt",
			"content": strings.Repeat("я", MaxInputString+10),
			"lines":   []any{strings.Repeat("x", MaxInputString+5)},
		})
		require.NoError(t, err)

		event := NewTool("t1", "Write", input, noMask)

		var decoded struct {
			Path    string   `json:"path"`
			Content string   `json:"content"`
			Lines   []string `json:"lines"`
		}
		require.NoError(t, json.Unmarshal(event.Input, &decoded))
		assert.Equal(t, "a.txt", decoded.Path)
		assert.Equal(t, strings.Repeat("я", MaxInputString)+"...", decoded.Content)
		assert.Equal(t, strings.Repeat("x", MaxInputString)+"...", decoded.Lines[0])
		assert.Equal(t, 15, event.Truncated)
	})

	t.Run("many long strings are cut shorter", func(t *testing.T) {
		items := make([]string, 20)
		for i := range items {
			items[i] = strings.Repeat("x", MaxInputString)
		}
		input, err := json.Marshal(map[string]any{"items": items})
		require.NoError(t, err)

		event := NewTool("t1", "Batch", input, noMask)
		assert.LessOrEqual(t, len(event.Input), maxInput)
		assert.Equal(t, 20*(MaxInputString-shortString), event.Truncated)
		assert.True(t, json.Valid(event.Input))
	})

	t.Run("strings and keys are masked before they are encoded", func(t *testing.T) {
		const secret = `tok"<&>\x`
		mask := strings.NewReplacer(secret, "********").Replace

		input, err := json.Marshal(map[string]any{"auth": "Bearer " + secret, secret: []string{secret}})
		require.NoError(t, err)

		event := NewTool("id-"+secret, "tool-"+secret, input, mask)
		assert.Equal(t, "id-********", event.ID)
		assert.Equal(t, "tool-********", event.Name)
		assert.JSONEq(t, `{"auth":"Bearer ********","********":["********"]}`, string(event.Input))
	})
}

func TestNewResult(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		assert.Equal(t,
			Event{Kind: KindResult, ID: "t1", Text: "first\nsecond", Error: true},
			NewResult("t1", "\r\nfirst\r\nsecond\n\n", true, noMask))
	})

	t.Run("long text", func(t *testing.T) {
		event := NewResult("t1", strings.Repeat("я", MaxResultText+7), false, noMask)
		assert.Equal(t, strings.Repeat("я", MaxResultText)+"...", event.Text)
		assert.Equal(t, 7, event.Truncated)
	})

	t.Run("JSON stays JSON, in its order", func(t *testing.T) {
		const result = `{"code":"SOURCE@1","status":"kept","n":[1,2.50,{"b":null,"a":false}]}`
		event := NewResult("t1", "  \n"+result+"\n", false, noMask)
		assert.Equal(t, Event{Kind: KindResult, ID: "t1", Text: result}, event)
	})

	t.Run("a long string of JSON is cut and the JSON stays valid", func(t *testing.T) {
		result, err := json.Marshal(map[string]any{"code": "SOURCE@1", "body": strings.Repeat("я", MaxResultString+9)})
		require.NoError(t, err)

		event := NewResult("t1", string(result), false, noMask)
		assert.Equal(t, 9, event.Truncated)
		assert.JSONEq(t, `{"code":"SOURCE@1","body":"`+strings.Repeat("я", MaxResultString)+`..."}`, event.Text)
	})

	t.Run("JSON too large even with short strings is cut as text", func(t *testing.T) {
		items := make([]string, 4000)
		for i := range items {
			items[i] = "item"
		}
		result, err := json.Marshal(items)
		require.NoError(t, err)

		event := NewResult("t1", string(result), false, noMask)
		assert.Equal(t, MaxResultText+len("..."), len(event.Text))
		assert.Equal(t, len(result)-MaxResultText, event.Truncated)
	})

	t.Run("what only looks like JSON is text", func(t *testing.T) {
		assert.Equal(t, "[image]", NewResult("t1", "[image]", false, noMask).Text)
		assert.Equal(t, `{"a":1} and more`, NewResult("t1", `{"a":1} and more`, false, noMask).Text)
	})

	t.Run("a secret is masked in text and inside JSON, where it is escaped", func(t *testing.T) {
		const secret = `tok"<&>\x`
		mask := strings.NewReplacer(secret, "********").Replace

		assert.Equal(t, "the token is ********", NewResult("t1", "the token is "+secret, false, mask).Text)

		result, err := json.Marshal(map[string]string{"env": "TOKEN=" + secret})
		require.NoError(t, err)
		event := NewResult("t1", string(result), false, mask)
		assert.Equal(t, `{"env":"TOKEN=********"}`, event.Text)
	})
}
