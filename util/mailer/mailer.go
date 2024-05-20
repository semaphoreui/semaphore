package mailer

import (
	"bytes"
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

// Send simply sends the defined mail via SMTP.
func Send(
	secure bool,
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
		Date:    time.Now().UTC().Format(time.RFC1123),
		To:      r.Replace(to),
		From:    r.Replace(from),
		Subject: r.Replace(subject),
		Body:    content,
	})

	if err != nil {
		return err
	}

	if secure {
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
	return smtp.SendMail(
		net.JoinHostPort(
			host,
			port,
		),
		smtp.PlainAuth(
			"",
			username,
			password,
			host,
		),
		from,
		[]string{to},
		body.Bytes(),
	)
}

func anonymous(
	host string,
	port string,
	from string,
	to string,
	body *bytes.Buffer,
) error {
	c, err := smtp.Dial(
		net.JoinHostPort(
			host,
			port,
		),
	)

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
