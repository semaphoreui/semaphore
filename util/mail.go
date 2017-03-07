package util

import (
	"bytes"
	"net/smtp"
)

func SendMail(emailHost string, mailSender string, mailRecipient string, subject string, body string) error {

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
	buf := bytes.NewBufferString("Subject: " + subject + "\r\n\r\n" + body + "\r\n")
	if _, err = buf.WriteTo(wc); err != nil {
		return err
	}

	return nil

}
