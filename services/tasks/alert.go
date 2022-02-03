package tasks

import (
	"bytes"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/ansible-semaphore/semaphore/util"
)

const emailTemplate = `Subject: Task '{{ .Name }}' failed

Task {{ .TaskID }} with template '{{ .Name }}' has failed!
Task log: <a href='{{ .TaskURL }}'>{{ .TaskURL }}</a>`

const telegramTemplate = `{"chat_id": "{{ .ChatID }}","parse_mode":"HTML","text":"<code>{{ .Name }}</code>\n#{{ .TaskID }} <b>{{ .TaskResult }}</b> <code>{{ .TaskVersion }}</code> {{ .TaskDescription }}\nby {{ .Author }}\n{{ .TaskURL }}"}`

// Alert represents an alert that will be templated and sent to the appropriate service
type Alert struct {
	TaskID          string
	Name            string
	TaskURL         string
	ChatID          string
	TaskResult      string
	TaskDescription string
	TaskVersion     string
	Author          string
}

func (t *TaskRunner) sendMailAlert() {
	if !util.Config.EmailAlert || !t.alert {
		return
	}

	mailHost := util.Config.EmailHost + ":" + util.Config.EmailPort

	var mailBuffer bytes.Buffer
	alert := Alert{
		TaskID:  strconv.Itoa(t.task.ID),
		Name:    t.template.Name,
		TaskURL: util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID),
	}
	tpl := template.New("mail body template")
	tpl, err := tpl.Parse(emailTemplate)
	util.LogError(err)

	t.panicOnError(tpl.Execute(&mailBuffer, alert), "Can't generate alert template!")

	for _, user := range t.users {
		userObj, err := t.pool.store.GetUser(user)

		if !userObj.Alert {
			return
		}
		t.panicOnError(err, "Can't find user Email!")

		t.Log("Sending email to " + userObj.Email + " from " + util.Config.EmailSender)
		if util.Config.EmailSecure {
			err = util.SendSecureMail(util.Config.EmailHost, util.Config.EmailPort, util.Config.EmailSender, util.Config.EmailUsername, util.Config.EmailPassword, userObj.Email, mailBuffer)
		} else {
			err = util.SendMail(mailHost, util.Config.EmailSender, userObj.Email, mailBuffer)
		}
		t.panicOnError(err, "Can't send email!")
	}
}

func (t *TaskRunner) sendTelegramAlert() {
	if !util.Config.TelegramAlert || !t.alert {
		return
	}

	chatID := util.Config.TelegramChat
	if t.alertChat != nil && *t.alertChat != "" {
		chatID = *t.alertChat
	}

	var telegramBuffer bytes.Buffer

	var version string
	if t.task.Version != nil {
		version = *t.task.Version
	} else if t.task.BuildTaskID != nil {
		version = "build " + strconv.Itoa(*t.task.BuildTaskID)
	} else {
		version = ""
	}

	var message string
	if t.task.Message != "" {
		message = "- " + t.task.Message
	}

	var author string
	if t.task.UserID != nil {
		user, err := t.pool.store.GetUser(*t.task.UserID)
		if err != nil {
			panic(err)
		}
		author = user.Name
	}

	alert := Alert{
		TaskID:          strconv.Itoa(t.task.ID),
		Name:            t.template.Name,
		TaskURL:         util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID) + "/templates/" + strconv.Itoa(t.template.ID) + "?t=" + strconv.Itoa(t.task.ID),
		ChatID:          chatID,
		TaskResult:      strings.ToUpper(t.task.Status),
		TaskVersion:     version,
		TaskDescription: message,
		Author:          author,
	}

	tpl := template.New("telegram body template")

	tpl, err := tpl.Parse(telegramTemplate)
	if err != nil {
		t.Log("Can't parse telegram template!")
		panic(err)
	}

	err = tpl.Execute(&telegramBuffer, alert)
	if err != nil {
		t.Log("Can't generate alert template!")
		panic(err)
	}

	resp, err := http.Post("https://api.telegram.org/bot"+util.Config.TelegramToken+"/sendMessage", "application/json", &telegramBuffer)

	if err != nil {
		t.Log("Can't send telegram alert! Response code not 200!")
	} else if resp.StatusCode != 200 {
		t.Log("Can't send telegram alert! Response code not 200!")
	}
}
