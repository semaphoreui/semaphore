package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SlackProvider implements NotificationProvider for Slack
type SlackProvider struct {
	config *SlackConfig
}

// NewSlackProvider creates a new Slack notification provider
func NewSlackProvider(config *SlackConfig) *SlackProvider {
	return &SlackProvider{
		config: config,
	}
}

// Send sends a notification to Slack
func (sp *SlackProvider) Send(ctx context.Context, message *NotificationMessage) error {
	if !sp.IsEnabled() {
		return fmt.Errorf("slack provider is not enabled")
	}

	if sp.config.WebhookURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}

	// Map status to color
	color := sp.getColorForStatus(message.Status)

	// Create Slack attachment
	attachment := map[string]interface{}{
		"color":      color,
		"title":      fmt.Sprintf("Task #%s - %s", message.TaskID, message.Status),
		"title_link": message.TaskURL,
		"text":       message.Description,
		"fields": []map[string]interface{}{
			{
				"title": "Project",
				"value": message.ProjectName,
				"short": true,
			},
			{
				"title": "Author",
				"value": message.Author,
				"short": true,
			},
		},
		"footer":    "Semaphore",
		"ts":        message.Timestamp.Unix(),
	}

	// Add version field if available
	if message.Version != "" {
		fields := attachment["fields"].([]map[string]interface{})
		fields = append(fields, map[string]interface{}{
			"title": "Version",
			"value": message.Version,
			"short": true,
		})
		attachment["fields"] = fields
	}

	// Create payload
	payload := map[string]interface{}{
		"text":        fmt.Sprintf("Semaphore notification for %s", message.ProjectName),
		"attachments": []map[string]interface{}{attachment},
	}

	// Add channel if specified
	if sp.config.Channel != "" {
		payload["channel"] = sp.config.Channel
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	// Send request
	resp, err := http.Post(sp.config.WebhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status code: %d", resp.StatusCode)
	}

	return nil
}

// getColorForStatus returns the appropriate color for the given status
func (sp *SlackProvider) getColorForStatus(status string) string {
	switch status {
	case "success":
		return "good"
	case "error", "fail":
		return "danger"
	case "running":
		return "#333CFF"
	case "waiting":
		return "#FFFC33"
	case "stopping":
		return "#BEBEBE"
	case "stopped":
		return "#5B5B5B"
	default:
		return "#808080"
	}
}

// GetName returns the provider name
func (sp *SlackProvider) GetName() string {
	return "slack"
}

// IsEnabled returns whether the provider is enabled
func (sp *SlackProvider) IsEnabled() bool {
	return sp.config != nil && sp.config.Enabled && sp.config.WebhookURL != ""
}

// Validate validates the provider configuration
func (sp *SlackProvider) Validate() error {
	if sp.config == nil {
		return fmt.Errorf("slack configuration is nil")
	}

	if !sp.config.Enabled {
		return nil // Skip validation if not enabled
	}

	if sp.config.WebhookURL == "" {
		return fmt.Errorf("slack webhook URL is required")
	}

	return nil
}