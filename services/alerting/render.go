package alerting

import (
	"bytes"
	"embed"
	"encoding/json"
	htmltemplate "html/template"
	"strings"
	"text/template"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

//go:embed templates/*.tmpl
var builtinTemplates embed.FS

// builtinBody reads the embedded default template of a channel.
func builtinBody(name string) string {
	body, err := builtinTemplates.ReadFile("templates/" + name)
	if err != nil {
		// The file list is fixed at build time; a missing file is a
		// programming error, not a runtime condition.
		panic("alerting: missing built-in template " + name)
	}
	return string(body)
}

// Render executes body with payload using the engine the channel asked for.
// A blank body falls back to the channel default.
func Render(channel Channel, body string, payload Payload) (string, error) {
	if strings.TrimSpace(body) == "" {
		body = channel.DefaultBody()
	}

	var out bytes.Buffer

	switch channel.BodyFormat() {
	case BodyFormatHTML:
		tpl, err := htmltemplate.New(string(channel.Type())).Parse(body)
		if err != nil {
			return "", common_errors.NewValidationError("invalid message template: " + err.Error())
		}
		if err = tpl.Execute(&out, payload); err != nil {
			return "", common_errors.NewValidationError("can not render message template: " + err.Error())
		}
	case BodyFormatJSON:
		tpl, err := template.New(string(channel.Type())).Parse(body)
		if err != nil {
			return "", common_errors.NewValidationError("invalid message template: " + err.Error())
		}
		if err = tpl.Execute(&out, jsonTemplatePayload(payload)); err != nil {
			return "", common_errors.NewValidationError("can not render message template: " + err.Error())
		}
	default:
		tpl, err := template.New(string(channel.Type())).Parse(body)
		if err != nil {
			return "", common_errors.NewValidationError("invalid message template: " + err.Error())
		}
		if err = tpl.Execute(&out, payload); err != nil {
			return "", common_errors.NewValidationError("can not render message template: " + err.Error())
		}
	}

	if strings.TrimSpace(out.String()) == "" {
		return "", common_errors.NewValidationError("rendered message is empty")
	}

	return out.String(), nil
}

// ValidateBody parses a custom template so the API can reject broken
// templates at save time instead of at send time.
func ValidateBody(channel Channel, body string) error {
	if strings.TrimSpace(body) == "" {
		return nil
	}
	var err error
	switch channel.BodyFormat() {
	case BodyFormatHTML:
		_, err = htmltemplate.New(string(channel.Type())).Parse(body)
	default:
		_, err = template.New(string(channel.Type())).Parse(body)
	}
	if err != nil {
		return common_errors.NewValidationError("invalid message template: " + err.Error())
	}
	return nil
}

func jsonTemplatePayload(payload Payload) Payload {
	payload.Name = jsonEscapeString(payload.Name)
	payload.Author = jsonEscapeString(payload.Author)
	payload.Color = jsonEscapeString(payload.Color)
	payload.Task.ID = jsonEscapeString(payload.Task.ID)
	payload.Task.URL = jsonEscapeString(payload.Task.URL)
	payload.Task.Result = jsonEscapeString(payload.Task.Result)
	payload.Task.Desc = jsonEscapeString(payload.Task.Desc)
	payload.Task.Version = jsonEscapeString(payload.Task.Version)
	payload.Task.Duration = jsonEscapeString(payload.Task.Duration)
	payload.Task.Trigger = jsonEscapeString(payload.Task.Trigger)
	payload.Chat.ID = jsonEscapeString(payload.Chat.ID)
	payload.Chat.ThreadID = jsonEscapeString(payload.Chat.ThreadID)
	payload.Project.Name = jsonEscapeString(payload.Project.Name)
	payload.Playbook = jsonEscapeString(payload.Playbook)
	payload.ScheduleName = jsonEscapeString(payload.ScheduleName)
	return payload
}

func jsonEscapeString(s string) string {
	quoted, err := json.Marshal(s)
	if err != nil || len(quoted) < 2 {
		return s
	}
	return string(quoted[1 : len(quoted)-1])
}
