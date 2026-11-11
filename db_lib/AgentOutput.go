package db_lib

import (
	"encoding/json"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/agentlog"
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
//
// It writes them to the log as agentlog events, one per line: what the CLI
// reports is too much and too much its own to be the log, and a client can
// not tell a tool call from the agent's text in plain lines. Whatever is not
// an event of the CLI goes to the log as it came.
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
	// init is the last init event logged: the CLI repeats it at every turn.
	init string
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
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`

	// tool_result
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content"`
	IsError   bool            `json:"is_error"`

	// tool_reference, a block of a tool search's result
	ToolName string `json:"tool_name"`
}

type claudeMessageEvent struct {
	Message struct {
		ID string `json:"id"`
		// Content is a list of blocks, or a string for a plain user message.
		Content json.RawMessage `json:"content"`
	} `json:"message"`
	// ParentToolUseID is set on the messages of a subagent: the tool call
	// which started it.
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
		return append(encodeAgentEvents(o.endText()), o.mask(line))
	}

	var (
		events []agentlog.Event
		err    error
	)

	switch head.Type {
	case "system":
		events, err = o.system(head.Subtype, raw)
	case "stream_event":
		events, err = o.streamEvent(raw)
	case "assistant":
		events, err = o.assistant(raw)
	case "user":
		events, err = o.user(raw)
	case "result":
		events, err = o.result(raw)
	}

	if err != nil {
		// An event of a shape this code does not know is shown as it came
		// rather than hidden.
		return append(encodeAgentEvents(o.endText()), o.mask(line))
	}

	return encodeAgentEvents(events)
}

func (o *claudeStreamOutput) Flush() []string {
	return encodeAgentEvents(o.endText())
}

func encodeAgentEvents(events []agentlog.Event) []string {
	lines := make([]string, 0, len(events))
	for _, event := range events {
		lines = append(lines, event.Line())
	}
	return lines
}

// textEvents makes an event of every line, with the secrets masked.
func (o *claudeStreamOutput) textEvents(agent string, lines []string) []agentlog.Event {
	events := make([]agentlog.Event, 0, len(lines))
	for _, line := range lines {
		events = append(events, agentlog.Event{Kind: agentlog.KindText, Agent: agent, Text: o.mask(line)})
	}
	return events
}

// endText returns the unfinished line of the streamed text, if any.
func (o *claudeStreamOutput) endText() []agentlog.Event {
	if o.partial == "" {
		return nil
	}
	line := strings.TrimRight(o.partial, "\r")
	o.partial = ""
	return o.textEvents("", []string{line})
}

func (o *claudeStreamOutput) system(subtype string, raw []byte) ([]agentlog.Event, error) {
	if subtype != "init" && subtype != "api_retry" {
		return nil, nil
	}

	var event claudeSystemEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, err
	}

	events := o.endText()

	if subtype == "api_retry" {
		return append(events, agentlog.Event{
			Kind:       agentlog.KindRetry,
			Attempt:    event.Attempt,
			MaxRetries: event.MaxRetries,
			DelayMs:    event.RetryDelayMs,
			Text:       o.mask(event.Error),
		}), nil
	}

	init := agentlog.Event{Kind: agentlog.KindInit, Model: o.mask(event.Model)}
	for _, server := range event.McpServers {
		init.Servers = append(init.Servers, agentlog.Server{Name: o.mask(server.Name), Status: o.mask(server.Status)})
	}

	// The CLI opens every turn with it; it is news only when it changed.
	if line := init.Line(); line != o.init {
		o.init = line
		events = append(events, init)
	}
	return events, nil
}

// streamEvent logs the agent's text line by line while it is being written.
func (o *claudeStreamOutput) streamEvent(raw []byte) ([]agentlog.Event, error) {
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
	return o.textEvents("", lines), nil
}

func (o *claudeStreamOutput) assistant(raw []byte) ([]agentlog.Event, error) {
	var event claudeMessageEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, err
	}

	agent, events := o.messageStart(event)

	streamed := agent == "" && event.Message.ID != "" && event.Message.ID == o.streamedID

	for _, block := range claudeBlocks(event.Message.Content) {
		switch block.Type {
		case "text":
			if streamed {
				continue
			}
			if text := splitAgentText(block.Text); len(text) > 0 {
				events = append(events, o.textEvents(agent, text)...)
				if agent == "" {
					o.said = true
				}
			}
		case "tool_use":
			events = append(events, o.toolCall(agent, block))
		}
	}

	return events, nil
}

func (o *claudeStreamOutput) user(raw []byte) ([]agentlog.Event, error) {
	var event claudeMessageEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, err
	}

	agent, events := o.messageStart(event)

	for _, block := range claudeBlocks(event.Message.Content) {
		if block.Type != "tool_result" {
			continue
		}

		result := agentlog.NewResult(block.ToolUseID, claudeToolResultText(block.Content), block.IsError, o.mask)
		result.Agent = agent
		events = append(events, result)
	}

	return events, nil
}

// messageStart returns the subagent a message belongs to, empty for the agent
// itself, and the line of the agent's text the message ends. A message of a
// subagent comes at any moment and ends nothing.
func (o *claudeStreamOutput) messageStart(event claudeMessageEvent) (string, []agentlog.Event) {
	if event.ParentToolUseID != nil && *event.ParentToolUseID != "" {
		return o.mask(*event.ParentToolUseID), nil
	}
	return "", o.endText()
}

func (o *claudeStreamOutput) result(raw []byte) ([]agentlog.Event, error) {
	var event claudeResultEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, err
	}

	events := o.endText()

	// The result repeats the last message; it is the only text when the run
	// printed nothing before it (the json output format, an early error).
	if !o.said {
		events = append(events, o.textEvents("", splitAgentText(event.Result))...)
	}

	done := agentlog.Event{
		Kind:       agentlog.KindDone,
		Error:      event.IsError,
		Turns:      event.NumTurns,
		DurationMs: event.DurationMs,
		CostUSD:    event.TotalCostUSD,
	}
	if event.Subtype != "success" {
		done.Subtype = o.mask(event.Subtype)
	}
	for _, text := range event.Errors {
		done.Errors = append(done.Errors, o.mask(text))
	}

	return append(events, done), nil
}

// toolCall makes the event of a tool call. The arguments are masked value by
// value: a secret with a character JSON escapes would not be found in the
// encoded text.
func (o *claudeStreamOutput) toolCall(agent string, block claudeBlock) agentlog.Event {
	call := agentlog.NewTool(block.ID, block.Name, block.Input, o.mask)
	call.Agent = agent
	return call
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
// or a list of blocks. A block that is not text is named: the tool a tool
// search found, "[image]" for an image.
func claudeToolResultText(content json.RawMessage) string {
	var text string
	if json.Unmarshal(content, &text) == nil {
		return text
	}

	parts := make([]string, 0, 1)
	for _, block := range claudeBlocks(content) {
		switch {
		case block.Type == "text":
			parts = append(parts, block.Text)
		case block.Type == "tool_reference" && block.ToolName != "":
			parts = append(parts, block.ToolName)
		case block.Type != "":
			parts = append(parts, "["+block.Type+"]")
		}
	}
	return strings.Join(parts, "\n")
}

// splitAgentText splits text into lines, without the empty lines around it.
func splitAgentText(text string) []string {
	text = strings.Trim(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return strings.Split(text, "\n")
}
