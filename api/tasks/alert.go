package tasks

import (
	"bytes"
	"html/template"
	"strconv"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
)

const emailTemplate = `Subject: Task '{{ .Alias }}' failed

Task {{ .TaskId }} with template '{{ .Alias }}' has failed!
Task log: <a href='{{ .TaskUrl }}'>{{ .TaskUrl }}</a>`

type Alert struct {
	TaskId  string
	Alias   string
	TaskUrl string
}

func (t *task) sendMailAlert() {

	if util.Config.EmailAlert != true {
		return
	}

	if t.alert != true {
		return
	}

	mailHost := util.Config.EmailHost + ":" + util.Config.EmailPort

	var mailBuffer bytes.Buffer
	alert := Alert{TaskId: strconv.Itoa(t.task.ID), Alias: t.template.Alias, TaskUrl: util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID)}
	tpl := template.New("mail body template")
	tpl, err := tpl.Parse(emailTemplate)
	err = tpl.Execute(&mailBuffer, alert)

	if err != nil {
		t.log("Can't generate alert template!")
		panic(err)
	}

	for _, user := range t.users {

		userObj, err := db.FetchUser(user)

		if userObj.Alert != true {
			return
		}

		if err != nil {
			t.log("Can't find user Email!")
			panic(err)
		}

		t.log("Sending email to " + userObj.Email + " from " + util.Config.EmailSender)
		err = util.SendMail(mailHost, util.Config.EmailSender, userObj.Email, mailBuffer)
		if err != nil {
			t.log("Can't send email!")
			t.log("Error: " + err.Error())
			panic(err)
		}

	}
}
