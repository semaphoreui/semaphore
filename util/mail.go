package util

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"io"
	"net/smtp"
)

// SendMail dispatches a mail using smtp
func SendMail(emailHost, mailSender, mailRecipient string, mail bytes.Buffer) error {
	c, err := smtp.Dial(emailHost)
	if err != nil {
		return err
	}

	defer func(c *smtp.Client) {
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

	defer func(wc io.WriteCloser) {
		err = wc.Close()
		if err != nil {
			log.Error(err)
		}
	}(wc)
	_, err = mail.WriteTo(wc)
	return err
}

// SendSecureMail dispatches a mail using smtp with authentication and StartTLS
func SendSecureMail(emailHost, emailPort, mailSender, mailUsername, mailPassword, mailRecipient string, mail bytes.Buffer) error {

	// Receiver email address.
	to := []string{
		mailRecipient,
	}

	// Authentication.
	auth := smtp.PlainAuth("", mailUsername, mailPassword, emailHost)

	// Sending email.
	err := smtp.SendMail(emailHost+":"+emailPort, auth, mailSender, to, mail.Bytes())
	if err != nil {
		log.Error(err)
	}
	return err
}
