package tasks

import (
	"bytes"
	"net/smtp"
	"strconv"

	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
)

func (t *task) sendMailAlert() {
	for _, user := range t.users {

		c, err := smtp.Dial(util.Config.EmailHost + ":" + util.Config.EmailPort)
		if err != nil {
			t.log("Can't connect to SMTP server!")
			panic(err)
		}

		userObj, err := models.FetchUser(user)
		if err != nil {
			t.log("Can't find user Email!")
			panic(err)
		}

		defer c.Close()
		// Set the sender and recipient.
		c.Mail(util.Config.EmailSender)
		c.Rcpt(userObj.Email)
		// Send the email body.
		wc, err := c.Data()
		if err != nil {
			t.log("Can't create Email!")
			panic(err)
		}
		defer wc.Close()
		mailSubj := "Task '" + t.template.Alias + "' failed"
		mailBody := "Task '" + strconv.Itoa(t.task.ID) + "' with template '" + t.template.Alias + "' was failed! \nTask log: " + util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID)
		t.log(mailBody)
		buf := bytes.NewBufferString("Subject: " + mailSubj + "\r\n\r\n" + mailBody + "\r\n")
		if _, err = buf.WriteTo(wc); err != nil {
			t.log("Can't send Email!")
			panic(err)
		}
		t.log("Email to " + userObj.Email + " successfully sent from " + util.Config.EmailSender)
	}

}
