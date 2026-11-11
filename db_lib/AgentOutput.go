package db_lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	// agentToolInputMax limits the arguments of a tool call in the task log,
	// in characters.
	agentToolInputMax = 2000
	// agentToolResultMaxLines and agentToolResultMaxLine limit the result of a
	// tool call in the task log: a file the agent read is not the log.
	agentToolResultMaxLines = 20
	agentToolResultMaxLine  = 1000
)

// AgentOutput turns the lines an agent CLI prints into task log lines, with
// the secrets of the launch masked. It is not safe for concurrent use: feed it
// the lines of one run, in the order they were printed.
type AgentOutput interface {
	// Lines returns the log lines for one line of the CLI's output; none when
	// the line carries nothing to show.
	Lines(line string) []string
	// Flush returns what is still held back, once the output has ended.
	Flush() []string
}

// newSecretMasker returns a function replacing every secret with maskedSecret.
func newSecretMasker(secrets []string) func(string) string {
	pairs := make([]string, 0, len(secrets)*2)
	for _, s := range secrets {
		if s != "" {
			pairs = append(pairs, s, maskedSecret)
		}
	}
	if len(pairs) == 0 {
		return func(s string) string { return s }
	}
	return strings.NewReplacer(pairs...).Replace
}

// plainAgentOutput is the output of a CLI that reports its progress in plain
// text as it goes (codex exec): every line is a log line.
type plainAgentOutput struct {
	mask func(string) string
}

func (o plainAgentOutput) Lines(line string) []string { return []string{o.mask(line)} }
func (plainAgentOutput) Flush() []string              { return nil }

// claudeStreamOutput reads the events of claude -p --output-format
// stream-json. In its text format Claude Code prints the answer only when the
// run is over; the events come as things happen: the text as it is written,
// every tool call and its result.
type claudeStreamOutput struct {
	mask func(string) string

	// partial is the unfinished last line of the text being streamed.
	partial string
	// messageID is the message being streamed, streamedID the one whose text
	// was logged from the stream and must not be logged again as a whole.
	messageID  string
	streamedID string
	// said tells that some text of the agent reached the log.
	said bool
}

type claudeEventHead struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
}

type claudeSystemEvent struct {
	Model      string `json:"model"`
	McpServers []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"mcp_servers"`

	// api_retry
	Attempt      int     `json:"attempt"`
	MaxRetries   int     `json:"max_retries"`
	RetryDelayMs float64 `json:"retry_delay_ms"`
	Error        string  `json:"error"`
}

type claudeBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`

	// tool_use
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`

	// tool_result
	Content json.RawMessage `json:"content"`
	IsError bool            `json:"is_error"`
}

type claudeMessageEvent struct {
	Message struct {
		ID string `json:"id"`
		// Content is a list of blocks, or a string for a plain user message.
		Content json.RawMessage `json:"content"`
	} `json:"message"`
	// ParentToolUseID is set on the messages of a subagent.
	ParentToolUseID *string `json:"parent_tool_use_id"`
}

type claudeStreamEvent struct {
	Event struct {
		Type    string `json:"type"`
		Message struct {
			ID string `json:"id"`
		} `json:"message"`
		Delta struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"delta"`
	} `json:"event"`
	ParentToolUseID *string `json:"parent_tool_use_id"`
}

type claudeResultEvent struct {
	Subtype      string   `json:"subtype"`
	IsError      bool     `json:"is_error"`
	Result       string   `json:"result"`
	Errors       []string `json:"errors"`
	NumTurns     int      `json:"num_turns"`
	DurationMs   float64  `json:"duration_ms"`
	TotalCostUSD float64  `json:"total_cost_usd"`
}

func (o *claudeStreamOutput) Lines(line string) []string {
	raw := []byte(strings.TrimSpace(line))

	var head claudeEventHead
	if len(raw) == 0 || raw[0] != '{' || json.Unmarshal(raw, &head) != nil || head.Type == "" {
		// Not an event: an error of the CLI itself, or the output of another
		// format set in the template's arguments.
		return o.masked(append(o.endText(), line))
	}

	var (
		lines []string
		err   error
	)

	switch head.Type {
	case "system":
		lines, err = o.system(head.Subtype, raw)
	case "stream_event":
		lines, err = o.streamEvent(raw)
	case "assistant":
		lines, err = o.assistant(raw)
	case "user":
		lines, err = o.user(raw)
	case "result":
		lines, err = o.result(raw)
	}

	if err != nil {
		// An event of a shape this code does not know is shown as it came
		// rather than hidden.
		return o.masked(append(o.endText(), line))
	}

	return o.masked(lines)
}

func (o *claudeStreamOutput) Flush() []string {
	return o.masked(o.endText())
}

func (o *claudeStreamOutput) masked(lines []string) []string {
	for i, line := range lines {
		lines[i] = o.mask(line)
	}
	return lines
}

// endText returns the unfinished line of the streamed text, if any.
func (o *claudeStreamOutput) endText() []string {
	if o.partial == "" {
		return nil
	}
	line := strings.TrimRight(o.partial, "\r")
	o.partial = ""
	return []string{line}
}

func (o *claudeStreamOutput) system(subtype string, raw []byte) ([]string, error) {
	if subtype != "init" && subtype != "api_retry" {
		return nil, nil
	}

	var event claudeSystemEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, err
	}

	lines := o.endText()

	if subtype == "api_retry" {
		line := fmt.Sprintf("API request failed, retry %d of %d", event.Attempt, event.MaxRetries)
		if event.RetryDelayMs > 0 {
			line += " in " + agentDuration(event.RetryDelayMs)
		}
		if event.Error != "" {
			line += ": " + event.Error
		}
		return append(lines, line), nil
	}

	if event.Model != "" {
		lines = append(lines, "Model: "+event.Model)
	}
	for _, server := range event.McpServers {
		lines = append(lines, fmt.Sprintf("MCP server %s: %s", server.Name, server.Status))
	}
	return lines, nil
}

// streamEvent logs the agent's text line by line while it is being written.
func (o *claudeStreamOutput) streamEvent(raw []byte) ([]string, error) {
	var event claudeStreamEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, err
	}

	// Subagents may write at the same time: their text is logged as whole
	// messages, not interleaved by the piece.
	if event.ParentToolUseID != nil {
		return nil, nil
	}

	switch event.Event.Type {
	case "message_start":
		o.messageID = event.Event.Message.ID
		return o.endText(), nil
	case "content_block_stop", "message_stop":
		return o.endText(), nil
	case "content_block_delta":
		if event.Event.Delta.Type != "text_delta" {
			return nil, nil
		}
	default:
		return nil, nil
	}

	if event.Event.Delta.Text == "" {
		return nil, nil
	}

	o.streamedID = o.messageID
	o.said = true
	o.partial += event.Event.Delta.Text

	var lines []string
	for {
		i := strings.IndexByte(o.partial, '\n')
		if i < 0 {
			break
		}
		lines = append(lines, strings.TrimRight(o.partial[:i], "\r"))
		o.partial = o.partial[i+1:]
	}
	return lines, nil
}

func (o *claudeStreamOutput) assistant(raw []byte) ([]string, error) {
	var event claudeMessageEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, err
	}

	lines := o.endText()

	streamed := event.ParentToolUseID == nil && event.Message.ID != "" && event.Message.ID == o.streamedID

	for _, block := range claudeBlocks(event.Message.Content) {
		switch block.Type {
		case "text":
			if streamed {
				continue
			}
			if text := splitAgentText(block.Text); len(text) > 0 {
				lines = append(lines, text...)
				o.said = true
			}
		case "tool_use":
			lines = append(lines, o.toolCall(block))
		}
	}

	return lines, nil
}

func (o *claudeStreamOutput) user(raw []byte) ([]string, error) {
	var event claudeMessageEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, err
	}

	lines := o.endText()

	for _, block := range claudeBlocks(event.Message.Content) {
		if block.Type != "tool_result" {
			continue
		}

		prefix := "  "
		if block.IsError {
			prefix = "  ! "
		}

		result := splitAgentText(claudeToolResultText(block.Content))
		for i, line := range result {
			if i == agentToolResultMaxLines {
				lines = append(lines, fmt.Sprintf("%s... %d more line(s)", prefix, len(result)-i))
				break
			}
			lines = append(lines, prefix+truncateChars(line, agentToolResultMaxLine))
		}
	}

	return lines, nil
}

func (o *claudeStreamOutput) result(raw []byte) ([]string, error) {
	var event claudeResultEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, err
	}

	lines := o.endText()

	// The result repeats the last message; it is the only text when the run
	// printed nothing before it (the json output format, an early error).
	if !o.said {
		lines = append(lines, splitAgentText(event.Result)...)
	}
	lines = append(lines, event.Errors...)

	status := "Agent finished"
	if event.IsError || (event.Subtype != "" && event.Subtype != "success") {
		status = "Agent failed"
		if event.Subtype != "" && event.Subtype != "success" {
			status += " (" + event.Subtype + ")"
		}
	}

	summary := fmt.Sprintf("%s: %d turn(s), %s", status, event.NumTurns, agentDuration(event.DurationMs))
	if event.TotalCostUSD > 0 {
		summary += fmt.Sprintf(", $%.4f", event.TotalCostUSD)
	}

	return append(lines, summary), nil
}

// toolCall renders a tool call as "> name {arguments}" on one line.
func (o *claudeStreamOutput) toolCall(block claudeBlock) string {
	line := "> " + block.Name

	var input any
	decoder := json.NewDecoder(bytes.NewReader(block.Input))
	decoder.UseNumber()
	if len(block.Input) == 0 || decoder.Decode(&input) != nil {
		return line
	}

	if object, ok := input.(map[string]any); input == nil || (ok && len(object) == 0) {
		return line
	}

	// Mask the values before they are encoded: a secret with a character JSON
	// escapes would not be found in the encoded line.
	input = maskJSONStrings(input, o.mask)

	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if encoder.Encode(input) != nil {
		return line
	}

	return line + " " + truncateChars(strings.TrimSpace(encoded.String()), agentToolInputMax)
}

// claudeBlocks decodes the content of a message; nil when it is not a list.
func claudeBlocks(content json.RawMessage) []claudeBlock {
	var blocks []claudeBlock
	if json.Unmarshal(content, &blocks) != nil {
		return nil
	}
	return blocks
}

// claudeToolResultText returns the text of a tool result, which is a string
// or a list of blocks; a block that is not text is named, as "[image]".
func claudeToolResultText(content json.RawMessage) string {
	var text string
	if json.Unmarshal(content, &text) == nil {
		return text
	}

	parts := make([]string, 0, 1)
	for _, block := range claudeBlocks(content) {
		if block.Type == "text" {
			parts = append(parts, block.Text)
		} else if block.Type != "" {
			parts = append(parts, "["+block.Type+"]")
		}
	}
	return strings.Join(parts, "\n")
}

// maskJSONStrings applies mask to every string of a decoded JSON value, the
// keys of its objects included.
func maskJSONStrings(value any, mask func(string) string) any {
	switch v := value.(type) {
	case string:
		return mask(v)
	case []any:
		for i := range v {
			v[i] = maskJSONStrings(v[i], mask)
		}
		return v
	case map[string]any:
		masked := make(map[string]any, len(v))
		for key, item := range v {
			masked[mask(key)] = maskJSONStrings(item, mask)
		}
		return masked
	default:
		return value
	}
}

// splitAgentText splits text into lines, without the empty lines around it.
func splitAgentText(text string) []string {
	text = strings.Trim(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

// truncateChars cuts s to max characters, marking the cut.
func truncateChars(s string, max int) string {
	count := 0
	for i := range s {
		if count == max {
			return s[:i] + "..."
		}
		count++
	}
	return s
}

// agentDuration renders milliseconds as a duration with a tenth of a second.
func agentDuration(ms float64) string {
	return (time.Duration(ms) * time.Millisecond).Round(100 * time.Millisecond).String()
}
