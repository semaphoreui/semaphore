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

// emailChannel sends through SMTP. By default it uses the server from
// config.json; a project alert may bring its own SMTP server (override
// fields) and/or its own credentials (a login_password access key). Every
// setting the alert leaves empty falls back to the server configuration.
// Recipients are either listed on the alert or, when the list is empty, the
// project members who opted in to alerts (the legacy behaviour).
type emailChannel struct {
	channelBase
}

func newEmailChannel() Channel {
	return &emailChannel{
		channelBase: channelBase{
			typ:   TypeEmail,
			title: "Email",
			icon:  "mdi-email-outline",
			fields: []Field{
				{Name: FieldRecipients, Kind: FieldKindText},
				{Name: FieldSMTPHost, Kind: FieldKindText, Override: true},
				{Name: FieldSMTPPort, Kind: FieldKindNumber, Override: true},
				{Name: FieldSMTPSender, Kind: FieldKindText, Override: true},
				{Name: FieldSMTPSecure, Kind: FieldKindBool, Override: true},
				{Name: FieldSMTPTLS, Kind: FieldKindBool, Override: true},
			},
			// E-mail historically reported failures only.
			defaultEvents: db.AlertEvents{db.AlertEventError},
			format:        BodyFormatHTML,
			templateFile:  "email.tmpl",
		},
	}
}

func (c *emailChannel) Secret() *SecretSpec {
	return &SecretSpec{KeyType: db.AccessKeyLoginPassword, Label: "SMTP credentials", Optional: true}
}

// smtpSettings is the effective SMTP configuration for one destination.
type smtpSettings struct {
	host, port, sender, username, password string
	secure, tls                            bool
}

func (c *emailChannel) Validate(cfg *util.ConfigType, dest Destination) error {
	for _, addr := range dest.Recipients {
		if !strings.Contains(addr, "@") || strings.ContainsAny(addr, " \t\r\n<>") {
			return common_errors.NewValidationError("invalid e-mail address: " + addr)
		}
	}
	_, err := c.settings(cfg, dest)
	return err
}

// settings merges the alert's own SMTP settings over the server ones.
func (c *emailChannel) settings(cfg *util.ConfigType, dest Destination) (smtpSettings, error) {
	var s smtpSettings
	if cfg != nil {
		s = smtpSettings{
			host:     strings.TrimSpace(cfg.EmailHost),
			port:     strings.TrimSpace(cfg.EmailPort),
			sender:   strings.TrimSpace(cfg.EmailSender),
			username: cfg.EmailUsername,
			password: cfg.EmailPassword,
			secure:   cfg.EmailSecure,
			tls:      cfg.EmailTls,
		}
	}

	if host := dest.Param(FieldSMTPHost); host != "" {
		// An own server never inherits the server-wide credentials.
		s = smtpSettings{
			host:   host,
			port:   dest.Param(FieldSMTPPort),
			sender: dest.Param(FieldSMTPSender),
			secure: dest.ParamBool(FieldSMTPSecure),
			tls:    dest.ParamBool(FieldSMTPTLS),
		}
		if s.port == "" {
			s.port = "25"
		}
		if s.sender == "" {
			return s, common_errors.NewValidationError("smtp_sender is required when an own SMTP server is set")
		}
	} else if dest.Param(FieldSMTPPort) != "" || dest.Param(FieldSMTPSender) != "" {
		return s, common_errors.NewValidationError("smtp_host is required when overriding the SMTP server")
	}

	if dest.Secret != nil {
		s.username = dest.Secret.LoginPassword.Login
		s.password = dest.Secret.LoginPassword.Password
	}

	if s.host == "" {
		return s, common_errors.NewValidationError("SMTP is not configured on the server (email_host); set your own SMTP server on the alert")
	}
	return s, nil
}

func (c *emailChannel) InstanceDestination(cfg *util.ConfigType, _ db.Project) (Destination, bool) {
	if cfg == nil || !cfg.EmailAlert || c.ServerReady(cfg) != nil {
		return Destination{}, false
	}
	return Destination{Trusted: true}, true
}

func (c *emailChannel) ServerReady(cfg *util.ConfigType) error {
	if cfg == nil || strings.TrimSpace(cfg.EmailHost) == "" {
		return common_errors.NewValidationError("SMTP is not configured on the server (email_host)")
	}
	return nil
}

func (c *emailChannel) Send(ctx context.Context, cfg *util.ConfigType, dest Destination, msg Message) error {
	s, err := c.settings(cfg, dest)
	if err != nil {
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
			s.secure,
			s.tls,
			s.host,
			s.port,
			s.username,
			s.password,
			s.sender,
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
