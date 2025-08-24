package notifications

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

//go:embed templates/gotify.tmpl
var gotifyTemplate embed.FS

type GotifyService struct {
	config *util.NotificationServiceConfig
	tmpl   *template.Template
}

type gotifyAlert struct {
	Name   string
	Author string
	Color  string
	Task   gotifyTask
}

type gotifyTask struct {
	ID      string
	URL     string
	Result  string
	Desc    string
	Version string
}

func NewGotifyService(config *util.NotificationServiceConfig) *GotifyService {
	tmpl, err := template.ParseFS(gotifyTemplate, "templates/gotify.tmpl")
	if err != nil {
		// Fallback to legacy template system if new template not found
		tmpl = nil
	}
	
	return &GotifyService{
		config: config,
		tmpl:   tmpl,
	}
}

func (g *GotifyService) GetName() string {
	return "gotify"
}

func (g *GotifyService) IsConfigured() bool {
	if g.config != nil && g.config.Enabled && g.config.URL != "" && g.config.Token != "" {
		return true
	}
	
	// Fallback to legacy configuration
	return util.Config.GotifyAlert && util.Config.GotifyUrl != "" && util.Config.GotifyToken != ""
}

func (g *GotifyService) SupportsProjectOverride() bool {
	return true // Gotify supports project-specific URLs and tokens
}

func (g *GotifyService) Send(ctx context.Context, notification *Notification) error {
	var serverURL, token string
	
	// Use new configuration if available
	if g.config != nil && g.config.Enabled && g.config.URL != "" && g.config.Token != "" {
		serverURL = g.config.URL
		token = g.config.Token
		
		// Check for project-specific overrides
		if projectURL, exists := notification.ProjectConfig["gotify_url"]; exists {
			if urlStr, ok := projectURL.(string); ok && urlStr != "" {
				serverURL = urlStr
			}
		}
		if projectToken, exists := notification.ProjectConfig["gotify_token"]; exists {
			if tokenStr, ok := projectToken.(string); ok && tokenStr != "" {
				token = tokenStr
			}
		}
	} else {
		// Fallback to legacy configuration
		if !util.Config.GotifyAlert {
			return fmt.Errorf("gotify notifications not enabled")
		}
		
		serverURL = util.Config.GotifyUrl
		token = util.Config.GotifyToken
	}
	
	if serverURL == "" {
		return fmt.Errorf("gotify server URL not configured")
	}
	
	if token == "" {
		return fmt.Errorf("gotify token not configured")
	}
	
	// Create alert data
	alert := gotifyAlert{
		Name:   notification.TemplateName,
		Author: notification.AuthorName,
		Color:  g.getColor(notification.TaskStatus),
		Task: gotifyTask{
			ID:      strconv.Itoa(notification.TaskID),
			URL:     notification.TaskURL,
			Result:  notification.TaskStatus.Format(),
			Version: notification.TaskVersion,
			Desc:    notification.TaskMessage,
		},
	}
	
	// Generate message body
	body := bytes.NewBufferString("")
	var err error
	
	if g.tmpl != nil {
		err = g.tmpl.Execute(body, alert)
	} else {
		// Fallback to legacy template
		err = g.executeLegacyTemplate(body, alert)
	}
	
	if err != nil {
		return fmt.Errorf("failed to generate gotify message: %w", err)
	}
	
	if body.Len() == 0 {
		return fmt.Errorf("gotify message body is empty")
	}
	
	// Send to Gotify server
	url := fmt.Sprintf("%s/message?token=%s", serverURL, token)
	
	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return fmt.Errorf("failed to send gotify message: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("gotify server returned status code: %d", resp.StatusCode)
	}
	
	return nil
}

func (g *GotifyService) getColor(status task_logger.TaskStatus) string {
	// Gotify doesn't use colors in the same way as Slack
	return ""
}

func (g *GotifyService) executeLegacyTemplate(body *bytes.Buffer, alert gotifyAlert) error {
	// Simple JSON message format for Gotify
	message := fmt.Sprintf(`{
		"title": "Task: %s",
		"message": "Author: %s\nStatus: %s\nMessage: %s\n\nView Task: %s",
		"priority": %d
	}`, alert.Name, alert.Author, alert.Task.Result, alert.Task.Desc, alert.Task.URL, g.getPriority(alert.Task.Result))
	
	body.WriteString(message)
	return nil
}

func (g *GotifyService) getPriority(status string) int {
	// Set priority based on task status
	switch status {
	case "FAILED":
		return 8 // High priority for failures
	case "SUCCESS":
		return 5 // Normal priority for success
	case "RUNNING":
		return 3 // Lower priority for running tasks
	default:
		return 5 // Default normal priority
	}
}