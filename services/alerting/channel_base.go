package alerting

import (
	"context"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

// channelBase carries the static description shared by every channel so a
// new implementation only writes the parts that differ.
type channelBase struct {
	typ           db.AlertType
	title         string
	icon          string
	fields        []Field
	defaultEvents db.AlertEvents
	format        BodyFormat
	templateFile  string
}

func (c channelBase) Type() db.AlertType { return c.typ }
func (c channelBase) Title() string      { return c.title }
func (c channelBase) Icon() string       { return c.icon }
func (c channelBase) Fields() []Field    { return c.fields }
func (c channelBase) DefaultEvents() db.AlertEvents {
	return append(db.AlertEvents(nil), c.defaultEvents...)
}
func (c channelBase) BodyFormat() BodyFormat { return c.format }
func (c channelBase) DefaultBody() string    { return builtinBody(c.templateFile) }

func (c channelBase) StatusColor(task_logger.TaskStatus) string { return "" }
func (c channelBase) ServerReady(*util.ConfigType) error        { return nil }

// chatEvents are the events chat channels sent historically: every
// notifiable status, including a task waiting for confirmation.
func chatEvents() db.AlertEvents {
	return db.AlertEvents{db.AlertEventSuccess, db.AlertEventError, db.AlertEventWaitingConfirmation}
}

// webhookChannel is a messenger that accepts a JSON document on an incoming
// webhook URL. Slack, Microsoft Teams, Rocket.Chat and DingTalk are all
// instances of it.
type webhookChannel struct {
	channelBase
	okCodes []int
	colors  map[task_logger.TaskStatus]string
	// enabled / url read the legacy server-wide settings from config.json.
	enabled func(cfg *util.ConfigType) bool
	url     func(cfg *util.ConfigType) string
}

func newWebhookChannel(
	typ db.AlertType,
	title string,
	icon string,
	templateFile string,
	okCodes []int,
	colors map[task_logger.TaskStatus]string,
	enabled func(cfg *util.ConfigType) bool,
	url func(cfg *util.ConfigType) string,
) *webhookChannel {
	return &webhookChannel{
		channelBase: channelBase{
			typ:           typ,
			title:         title,
			icon:          icon,
			fields:        []Field{{Name: FieldURL, Required: true}},
			defaultEvents: chatEvents(),
			format:        BodyFormatText,
			templateFile:  templateFile,
		},
		okCodes: okCodes,
		colors:  colors,
		enabled: enabled,
		url:     url,
	}
}

func (c *webhookChannel) StatusColor(status task_logger.TaskStatus) string {
	return c.colors[status]
}

func (c *webhookChannel) Validate(dest Destination) error {
	if strings.TrimSpace(dest.URL) == "" {
		return common_errors.NewValidationError(c.title + " webhook URL can not be empty")
	}
	return ValidateWebhookURL(dest.URL)
}

func (c *webhookChannel) InstanceDestination(cfg *util.ConfigType, _ db.Project) (Destination, bool) {
	if cfg == nil || !c.enabled(cfg) || strings.TrimSpace(c.url(cfg)) == "" {
		return Destination{}, false
	}
	return Destination{URL: strings.TrimSpace(c.url(cfg)), Trusted: true}, true
}

func (c *webhookChannel) Send(ctx context.Context, _ *util.ConfigType, dest Destination, msg Message) error {
	return postJSON(ctx, dest, dest.URL, msg.Body, nil, c.okCodes...)
}
