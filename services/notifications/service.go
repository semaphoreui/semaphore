package notifications

import (
	"context"
	"text/template"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

// NotificationService defines the interface for all notification services
type NotificationService interface {
	// Send sends a notification using the configured service
	Send(ctx context.Context, notification *Notification) error
	
	// GetName returns the service name for identification
	GetName() string
	
	// IsConfigured returns true if the service is properly configured
	IsConfigured() bool
	
	// SupportsProjectOverride returns true if the service supports project-specific settings
	SupportsProjectOverride() bool
}

// Notification contains all the information needed to send a notification
type Notification struct {
	// Template information
	TemplateName string
	TemplateID   int
	ProjectID    int
	
	// Task information
	TaskID      int
	TaskStatus  task_logger.TaskStatus
	TaskMessage string
	TaskVersion string
	TaskURL     string
	
	// User information
	AuthorName string
	
	// Recipients (for services that support multiple recipients)
	Recipients []string
	
	// Project-specific overrides
	ProjectConfig map[string]interface{}
	
	// Additional context
	SuppressSuccessAlerts bool
}

// ServiceConfig holds configuration for a notification service
type ServiceConfig struct {
	Enabled bool                   `json:"enabled"`
	Token   string                 `json:"token,omitempty"`
	URL     string                 `json:"url,omitempty"`
	Channel string                 `json:"channel,omitempty"`
	Config  map[string]interface{} `json:"config,omitempty"`
}

// NotificationManager manages all notification services
type NotificationManager struct {
	services map[string]NotificationService
	template *template.Template
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		services: make(map[string]NotificationService),
	}
}

// RegisterService registers a notification service
func (nm *NotificationManager) RegisterService(service NotificationService) {
	nm.services[service.GetName()] = service
}

// GetService returns a notification service by name
func (nm *NotificationManager) GetService(name string) (NotificationService, bool) {
	service, exists := nm.services[name]
	return service, exists
}

// GetConfiguredServices returns all configured notification services
func (nm *NotificationManager) GetConfiguredServices() []NotificationService {
	var configured []NotificationService
	for _, service := range nm.services {
		if service.IsConfigured() {
			configured = append(configured, service)
		}
	}
	return configured
}

// SendNotification sends a notification to all configured services
func (nm *NotificationManager) SendNotification(ctx context.Context, notification *Notification) error {
	services := nm.GetConfiguredServices()
	
	for _, service := range services {
		// Skip if success alerts are suppressed and task succeeded
		if notification.SuppressSuccessAlerts && 
		   notification.TaskStatus == task_logger.TaskSuccessStatus &&
		   service.GetName() != "email" { // Email only sends on failure anyway
			continue
		}
		
		if err := service.Send(ctx, notification); err != nil {
			// Log error but continue with other services
			// TODO: Add proper logging
			continue
		}
	}
	
	return nil
}

// SendTestNotification sends a test notification to all configured services
func (nm *NotificationManager) SendTestNotification(ctx context.Context, project db.Project) error {
	notification := &Notification{
		TemplateName: "Test Notification",
		TemplateID:   0,
		ProjectID:    project.ID,
		TaskID:       0,
		TaskStatus:   task_logger.TaskSuccessStatus,
		TaskMessage:  "This is a test notification",
		TaskVersion:  "",
		TaskURL:      "",
		AuthorName:   "—",
		Recipients:   []string{},
		ProjectConfig: make(map[string]interface{}),
		SuppressSuccessAlerts: false,
	}
	
	return nm.SendNotification(ctx, notification)
}