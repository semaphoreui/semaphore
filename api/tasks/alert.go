package tasks

import (
	"strconv"

	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
)

func (t *task) sendMailAlert() {
	for _, user := range t.users {

		mailHost := util.Config.EmailHost + ":" + util.Config.EmailPort

		userObj, err := models.FetchUser(user)
		if err != nil {
			t.log("Can't find user Email!")
			panic(err)
		}

		mailSubj := "Task '" + t.template.Alias + "' failed"
		mailBody := "Task '" + strconv.Itoa(t.task.ID) + "' with template '" + t.template.Alias + "' was failed! \nTask log: " + util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID)

		t.log("Sending email to " + userObj.Email + " from " + util.Config.EmailSender)
		err = util.SendMail(mailHost, util.Config.EmailSender, userObj.Email, mailSubj, mailBody)
		if err != nil {
			t.log("Can't send email!")
			t.log("Error: " + err.Error())
			panic(err)
		}
	}

}
