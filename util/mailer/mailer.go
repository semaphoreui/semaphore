package mailer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

const (
	mailerBase = "MIME-version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		//"Content-Transfer-Encoding: quoted-printable\r\n" +
		"Date: {{ .Date }}\r\n" +
		"To: {{ .To }}\r\n" +
		"From: {{ .From }}\r\n" +
		"Subject: {{ .Subject }}\r\n\r\n" +
		"{{ .Body }}"
)

var r = strings.NewReplacer(
	"\r\n", "",
	"\r", "",
	"\n", "",
	"%0a", "",
	"%0d", "",
)

func parseTlsVersion(version string) (uint16, error) {
	switch version {
	case "1.0":
		return tls.VersionTLS10, nil
	case "1.1":
		return tls.VersionTLS11, nil
	case "1.2":
		return tls.VersionTLS12, nil
	case "1.3":
		return tls.VersionTLS13, nil
	}

	return 0, fmt.Errorf("Unsupported TLS version %s", version)
}

// Send simply sends the defined mail via SMTP.
// 
// Email sending logic:
// - If username and password are provided, authentication will be used
// - If useTls=true: Direct TLS connection (typically port 465)
// - If secure=true and useTls=false: STARTTLS connection (typically port 587)
// - If secure=false and useTls=false: Plain authentication (not recommended)
// - If no username/password: Anonymous SMTP (most servers reject this)
//
// Common configurations:
// - Gmail/Office365: secure=true, useTls=false, port=587 (STARTTLS)
// - Some providers: secure=false, useTls=true, port=465 (Direct TLS)
// - Local dev servers: secure=false, useTls=false, port=25/1025 (Plain/Anonymous)
func Send(
	secure bool,
	useTls bool,
	host string,
	port string,
	username,
	password,
	from,
	to,
	subject string,
	content string,
) error {
	log.WithFields(log.Fields{
		"host":     host,
		"port":     port,
		"secure":   secure,
		"useTls":   useTls,
		"from":     from,
		"to":       to,
		"subject":  subject,
		"username": username,
		"context":  "mailer_send",
	}).Debug("attempting to send email")

	// Validate required parameters
	if host == "" || port == "" {
		err := fmt.Errorf("email host and port are required")
		log.WithError(err).Error("email configuration validation failed")
		return err
	}

	body := bytes.NewBufferString("")
	tpl, err := template.New("").Parse(mailerBase)
	if err != nil {
		log.WithError(err).Error("failed to parse email template")
		return err
	}

	err = tpl.Execute(body, struct {
		Date    string
		To      string
		From    string
		Subject string
		Body    string
	}{
		Date:    tz.Now().Format(time.RFC1123),
		To:      r.Replace(to),
		From:    r.Replace(from),
		Subject: r.Replace(subject),
		Body:    content,
	})
	if err != nil {
		log.WithError(err).Error("failed to execute email template")
		return err
	}

	// Determine email sending method based on configuration
	if username != "" && password != "" {
		// Authentication is required
		if useTls {
			// Direct TLS connection (usually port 465)
			log.Debug("using direct TLS connection with authentication")
			return sendTls(host, port, username, password, from, to, body)
		} else if secure {
			// STARTTLS connection (usually port 587) 
			log.Debug("using STARTTLS connection with authentication")
			return sendStartTls(host, port, username, password, from, to, body)
		} else {
			// Plain authentication without encryption (not recommended)
			log.Debug("using plain authentication without encryption")
			return plainauth(host, port, username, password, from, to, body)
		}
	} else {
		// No authentication (anonymous)
		if secure || useTls {
			log.Warn("secure/TLS requested but no credentials provided, falling back to anonymous")
		}
		log.Debug("using anonymous SMTP connection")
		return anonymous(host, port, from, to, body)
	}
}

func plainauth(
	host string,
	port string,
	username string,
	password string,
	from string,
	to string,
	body *bytes.Buffer,
) error {
	log.WithFields(log.Fields{
		"host":     host,
		"port":     port,
		"username": username,
		"context":  "plainauth",
	}).Debug("attempting plain auth email")

	auth := PlainOrLoginAuth(username, password, host)
	// auth := smtp.PlainAuth("", username, password, host)

	err := smtp.SendMail(
		net.JoinHostPort(host, port),
		auth,
		from,
		[]string{to},
		body.Bytes(),
	)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"host":    host,
			"port":    port,
			"context": "plainauth",
		}).Error("failed to send email via plain auth")
	} else {
		log.WithFields(log.Fields{
			"host":    host,
			"port":    port,
			"context": "plainauth",
		}).Debug("email sent successfully via plain auth")
	}

	return err
}

func sendStartTls(
	host,
	port,
	username,
	password,
	from,
	to string,
	body *bytes.Buffer,
) error {
	log.WithFields(log.Fields{
		"host":     host,
		"port":     port,
		"username": username,
		"context":  "sendStartTls",
	}).Debug("attempting STARTTLS email")

	// Connect to the server
	c, err := smtp.Dial(net.JoinHostPort(host, port))
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"host":    host,
			"port":    port,
			"context": "sendStartTls",
		}).Error("failed to connect to SMTP server")
		return err
	}
	defer c.Close() //nolint:errcheck

	// Start TLS if supported
	if ok, _ := c.Extension("STARTTLS"); ok {
		tlsVersion, err := parseTlsVersion(util.Config.EmailTlsMinVersion)
		if err != nil {
			log.WithError(err).Error("failed to parse TLS version")
			return err
		}

		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         host,
			MinVersion:         tlsVersion,
		}

		if err = c.StartTLS(tlsConfig); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"host":    host,
				"port":    port,
				"context": "sendStartTls",
			}).Error("failed to start TLS")
			return err
		}
	} else {
		log.WithFields(log.Fields{
			"host":    host,
			"port":    port,
			"context": "sendStartTls",
		}).Warn("STARTTLS not supported by server, continuing with plain connection")
	}

	// Authenticate
	auth := PlainOrLoginAuth(username, password, host)
	if err = c.Auth(auth); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"host":     host,
			"port":     port,
			"username": username,
			"context":  "sendStartTls",
		}).Error("failed to authenticate with SMTP server")
		return err
	}

	// Send email
	if err = c.Mail(from); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"from":    from,
			"context": "sendStartTls",
		}).Error("failed to set sender")
		return err
	}

	if err = c.Rcpt(to); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"to":      to,
			"context": "sendStartTls",
		}).Error("failed to set recipient")
		return err
	}

	w, err := c.Data()
	if err != nil {
		log.WithError(err).WithField("context", "sendStartTls").Error("failed to get data writer")
		return err
	}

	_, err = w.Write(body.Bytes())
	if err != nil {
		log.WithError(err).WithField("context", "sendStartTls").Error("failed to write email data")
		return err
	}

	err = w.Close()
	if err != nil {
		log.WithError(err).WithField("context", "sendStartTls").Error("failed to close data writer")
		return err
	}

	err = c.Quit()
	if err != nil {
		log.WithError(err).WithField("context", "sendStartTls").Error("failed to quit SMTP connection")
		return err
	}

	log.WithFields(log.Fields{
		"host":    host,
		"port":    port,
		"context": "sendStartTls",
	}).Debug("email sent successfully via STARTTLS")

	return nil
}

func sendTls(
	host,
	port,
	username,
	password,
	from,
	to string,
	body *bytes.Buffer,
) error {
	log.WithFields(log.Fields{
		"host":     host,
		"port":     port,
		"username": username,
		"context":  "sendTls",
	}).Debug("attempting direct TLS email")

	auth := PlainOrLoginAuth(username, password, host)

	tlsVersion, err := parseTlsVersion(util.Config.EmailTlsMinVersion)
	if err != nil {
		log.WithError(err).Error("failed to parse TLS version")
		return err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
		MinVersion:         tlsVersion,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", net.JoinHostPort(host, port), tlsConfig)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"host":    host,
			"port":    port,
			"context": "sendTls",
		}).Error("failed to establish TLS connection")
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"host":    host,
			"context": "sendTls",
		}).Error("failed to create SMTP client")
		return err
	}

	if err = c.Auth(auth); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"host":     host,
			"port":     port,
			"username": username,
			"context":  "sendTls",
		}).Error("failed to authenticate with SMTP server")
		return err
	}

	if err = c.Mail(from); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"from":    from,
			"context": "sendTls",
		}).Error("failed to set sender")
		return err
	}

	if err = c.Rcpt(to); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"to":      to,
			"context": "sendTls",
		}).Error("failed to set recipient")
		return err
	}

	w, err := c.Data()
	if err != nil {
		log.WithError(err).WithField("context", "sendTls").Error("failed to get data writer")
		return err
	}

	_, err = w.Write(body.Bytes())
	if err != nil {
		log.WithError(err).WithField("context", "sendTls").Error("failed to write email data")
		return err
	}

	err = w.Close()
	if err != nil {
		log.WithError(err).WithField("context", "sendTls").Error("failed to close data writer")
		return err
	}

	err = c.Quit()
	if err != nil {
		log.WithError(err).WithField("context", "sendTls").Error("failed to quit SMTP connection")
		return err
	}

	log.WithFields(log.Fields{
		"host":    host,
		"port":    port,
		"context": "sendTls",
	}).Debug("email sent successfully via direct TLS")

	return nil
}

func anonymous(
	host string,
	port string,
	from string,
	to string,
	body *bytes.Buffer,
) error {
	log.WithFields(log.Fields{
		"host":    host,
		"port":    port,
		"context": "anonymous",
	}).Debug("attempting anonymous email")

	c, err := smtp.Dial(net.JoinHostPort(host, port))
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"host":    host,
			"port":    port,
			"context": "anonymous",
		}).Error("failed to connect to SMTP server")
		return err
	}

	defer c.Close() //nolint:errcheck

	if err := c.Mail(r.Replace(from)); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"from":    from,
			"context": "anonymous",
		}).Error("failed to set sender")
		return err
	}

	if err = c.Rcpt(r.Replace(to)); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"to":      to,
			"context": "anonymous",
		}).Error("failed to set recipient")
		return err
	}

	w, err := c.Data()
	if err != nil {
		log.WithError(err).WithField("context", "anonymous").Error("failed to get data writer")
		return err
	}

	defer w.Close() //nolint:errcheck

	if _, err := body.WriteTo(w); err != nil {
		log.WithError(err).WithField("context", "anonymous").Error("failed to write email data")
		return err
	}

	log.WithFields(log.Fields{
		"host":    host,
		"port":    port,
		"context": "anonymous",
	}).Debug("email sent successfully via anonymous SMTP")

	return nil
}
