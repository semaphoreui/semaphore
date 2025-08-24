package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// GotifyProvider implements NotificationProvider for Gotify
type GotifyProvider struct {
	config *GotifyConfig
}

// NewGotifyProvider creates a new Gotify notification provider
func NewGotifyProvider(config *GotifyConfig) *GotifyProvider {
	return &GotifyProvider{
		config: config,
	}
}

// Send sends a notification to Gotify
func (gp *GotifyProvider) Send(ctx context.Context, message *NotificationMessage) error {
	if !gp.IsEnabled() {
		return fmt.Errorf("gotify provider is not enabled")
	}

	if gp.config.URL == "" {
		return fmt.Errorf("gotify URL not configured")
	}

	if gp.config.Token == "" {
		return fmt.Errorf("gotify token not configured")
	}

	// Prepare message content
	content := fmt.Sprintf("Task #%s - %s\n", message.TaskID, message.Status)
	content += fmt.Sprintf("Author: %s\n", message.Author)
	if message.Version != "" {
		content += fmt.Sprintf("Version: %s\n", message.Version)
	}
	if message.Description != "" {
		content += fmt.Sprintf("Description: %s\n", message.Description)
	}
	content += fmt.Sprintf("Link: %s", message.TaskURL)

	// Determine priority
	priority := gp.config.Priority
	if priority == 0 {
		priority = gp.getPriorityForStatus(message.Status)
	}

	// Create payload
	payload := map[string]interface{}{
		"title":    fmt.Sprintf("%s - %s", message.ProjectName, message.Status),
		"message":  content,
		"priority": priority,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal gotify payload: %w", err)
	}

	// Prepare URL
	url := strings.TrimRight(gp.config.URL, "/")
	url = fmt.Sprintf("%s/message?token=%s", url, gp.config.Token)

	// Send request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send gotify notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gotify API returned status code: %d", resp.StatusCode)
	}

	return nil
}

// getPriorityForStatus returns the appropriate priority for the given status
func (gp *GotifyProvider) getPriorityForStatus(status string) int {
	switch status {
	case "success":
		return 2 // Low priority for success
	case "error", "fail":
		return 8 // High priority for errors
	case "running":
		return 4 // Normal priority for running
	case "waiting":
		return 2 // Low priority for waiting
	case "stopping", "stopped":
		return 4 // Normal priority for stopped
	default:
		return 5 // Default normal priority
	}
}

// GetName returns the provider name
func (gp *GotifyProvider) GetName() string {
	return "gotify"
}

// IsEnabled returns whether the provider is enabled
func (gp *GotifyProvider) IsEnabled() bool {
	return gp.config != nil && gp.config.Enabled && gp.config.URL != "" && gp.config.Token != ""
}

// Validate validates the provider configuration
func (gp *GotifyProvider) Validate() error {
	if gp.config == nil {
		return fmt.Errorf("gotify configuration is nil")
	}

	if !gp.config.Enabled {
		return nil // Skip validation if not enabled
	}

	if gp.config.URL == "" {
		return fmt.Errorf("gotify URL is required")
	}

	if gp.config.Token == "" {
		return fmt.Errorf("gotify token is required")
	}

	return nil
}