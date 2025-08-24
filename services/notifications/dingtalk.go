package notifications

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// DingtalkProvider implements NotificationProvider for Dingtalk
type DingtalkProvider struct {
	config *DingtalkConfig
}

// NewDingtalkProvider creates a new Dingtalk notification provider
func NewDingtalkProvider(config *DingtalkConfig) *DingtalkProvider {
	return &DingtalkProvider{
		config: config,
	}
}

// Send sends a notification to Dingtalk
func (dp *DingtalkProvider) Send(ctx context.Context, message *NotificationMessage) error {
	if !dp.IsEnabled() {
		return fmt.Errorf("dingtalk provider is not enabled")
	}

	if dp.config.WebhookURL == "" {
		return fmt.Errorf("dingtalk webhook URL not configured")
	}

	// Prepare message content
	content := fmt.Sprintf("**%s**\n\n", message.ProjectName)
	content += fmt.Sprintf("**Task:** #%s - %s\n", message.TaskID, message.Status)
	content += fmt.Sprintf("**Author:** %s\n", message.Author)
	if message.Version != "" {
		content += fmt.Sprintf("**Version:** %s\n", message.Version)
	}
	if message.Description != "" {
		content += fmt.Sprintf("**Description:** %s\n", message.Description)
	}
	content += fmt.Sprintf("**Link:** [View Task](%s)\n", message.TaskURL)
	content += fmt.Sprintf("**Time:** %s", message.Timestamp.Format("2006-01-02 15:04:05"))

	// Create payload
	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": fmt.Sprintf("%s - %s", message.ProjectName, message.Status),
			"text":  content,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal dingtalk payload: %w", err)
	}

	// Prepare URL with signature if secret is provided
	requestURL := dp.config.WebhookURL
	if dp.config.Secret != "" {
		timestamp := time.Now().UnixNano() / 1000000
		sign := dp.generateSignature(timestamp, dp.config.Secret)
		
		u, err := url.Parse(requestURL)
		if err != nil {
			return fmt.Errorf("failed to parse webhook URL: %w", err)
		}
		
		query := u.Query()
		query.Set("timestamp", strconv.FormatInt(timestamp, 10))
		query.Set("sign", sign)
		u.RawQuery = query.Encode()
		requestURL = u.String()
	}

	// Send request
	resp, err := http.Post(requestURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send dingtalk notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk webhook returned status code: %d", resp.StatusCode)
	}

	return nil
}

// generateSignature generates HMAC-SHA256 signature for Dingtalk webhook
func (dp *DingtalkProvider) generateSignature(timestamp int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// GetName returns the provider name
func (dp *DingtalkProvider) GetName() string {
	return "dingtalk"
}

// IsEnabled returns whether the provider is enabled
func (dp *DingtalkProvider) IsEnabled() bool {
	return dp.config != nil && dp.config.Enabled && dp.config.WebhookURL != ""
}

// Validate validates the provider configuration
func (dp *DingtalkProvider) Validate() error {
	if dp.config == nil {
		return fmt.Errorf("dingtalk configuration is nil")
	}

	if !dp.config.Enabled {
		return nil // Skip validation if not enabled
	}

	if dp.config.WebhookURL == "" {
		return fmt.Errorf("dingtalk webhook URL is required")
	}

	return nil
}