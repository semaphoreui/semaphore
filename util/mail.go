package util

import (
	"bytes"
	"net/smtp"
	log "github.com/Sirupsen/logrus"
	"io"
)

// SendMail dispatches a mail using smtp
func SendMail(emailHost, mailSender, mailRecipient string, mail bytes.Buffer) error {
	c, err := smtp.Dial(emailHost)
	if err != nil {
		return err
	}

	defer func (c *smtp.Client) {
		err = c.Close()
		if err != nil {
			log.Error(err)
		}
	}(c)

	// Set the sender and recipient.
	err = c.Mail(mailSender)
	if err != nil {
		return err
	}
	err = c.Rcpt(mailRecipient)
	if err != nil {
		return err
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}

	defer func (wc io.WriteCloser) {
		err = wc.Close()
		if err != nil {
			log.Error(err)
		}
	}(wc)
	_, err = mail.WriteTo(wc)
	return err
}
