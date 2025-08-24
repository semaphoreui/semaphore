package notifications

import (
	"context"
	"time"
)

// NotificationProvider defines the interface that all notification providers must implement
type NotificationProvider interface {
	// Send sends a notification with the given message and context
	Send(ctx context.Context, message *NotificationMessage) error
	// GetName returns the name/type of the notification provider
	GetName() string
	// IsEnabled returns whether this provider is enabled
	IsEnabled() bool
	// Validate validates the provider configuration
	Validate() error
}

// NotificationMessage represents a notification message to be sent
type NotificationMessage struct {
	Title       string
	Content     string
	Color       string
	ProjectName string
	TaskID      string
	TaskURL     string
	Status      string
	Author      string
	Version     string
	Description string
	Timestamp   time.Time
}

// NotificationConfig represents the base configuration for all notification providers
type NotificationConfig struct {
	Enabled bool   `json:"enabled" env:"ENABLED"`
	Token   string `json:"token,omitempty" env:"TOKEN"`
	Channel string `json:"channel,omitempty" env:"CHANNEL"`
}

// TelegramConfig represents Telegram notification configuration
type TelegramConfig struct {
	NotificationConfig
	ChatID string `json:"chat_id,omitempty" env:"CHAT_ID"`
}

// SlackConfig represents Slack notification configuration
type SlackConfig struct {
	NotificationConfig
	WebhookURL string `json:"webhook_url,omitempty" env:"WEBHOOK_URL"`
}

// GotifyConfig represents Gotify notification configuration
type GotifyConfig struct {
	NotificationConfig
	URL      string `json:"url,omitempty" env:"URL"`
	Priority int    `json:"priority,omitempty" env:"PRIORITY"`
}

// DingtalkConfig represents Dingtalk notification configuration
type DingtalkConfig struct {
	NotificationConfig
	WebhookURL string `json:"webhook_url,omitempty" env:"WEBHOOK_URL"`
	Secret     string `json:"secret,omitempty" env:"SECRET"`
}

// NotificationsConfig represents the complete notifications configuration
type NotificationsConfig struct {
	Telegram *TelegramConfig `json:"telegram,omitempty"`
	Slack    *SlackConfig    `json:"slack,omitempty"`
	Gotify   *GotifyConfig   `json:"gotify,omitempty"`
	Dingtalk *DingtalkConfig `json:"dingtalk,omitempty"`
}

// NotificationManager manages all notification providers
type NotificationManager struct {
	providers []NotificationProvider
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		providers: make([]NotificationProvider, 0),
	}
}

// AddProvider adds a notification provider to the manager
func (nm *NotificationManager) AddProvider(provider NotificationProvider) {
	nm.providers = append(nm.providers, provider)
}

// SendNotification sends a notification to all enabled providers
func (nm *NotificationManager) SendNotification(ctx context.Context, message *NotificationMessage) {
	for _, provider := range nm.providers {
		if provider.IsEnabled() {
			go func(p NotificationProvider) {
				if err := p.Send(ctx, message); err != nil {
					// Log error but don't fail the entire process
					// TODO: Add proper logging
				}
			}(provider)
		}
	}
}

// GetProviders returns all registered providers
func (nm *NotificationManager) GetProviders() []NotificationProvider {
	return nm.providers
}