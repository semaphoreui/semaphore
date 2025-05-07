package mailer

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
	"net"
	"net/smtp"
	"strings"
	"text/template"
	"time"
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

var (
	r = strings.NewReplacer(
		"\r\n", "",
		"\r", "",
		"\n", "",
		"%0a", "",
		"%0d", "",
	)
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

	return 0, errors.New(fmt.Sprintf("Unsupported TLS version %s", version))
}

// Send simply sends the defined mail via SMTP.
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
	body := bytes.NewBufferString("")
	tpl, err := template.New("").Parse(mailerBase)

	if err != nil {
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
		return err
	}

	if secure {
		if useTls {
			return sendTls(
				host,
				port,
				username,
				password,
				from,
				to,
				body,
			)
		} else {
			return plainauth(
				host,
				port,
				username,
				password,
				from,
				to,
				body,
			)
		}
	}

	return anonymous(
		host,
		port,
		from,
		to,
		body,
	)
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
	auth := PlainOrLoginAuth(username, password, host)
	//auth := smtp.PlainAuth("", username, password, host)

	return smtp.SendMail(
		net.JoinHostPort(host, port),
		auth,
		from,
		[]string{to},
		body.Bytes(),
	)
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
	auth := PlainOrLoginAuth(username, password, host)

	tlsVersion, err := parseTlsVersion(util.Config.EmailTlsMinVersion)
	if err != nil {
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
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if err = c.Auth(auth); err != nil {
		return err
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

	_, err = w.Write(body.Bytes())
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	err = c.Quit()
	if err != nil {
		return err
	}

	return nil
}

func anonymous(
	host string,
	port string,
	from string,
	to string,
	body *bytes.Buffer,
) error {
	c, err := smtp.Dial(net.JoinHostPort(host, port))

	if err != nil {
		return err
	}

	defer c.Close()

	if err := c.Mail(r.Replace(from)); err != nil {
		return err
	}

	if err = c.Rcpt(r.Replace(to)); err != nil {
		return err
	}

	w, err := c.Data()

	if err != nil {
		return err
	}

	defer w.Close()

	if _, err := body.WriteTo(w); err != nil {
		return err
	}

	return nil
}
