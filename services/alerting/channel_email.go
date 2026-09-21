package alerting

import (
	"context"
	"errors"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/util"
	"github.com/semaphoreui/semaphore/util/mailer"
)

const TypeEmail db.AlertType = "email"

// emailChannel sends through the SMTP server from config.json. Recipients
// are either listed on the alert or, when the list is empty, the project
// members who opted in to alerts (the legacy behaviour).
type emailChannel struct {
	channelBase
}

func newEmailChannel() Channel {
	return &emailChannel{
		channelBase: channelBase{
			typ:    TypeEmail,
			title:  "Email",
			icon:   "mdi-email-outline",
			fields: []Field{{Name: FieldRecipients}},
			// E-mail historically reported failures only.
			defaultEvents: db.AlertEvents{db.AlertEventError},
			format:        BodyFormatHTML,
			templateFile:  "email.tmpl",
		},
	}
}

func (c *emailChannel) Validate(dest Destination) error {
	for _, addr := range dest.Recipients {
		if !strings.Contains(addr, "@") || strings.ContainsAny(addr, " \t\r\n<>") {
			return common_errors.NewValidationError("invalid e-mail address: " + addr)
		}
	}
	return nil
}

func (c *emailChannel) InstanceDestination(cfg *util.ConfigType, _ db.Project) (Destination, bool) {
	if c.ServerReady(cfg) != nil {
		return Destination{}, false
	}
	return Destination{Trusted: true}, true
}

func (c *emailChannel) ServerReady(cfg *util.ConfigType) error {
	if cfg == nil || !cfg.EmailAlert || strings.TrimSpace(cfg.EmailHost) == "" {
		return common_errors.NewValidationError("SMTP is not configured on the server (email_alert, email_host)")
	}
	return nil
}

func (c *emailChannel) Send(ctx context.Context, cfg *util.ConfigType, dest Destination, msg Message) error {
	if err := c.ServerReady(cfg); err != nil {
		return err
	}
	if len(dest.Recipients) == 0 {
		return common_errors.NewValidationError("no e-mail recipients")
	}

	var errs []error
	for _, to := range dest.Recipients {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := mailer.Send(
			cfg.EmailSecure,
			cfg.EmailTls,
			cfg.EmailHost,
			cfg.EmailPort,
			cfg.EmailUsername,
			cfg.EmailPassword,
			cfg.EmailSender,
			to,
			msg.Subject,
			msg.Body,
		)
		if err != nil {
			errs = append(errs, common_errors.NewUserError(err))
		}
	}
	return errors.Join(errs...)
}
