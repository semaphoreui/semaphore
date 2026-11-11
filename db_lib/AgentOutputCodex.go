package db_lib

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/pkg/agentlog"
)

// codexJSONOutput reads the events of codex exec --json and writes them to
// the log as agentlog events, as claudeStreamOutput does for Claude Code: a
// client makes one transcript of both. Whatever is not an event of the CLI —
// its own log on stderr, as a failed connection to an MCP server — goes to
// the log as it came.
type codexJSONOutput struct {
	mask func(string) string
	// now is the clock: the CLI does not tell how long a turn took.
	now func() time.Time

	turnStart time.Time
	// called are the items logged as tool calls and waiting for their result.
	called map[string]bool
	// lastError is the last error logged: a failed turn repeats it.
	lastError string
}

type codexEvent struct {
	Type string     `json:"type"`
	Item *codexItem `json:"item"`

	// turn.completed
	Usage struct {
		InputTokens  int64 `json:"input_tokens"`
		OutputTokens int64 `json:"output_tokens"`
	} `json:"usage"`

	// error
	Message string `json:"message"`
	// turn.failed
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// codexItem is what a turn consists of; Type says which fields are set.
type codexItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`

	// agent_message, reasoning
	Text string `json:"text"`

	// command_execution
	Command          string `json:"command"`
	AggregatedOutput string `json:"aggregated_output"`
	ExitCode         *int   `json:"exit_code"`

	// in_progress, completed, failed, declined
	Status string `json:"status"`

	// mcp_tool_call; Tool is also the action of a collab_tool_call
	Server    string          `json:"server"`
	Tool      string          `json:"tool"`
	Arguments json.RawMessage `json:"arguments"`
	Result    *struct {
		Content           []json.RawMessage `json:"content"`
		StructuredContent json.RawMessage   `json:"structured_content"`
	} `json:"result"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`

	// file_change
	Changes []struct {
		Path string `json:"path"`
		Kind string `json:"kind"`
	} `json:"changes"`

	// web_search
	Query string `json:"query"`

	// todo_list
	Items []struct {
		Text      string `json:"text"`
		Completed bool   `json:"completed"`
	} `json:"items"`

	// collab_tool_call
	Prompt *string `json:"prompt"`

	// error
	Message string `json:"message"`
}

func (o *codexJSONOutput) Lines(line string) []string {
	raw := []byte(strings.TrimSpace(line))

	var event codexEvent
	if len(raw) == 0 || raw[0] != '{' || json.Unmarshal(raw, &event) != nil || event.Type == "" {
		// Not an event: the CLI's own log, or its plain output when the
		// template's arguments turned the events off.
		return []string{o.mask(line)}
	}

	switch event.Type {
	case "turn.started":
		o.turnStart = o.now()
		return nil

	case "item.started", "item.updated", "item.completed":
		if event.Item == nil {
			return nil
		}
		return o.item(*event.Item, event.Type == "item.completed")

	case "error":
		return o.errorLine(event.Message)

	case "turn.completed":
		return encodeAgentEvents([]agentlog.Event{o.done(agentlog.Event{
			InputTokens:  event.Usage.InputTokens,
			OutputTokens: event.Usage.OutputTokens,
		})})

	case "turn.failed":
		done := agentlog.Event{Error: true}
		// The error that failed the turn has been logged when it came.
		if message := event.Error.Message; message != "" && message != o.lastError {
			done.Errors = []string{o.mask(message)}
		}
		return encodeAgentEvents([]agentlog.Event{o.done(done)})
	}

	return nil
}

func (o *codexJSONOutput) Flush() []string { return nil }

// done completes the event of a turn's end with the time the turn took.
func (o *codexJSONOutput) done(event agentlog.Event) agentlog.Event {
	event.Kind = agentlog.KindDone
	if !o.turnStart.IsZero() {
		event.DurationMs = float64(o.now().Sub(o.turnStart).Milliseconds())
		o.turnStart = time.Time{}
	}
	return event
}

// errorLine logs an error the CLI reports as a plain line.
func (o *codexJSONOutput) errorLine(message string) []string {
	if message == "" {
		return nil
	}
	o.lastError = message
	return []string{o.mask("Error: " + message)}
}

// item makes the events of an item. An item that takes time — a command, a
// tool call — is reported when it starts and when it is completed: the call
// is logged at the first of the two, the result at the completion.
func (o *codexJSONOutput) item(item codexItem, completed bool) []string {
	switch item.Type {
	case "agent_message":
		return o.text(agentlog.KindText, item.Text, completed)
	case "reasoning":
		return o.text(agentlog.KindThinking, item.Text, completed)
	case "error":
		if !completed {
			return nil
		}
		return o.errorLine(item.Message)
	}

	name, input, repeat := o.toolCall(item)
	if name == "" {
		return nil
	}

	var events []agentlog.Event

	if repeat || !o.called[item.ID] {
		if o.called == nil {
			o.called = map[string]bool{}
		}
		o.called[item.ID] = true
		events = append(events, agentlog.NewTool(item.ID, name, input, o.mask))
	}

	if completed {
		delete(o.called, item.ID)
		text, failed := codexToolResult(item)
		events = append(events, agentlog.NewResult(item.ID, text, failed, o.mask))
	}

	return encodeAgentEvents(events)
}

// text makes an event of every line of a completed message.
func (o *codexJSONOutput) text(kind string, text string, completed bool) []string {
	if !completed {
		return nil
	}

	lines := splitAgentText(text)
	events := make([]agentlog.Event, 0, len(lines))
	for _, line := range lines {
		events = append(events, agentlog.Event{Kind: kind, Text: o.mask(line)})
	}
	return encodeAgentEvents(events)
}

// toolCall returns the name and the arguments, as JSON, an item is logged as
// a tool call with; an empty name for an item that is not one. repeat tells
// that the call is logged again every time it is reported: a plan changes.
func (o *codexJSONOutput) toolCall(item codexItem) (name string, input []byte, repeat bool) {
	var arguments any

	switch item.Type {
	case "command_execution":
		name, arguments = "exec", map[string]string{"command": item.Command}

	case "mcp_tool_call":
		// The name of an MCP tool as Claude Code spells it: a client tells
		// the server from the tool by it.
		return "mcp__" + item.Server + "__" + item.Tool, item.Arguments, false

	case "file_change":
		name, arguments = "file_change", map[string]any{"changes": item.Changes}

	case "web_search":
		name, arguments = "web_search", map[string]string{"query": item.Query}

	case "todo_list":
		steps := make([]string, 0, len(item.Items))
		completed := 0
		for _, step := range item.Items {
			mark := "[ ] "
			if step.Completed {
				mark = "[x] "
				completed++
			}
			steps = append(steps, mark+step.Text)
		}
		name, repeat = "todo_list", true
		arguments = struct {
			Progress string   `json:"progress"`
			Steps    []string `json:"steps"`
		}{fmt.Sprintf("%d/%d", completed, len(steps)), steps}

	case "collab_tool_call":
		name = item.Tool
		if name == "" {
			name = "collab"
		}
		if item.Prompt != nil {
			arguments = map[string]string{"prompt": *item.Prompt}
		}

	default:
		return "", nil, false
	}

	if arguments != nil {
		input, _ = json.Marshal(arguments)
	}
	return name, input, repeat
}

// codexToolResult returns the result of a completed item and whether it
// failed.
func codexToolResult(item codexItem) (string, bool) {
	failed := item.Status == "failed" || item.Status == "declined"

	switch item.Type {
	case "command_execution":
		text := item.AggregatedOutput
		if item.ExitCode != nil && *item.ExitCode != 0 {
			failed = true
			text = fmt.Sprintf("Exit code %d\n%s", *item.ExitCode, text)
		} else if item.Status == "declined" && text == "" {
			text = "The command was declined"
		}
		return text, failed

	case "mcp_tool_call":
		if item.Error != nil && item.Error.Message != "" {
			return item.Error.Message, true
		}
		if item.Result == nil {
			return "", failed
		}
		if len(item.Result.Content) == 0 && string(item.Result.StructuredContent) != "null" {
			return string(item.Result.StructuredContent), failed
		}
		return codexContentText(item.Result.Content), failed

	case "file_change":
		if failed {
			return "The change was not applied", true
		}
	}

	return "", failed
}

// codexContentText returns the text of the content of an MCP tool's result; a
// block that is not text is named, as "[image]".
func codexContentText(content []json.RawMessage) string {
	parts := make([]string, 0, len(content))
	for _, raw := range content {
		var block struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if json.Unmarshal(raw, &block) != nil {
			continue
		}
		if block.Type == "text" {
			parts = append(parts, block.Text)
		} else if block.Type != "" {
			parts = append(parts, "["+block.Type+"]")
		}
	}
	return strings.Join(parts, "\n")
}
