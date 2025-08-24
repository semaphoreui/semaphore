package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
)

const telegramTemplate = `
🔔 *{{.ProjectName}}*

{{if eq .Status "success"}}✅{{else if eq .Status "error"}}❌{{else if eq .Status "running"}}🔄{{else if eq .Status "waiting"}}⏳{{else if eq .Status "stopped"}}⏹️{{else}}ℹ️{{end}} **{{.Status | title}}**

**Task:** [#{{.TaskID}}]({{.TaskURL}})
**Author:** {{.Author}}
{{if .Version}}**Version:** {{.Version}}{{end}}
{{if .Description}}**Description:** {{.Description}}{{end}}

*{{.Timestamp.Format "2006-01-02 15:04:05"}}*
`

// TelegramProvider implements NotificationProvider for Telegram
type TelegramProvider struct {
	config *TelegramConfig
}

// NewTelegramProvider creates a new Telegram notification provider
func NewTelegramProvider(config *TelegramConfig) *TelegramProvider {
	return &TelegramProvider{
		config: config,
	}
}

// Send sends a notification to Telegram
func (tp *TelegramProvider) Send(ctx context.Context, message *NotificationMessage) error {
	if !tp.IsEnabled() {
		return fmt.Errorf("telegram provider is not enabled")
	}

	// Use channel from config, fallback to chat_id for backward compatibility
	chatID := tp.config.Channel
	if chatID == "" {
		chatID = tp.config.ChatID
	}

	if chatID == "" {
		return fmt.Errorf("telegram chat ID or channel not configured")
	}

	// Parse template
	tmpl, err := template.New("telegram").Funcs(template.FuncMap{
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return string(s[0]-32) + s[1:]
		},
	}).Parse(telegramTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse telegram template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, message); err != nil {
		return fmt.Errorf("failed to execute telegram template: %w", err)
	}

	// Prepare payload
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       buf.String(),
		"parse_mode": "Markdown",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	// Send request
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tp.config.Token)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send telegram notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status code: %d", resp.StatusCode)
	}

	return nil
}

// GetName returns the provider name
func (tp *TelegramProvider) GetName() string {
	return "telegram"
}

// IsEnabled returns whether the provider is enabled
func (tp *TelegramProvider) IsEnabled() bool {
	return tp.config != nil && tp.config.Enabled && tp.config.Token != ""
}

// Validate validates the provider configuration
func (tp *TelegramProvider) Validate() error {
	if tp.config == nil {
		return fmt.Errorf("telegram configuration is nil")
	}

	if !tp.config.Enabled {
		return nil // Skip validation if not enabled
	}

	if tp.config.Token == "" {
		return fmt.Errorf("telegram token is required")
	}

	if tp.config.Channel == "" && tp.config.ChatID == "" {
		return fmt.Errorf("telegram channel or chat_id is required")
	}

	return nil
}