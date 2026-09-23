package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
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

	defaultTimeout = 15 * time.Second
)

var r = strings.NewReplacer(
	"\r\n", "",
	"\r", "",
	"\n", "",
	"%0a", "",
	"%0d", "",
)

// DialFunc opens the TCP connection to the SMTP server. It exists so a
// caller sending to a user supplied host can enforce its own outbound
// policy (which addresses may be dialed) instead of trusting the host name.
type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

// Options describes one message and the server it goes through.
type Options struct {
	Host string
	Port string
	// Secure enables authentication with Username/Password. Without TLS the
	// connection is upgraded with STARTTLS when the server offers it.
	Secure bool
	// TLS connects with implicit TLS (port 465 style) instead of STARTTLS.
	TLS      bool
	Username string
	Password string
	// TLSMinVersion is "1.0" .. "1.3"; empty means 1.2.
	TLSMinVersion string

	From    string
	To      string
	Subject string
	Body    string

	// Dial opens the connection; nil uses net.Dialer.
	Dial DialFunc
}

func parseTlsVersion(version string) (uint16, error) {
	switch version {
	case "", "1.2":
		return tls.VersionTLS12, nil
	case "1.0":
		return tls.VersionTLS10, nil
	case "1.1":
		return tls.VersionTLS11, nil
	case "1.3":
		return tls.VersionTLS13, nil
	}

	return 0, fmt.Errorf("unsupported TLS version %s", version)
}

// Send sends the mail through the server-wide SMTP settings semantics
// (net.Dialer, TLS minimum version from the loaded config). Callers that
// deliver to a user supplied host must use SendMail with a guarded Dial.
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
	var minVersion string
	if util.Config != nil {
		minVersion = util.Config.EmailTlsMinVersion
	}
	return SendMail(context.Background(), Options{
		Host:          host,
		Port:          port,
		Secure:        secure,
		TLS:           useTls,
		Username:      username,
		Password:      password,
		TLSMinVersion: minVersion,
		From:          from,
		To:            to,
		Subject:       subject,
		Body:          content,
	})
}

// SendMail sends one message. The whole exchange is bounded by the context
// deadline, or by defaultTimeout when the context has none.
func SendMail(ctx context.Context, o Options) error {
	if strings.TrimSpace(o.Host) == "" {
		return errors.New("smtp host is empty")
	}
	if o.Port == "" {
		o.Port = "25"
	}

	from := r.Replace(o.From)
	to := r.Replace(o.To)

	body, err := render(from, to, r.Replace(o.Subject), o.Body)
	if err != nil {
		return err
	}

	tlsVersion, err := parseTlsVersion(o.TLSMinVersion)
	if err != nil {
		return err
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         o.Host,
		MinVersion:         tlsVersion,
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	dial := o.Dial
	if dial == nil {
		dial = (&net.Dialer{Timeout: 10 * time.Second}).DialContext
	}

	conn, err := dial(ctx, "tcp", net.JoinHostPort(o.Host, o.Port))
	if err != nil {
		return err
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	if o.Secure && o.TLS {
		// Implicit TLS: the handshake happens before any SMTP command.
		tlsConn := tls.Client(conn, tlsConfig)
		if err = tlsConn.HandshakeContext(ctx); err != nil {
			_ = conn.Close()
			return err
		}
		conn = tlsConn
	}

	c, err := smtp.NewClient(conn, o.Host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer c.Close() //nolint:errcheck

	if o.Secure {
		if !o.TLS {
			if ok, _ := c.Extension("STARTTLS"); ok {
				if err = c.StartTLS(tlsConfig); err != nil {
					return err
				}
			}
		}
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(PlainOrLoginAuth(o.Username, o.Password, o.Host)); err != nil {
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = body.WriteTo(w); err != nil {
		_ = w.Close()
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}

	return c.Quit()
}

func render(from, to, subject, content string) (*bytes.Buffer, error) {
	body := bytes.NewBufferString("")
	tpl, err := template.New("").Parse(mailerBase)
	if err != nil {
		return nil, err
	}

	err = tpl.Execute(body, struct {
		Date    string
		To      string
		From    string
		Subject string
		Body    string
	}{
		Date:    tz.Now().Format(time.RFC1123),
		To:      to,
		From:    from,
		Subject: subject,
		Body:    content,
	})
	if err != nil {
		return nil, err
	}
	return body, nil
}
