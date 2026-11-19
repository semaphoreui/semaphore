package db_lib

import (
	"cmp"
	"encoding/json"
	"slices"
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

// secretMask masks the secrets of a launch in what the agent prints.
type secretMask struct {
	// apply replaces every secret with maskedSecret.
	apply func(string) string
	// multiline are the secrets that span lines: a key, a certificate. Text
	// is masked whole before it is split into lines, never line by line, or
	// no line would match them.
	multiline []string
}

// newSecretMask makes the mask of secrets. Where two secrets overlap the
// longer one wins: a secret that is the first line of another must not leave
// the rest of that one in the log.
func newSecretMask(secrets []string) secretMask {
	sorted := slices.Clone(secrets)
	slices.SortStableFunc(sorted, func(a, b string) int { return cmp.Compare(len(b), len(a)) })

	var m secretMask
	pairs := make([]string, 0, len(sorted)*2)
	for _, s := range sorted {
		if s == "" {
			continue
		}
		pairs = append(pairs, s, maskedSecret)
		if strings.Contains(s, "\n") {
			m.multiline = append(m.multiline, s)
		}
	}

	if len(pairs) == 0 {
		m.apply = func(s string) string { return s }
	} else {
		m.apply = strings.NewReplacer(pairs...).Replace
	}
	return m
}

// newSecretMasker returns a function replacing every secret with maskedSecret.
func newSecretMasker(secrets []string) func(string) string {
	return newSecretMask(secrets).apply
}

// lineCut returns where text that is still being streamed can be cut: the
// lines before the cut can be masked and logged, the rest waits for more
// text. It is the last line break no multiline secret spans — neither a
// whole one nor one whose beginning ends the text, which may be completed by
// what comes next. -1 when there is no such line break.
func (m secretMask) lineCut(text string) int {
	for cut := strings.LastIndexByte(text, '\n'); cut >= 0; cut = strings.LastIndexByte(text[:cut], '\n') {
		if !m.spans(text, cut) {
			return cut
		}
	}
	return -1
}

// spans tells whether a multiline secret, or the beginning of one at the end
// of text, covers the byte text[at].
func (m secretMask) spans(text string, at int) bool {
	for _, secret := range m.multiline {
		for start := max(0, at-len(secret)+1); start <= at; start++ {
			rest := text[start:]
			if strings.HasPrefix(rest, secret) || strings.HasPrefix(secret, rest) {
				return true
			}
		}
	}
	return false
}

// maskedLines masks text whole and splits it into lines.
func maskedLines(mask func(string) string, text string) []string {
	lines := strings.Split(mask(text), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, "\r")
	}
	return lines
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
	// cut is where the streamed text can be cut into lines (secretMask.lineCut).
	cut func(string) int

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

// textEvents makes an event of every line. The lines are masked already: a
// secret of several lines is found only in the text before it is split.
func (o *claudeStreamOutput) textEvents(agent string, lines []string) []agentlog.Event {
	events := make([]agentlog.Event, 0, len(lines))
	for _, line := range lines {
		events = append(events, agentlog.Event{Kind: agentlog.KindText, Agent: agent, Text: line})
	}
	return events
}

// endText returns what is left of the streamed text, if any: the unfinished
// last line, and the lines held back while they might begin a secret.
func (o *claudeStreamOutput) endText() []agentlog.Event {
	if o.partial == "" {
		return nil
	}
	text := o.partial
	o.partial = ""

	lines := maskedLines(o.mask, text)
	// A line break ends the line before it; it does not open an empty one.
	if len(lines) > 1 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return o.textEvents("", lines)
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

	// The complete lines are logged, masked together: a line that may begin
	// a secret of several lines waits until the secret is whole or is not
	// one.
	cut := o.cut(o.partial)
	if cut < 0 {
		return nil, nil
	}
	text := o.partial[:cut]
	o.partial = o.partial[cut+1:]
	return o.textEvents("", maskedLines(o.mask, text)), nil
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
			if text := splitAgentText(o.mask(block.Text)); len(text) > 0 {
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
		events = append(events, o.textEvents("", splitAgentText(o.mask(event.Result)))...)
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
