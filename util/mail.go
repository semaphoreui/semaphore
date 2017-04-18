package util

import (
	"bytes"
	"net/smtp"
)

func SendMail(emailHost, mailSender, mailRecipient string, mail bytes.Buffer) error {

	c, err := smtp.Dial(emailHost)
	if err != nil {
		return err
	}

	defer c.Close()
	// Set the sender and recipient.
	c.Mail(mailSender)
	c.Rcpt(mailRecipient)

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}

	defer wc.Close()
	if _, err = mail.WriteTo(wc); err != nil {
		return err
	}
	return nil

}
