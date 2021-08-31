package tasks

import (
	"bytes"
	"html/template"
	"net/http"
	"strconv"

	"github.com/ansible-semaphore/semaphore/util"
)

const emailTemplate = `Subject: Task '{{ .Alias }}' failed

Task {{ .TaskID }} with template '{{ .Alias }}' has failed!
Task log: <a href='{{ .TaskURL }}'>{{ .TaskURL }}</a>`

const telegramTemplate = `{"chat_id": "{{ .ChatID }}","text":"<b>Task {{ .TaskID }} with template '{{ .Alias }}' has failed!</b>\nTask log: <a href='{{ .TaskURL }}'>{{ .TaskURL }}</a>","parse_mode":"HTML"}`

// Alert represents an alert that will be templated and sent to the appropriate service
type Alert struct {
	TaskID  string
	Alias   string
	TaskURL string
	ChatID  string
}

func (t *task) sendMailAlert() {
	if !util.Config.EmailAlert || !t.alert {
		return
	}

	mailHost := util.Config.EmailHost + ":" + util.Config.EmailPort

	var mailBuffer bytes.Buffer
	alert := Alert{
		TaskID:  strconv.Itoa(t.task.ID),
		Alias:   t.template.Alias,
		TaskURL: util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID),
	}
	tpl := template.New("mail body template")
	tpl, err := tpl.Parse(emailTemplate)
	util.LogError(err)

	t.panicOnError(tpl.Execute(&mailBuffer, alert), "Can't generate alert template!")

	for _, user := range t.users {
		userObj, err := t.store.GetUser(user)

		if !userObj.Alert {
			return
		}
		t.panicOnError(err,"Can't find user Email!")

		t.log("Sending email to " + userObj.Email + " from " + util.Config.EmailSender)
		err = util.SendMail(mailHost, util.Config.EmailSender, userObj.Email, mailBuffer)
		t.panicOnError(err, "Can't send email!")
	}
}

func (t *task) sendTelegramAlert() {
	if !util.Config.TelegramAlert || !t.alert {
		return
	}

	chatID := util.Config.TelegramChat
	if t.alertChat != "" {
		chatID = t.alertChat
	}

	var telegramBuffer bytes.Buffer
	alert := Alert{
		TaskID:  strconv.Itoa(t.task.ID),
		Alias:   t.template.Alias,
		TaskURL: util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID) + "/templates/" + strconv.Itoa(t.template.ID) + "?t=" + strconv.Itoa(t.task.ID),
		ChatID:  chatID,
	}

	tpl := template.New("telegram body template")
	tpl, err := tpl.Parse(telegramTemplate)
	util.LogError(err)

	t.panicOnError(tpl.Execute(&telegramBuffer, alert),"Can't generate alert template!")

	resp, err := http.Post("https://api.telegram.org/bot"+util.Config.TelegramToken+"/sendMessage", "application/json", &telegramBuffer)
	t.panicOnError(err, "Can't send telegram alert!")

	if resp.StatusCode != 200 {
		t.log("Can't send telegram alert! Response code not 200!")
	}
}
