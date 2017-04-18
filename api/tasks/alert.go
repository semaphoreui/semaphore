package tasks

import (
	"bytes"
	"html/template"
	"net/http"
	"strconv"

	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
)

const emailTemplate = `Subject: Task '{{ .Alias }}' failed

Task {{ .TaskId }} with template '{{ .Alias }}' has failed!
Task log: <a href='{{ .TaskUrl }}'>{{ .TaskUrl }}</a>`

const telegramTemplate = `{"chat_id": "{{ .ChatId }}","text":"<b>Task {{ .TaskId }} with template '{{ .Alias }}' has failed!</b>\nTask log: <a href='{{ .TaskUrl }}'>{{ .TaskUrl }}</a>","parse_mode":"HTML"}`

type Alert struct {
	TaskId  string
	Alias   string
	TaskUrl string
	ChatId  string
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

		userObj, err := models.FetchUser(user)

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

func (t *task) sendTelegramAlert() {

	if util.Config.TelegramAlert != true {
		return
	}

	if t.alert != true {
		return
	}

	var telegramBuffer bytes.Buffer
	alert := Alert{TaskId: strconv.Itoa(t.task.ID), Alias: t.template.Alias, TaskUrl: util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID), ChatId: util.Config.TelegramChat}
	tpl := template.New("telegram body template")
	tpl, err := tpl.Parse(telegramTemplate)
	err = tpl.Execute(&telegramBuffer, alert)

	if err != nil {
		t.log("Can't generate alert template!")
		panic(err)
	}

	resp, err := http.Post("https://api.telegram.org/bot"+util.Config.TelegramToken+"/sendMessage", "application/json", &telegramBuffer)

	if err != nil {
		t.log("Can't send telegram alert!")
		panic(err)
	}

	if resp.StatusCode != 200 {
		t.log("Can't send telegram alert! Response code not 200!")
	}

}
