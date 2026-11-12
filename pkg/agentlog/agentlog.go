// Package agentlog is the format the progress of an AI agent is written to
// the task log in: one event per log line, as JSON. The web client makes a
// transcript of the events — a tool call with its arguments and its result,
// the agent's text, the subagents; Lines makes plain log lines of them for
// whatever shows the log as text: the raw log, the command line client.
package agentlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// The kinds of events.
const (
	// KindInit opens a run: the model and the MCP servers with their status.
	KindInit = "init"
	// KindText is one line of the agent's text.
	KindText = "text"
	// KindThinking is one line of the agent's reasoning, when the CLI shows it.
	KindThinking = "thinking"
	// KindTool is a tool call: ID, Name and Input. A call logged again with
	// the same ID replaces what was logged: a plan the agent ticks off.
	KindTool = "tool"
	// KindResult is the result of the tool call ID: Text, Error.
	KindResult = "result"
	// KindRetry is a failed API request about to be retried.
	KindRetry = "retry"
	// KindDone closes a turn of the agent: the outcome and what it took.
	KindDone = "done"
)

const (
	// MaxInputString limits every string of a tool call's arguments, in
	// characters: a file the agent wrote is not the log. Arguments which still
	// take more than maxInput bytes get their strings cut to shortString.
	MaxInputString = 4000
	maxInput       = 32 * 1024

	// MaxResultString limits every string of a tool call's result that is
	// JSON, the same way with maxResult; MaxResultText limits a result that is
	// text, or JSON too large even so.
	MaxResultString = 2000
	maxResult       = 16 * 1024
	MaxResultText   = 8000

	shortString = 200

	// The limits of the plain text made of an event.
	textInputMax       = 2000
	textResultMaxLines = 20
	textResultMaxLine  = 1000

	// linePrefix is what every event line starts with.
	linePrefix = `{"@agent":"`
)

// Server is an MCP server given to the agent and whether it is connected.
type Server struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Event is one event of an agent run. Kind says which fields are set.
type Event struct {
	// Kind goes first: a line starting with {"@agent":" is an event.
	Kind string `json:"@agent"`

	// Agent is the ID of the tool call which started the subagent the event
	// belongs to; empty for the agent itself.
	Agent string `json:"agent,omitempty"`

	// ID is the tool call, of the call and of its result.
	ID string `json:"id,omitempty"`
	// Name and Input are the tool and its arguments, a JSON value.
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`

	// Text is a line of the agent's text or reasoning, the result of a tool
	// call or the reason of a retry.
	Text string `json:"text,omitempty"`
	// Error tells that the tool call or the turn failed.
	Error bool `json:"error,omitempty"`
	// Truncated is the number of characters left out of Input or Text.
	Truncated int `json:"truncated,omitempty"`

	Model   string   `json:"model,omitempty"`
	Servers []Server `json:"mcp,omitempty"`

	Attempt    int     `json:"attempt,omitempty"`
	MaxRetries int     `json:"max_retries,omitempty"`
	DelayMs    float64 `json:"delay_ms,omitempty"`

	// Subtype is how the turn ended when not with success, as error_max_turns.
	Subtype string   `json:"subtype,omitempty"`
	Errors  []string `json:"errors,omitempty"`
	// What the turn took, as far as the CLI tells.
	Turns        int     `json:"turns,omitempty"`
	DurationMs   float64 `json:"duration_ms,omitempty"`
	CostUSD      float64 `json:"cost_usd,omitempty"`
	InputTokens  int64   `json:"input_tokens,omitempty"`
	OutputTokens int64   `json:"output_tokens,omitempty"`
}

// Line encodes the event as a log line.
func (e Event) Line() string {
	var line bytes.Buffer
	encoder := json.NewEncoder(&line)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(e); err != nil {
		// Only Input can fail to encode; the event goes without it.
		e.Input = nil
		line.Reset()
		_ = encoder.Encode(e)
	}
	return strings.TrimSuffix(line.String(), "\n")
}

// Parse decodes a log line; false when the line is not an event.
func Parse(line string) (Event, bool) {
	var event Event
	if !strings.HasPrefix(line, linePrefix) || json.Unmarshal([]byte(line), &event) != nil {
		return Event{}, false
	}

	switch event.Kind {
	case KindInit, KindText, KindThinking, KindTool, KindResult, KindRetry, KindDone:
		return event, true
	default:
		return Event{}, false
	}
}

// Lines returns a log line as plain text: the lines Plain makes of an event,
// the line itself when it is not one.
func Lines(line string) []string {
	if event, ok := Parse(line); ok {
		return event.Plain()
	}
	return []string{line}
}

// NewTool makes the event of a tool call. input is the JSON of its
// arguments: mask is applied to every string of it, the keys included, and
// the strings are cut to MaxInputString. The order of the keys is kept.
func NewTool(id string, name string, input []byte, mask func(string) string) Event {
	event := Event{Kind: KindTool, ID: mask(id), Name: mask(name)}

	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("{}")) || bytes.Equal(trimmed, []byte("null")) {
		return event
	}

	event.Input, event.Truncated, _ = capJSON(trimmed, mask, MaxInputString, maxInput)
	return event
}

// NewResult makes the event of a tool call's result, masked and cut. A result
// that is JSON stays JSON: its strings are masked and cut one by one, so that
// a client can still read it.
func NewResult(id string, text string, failed bool, mask func(string) string) Event {
	event := Event{Kind: KindResult, ID: mask(id), Error: failed}

	text = strings.Trim(strings.ReplaceAll(text, "\r\n", "\n"), "\n")

	if trimmed := strings.TrimSpace(text); strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		if result, cut, ok := capJSON([]byte(trimmed), mask, MaxResultString, maxResult); ok {
			event.Text, event.Truncated = string(result), cut
			return event
		}
	}

	event.Text, event.Truncated = capString(mask(text), MaxResultText)
	return event
}

// capJSON rewrites a JSON text with mask applied to its strings and the
// strings cut to limit characters, or to shortString when the result would
// take more than maxBytes. It returns the JSON and the number of characters
// left out; false when raw is not JSON or is too large even so.
func capJSON(raw []byte, mask func(string) string, limit int, maxBytes int) (json.RawMessage, int, bool) {
	for _, limit := range []int{limit, shortString} {
		cut := 0
		result, err := rewriteJSON(raw, func(s string, key bool) string {
			s = mask(s)
			if key {
				return s
			}
			s, n := capString(s, limit)
			cut += n
			return s
		})
		if err != nil {
			return nil, 0, false
		}
		if len(result) <= maxBytes {
			return result, cut, true
		}
	}
	return nil, 0, false
}

// rewriteJSON returns a JSON text with every string of it, the keys of its
// objects included, replaced by rewrite. Everything else is kept: the numbers
// as they are written, the order of the keys.
func rewriteJSON(raw []byte, rewrite func(s string, key bool) string) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var out bytes.Buffer
	if err := rewriteJSONValue(decoder, &out, rewrite); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, errors.New("agentlog: more than one JSON value")
	}
	return out.Bytes(), nil
}

func rewriteJSONValue(decoder *json.Decoder, out *bytes.Buffer, rewrite func(s string, key bool) string) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}

	switch value := token.(type) {
	case json.Delim:
		closing := byte('}')
		if value == '[' {
			closing = ']'
		}
		out.WriteRune(rune(value))
		for first := true; decoder.More(); first = false {
			if !first {
				out.WriteByte(',')
			}
			if value == '{' {
				key, err := decoder.Token()
				if err != nil {
					return err
				}
				name, _ := key.(string)
				writeJSONString(out, rewrite(name, true))
				out.WriteByte(':')
			}
			if err = rewriteJSONValue(decoder, out, rewrite); err != nil {
				return err
			}
		}
		if _, err = decoder.Token(); err != nil {
			return err
		}
		out.WriteByte(closing)
	case string:
		writeJSONString(out, rewrite(value, false))
	case json.Number:
		out.WriteString(value.String())
	case bool:
		out.WriteString(strconv.FormatBool(value))
	case nil:
		out.WriteString("null")
	}

	return nil
}

// writeJSONString writes s as a JSON string, without escaping for HTML: an
// event is never inserted into a page as HTML.
func writeJSONString(out *bytes.Buffer, s string) {
	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(s)
	out.Truncate(out.Len() - 1) // the newline Encode adds
}

// Plain returns the event as plain log lines.
func (e Event) Plain() []string {
	switch e.Kind {
	case KindInit:
		lines := make([]string, 0, len(e.Servers)+1)
		if e.Model != "" {
			lines = append(lines, "Model: "+e.Model)
		}
		for _, server := range e.Servers {
			lines = append(lines, fmt.Sprintf("MCP server %s: %s", server.Name, server.Status))
		}
		return lines

	case KindText, KindThinking:
		return []string{e.Text}

	case KindTool:
		line := "> " + e.Name
		if len(e.Input) > 0 {
			input, _ := capString(string(e.Input), textInputMax)
			line += " " + input
		}
		return []string{line}

	case KindResult:
		prefix := "  "
		if e.Error {
			prefix = "  ! "
		}
		if e.Text == "" {
			return nil
		}
		result := strings.Split(e.Text, "\n")
		lines := make([]string, 0, min(len(result), textResultMaxLines+1))
		for i, line := range result {
			if i == textResultMaxLines {
				lines = append(lines, fmt.Sprintf("%s... %d more line(s)", prefix, len(result)-i))
				break
			}
			line, _ = capString(line, textResultMaxLine)
			lines = append(lines, prefix+line)
		}
		return lines

	case KindRetry:
		line := fmt.Sprintf("API request failed, retry %d of %d", e.Attempt, e.MaxRetries)
		if e.DelayMs > 0 {
			line += " in " + duration(e.DelayMs)
		}
		if e.Text != "" {
			line += ": " + e.Text
		}
		return []string{line}

	case KindDone:
		status := "Agent finished"
		if e.Error || e.Subtype != "" {
			status = "Agent failed"
			if e.Subtype != "" {
				status += " (" + e.Subtype + ")"
			}
		}
		var took []string
		if e.Turns > 0 {
			took = append(took, fmt.Sprintf("%d turn(s)", e.Turns))
		}
		if e.DurationMs > 0 {
			took = append(took, duration(e.DurationMs))
		}
		if e.InputTokens > 0 || e.OutputTokens > 0 {
			took = append(took, fmt.Sprintf("%d input / %d output tokens", e.InputTokens, e.OutputTokens))
		}
		if e.CostUSD > 0 {
			took = append(took, fmt.Sprintf("$%.4f", e.CostUSD))
		}

		summary := status
		if len(took) > 0 {
			summary += ": " + strings.Join(took, ", ")
		}
		return append(append([]string{}, e.Errors...), summary)
	}

	return nil
}

// capString cuts s to limit characters, marking the cut, and returns it with
// the number of characters left out.
func capString(s string, limit int) (string, int) {
	count := 0
	for i := range s {
		if count == limit {
			return s[:i] + "...", len([]rune(s[i:]))
		}
		count++
	}
	return s, 0
}

// duration renders milliseconds as a duration with a tenth of a second.
func duration(ms float64) string {
	return (time.Duration(ms) * time.Millisecond).Round(100 * time.Millisecond).String()
}
