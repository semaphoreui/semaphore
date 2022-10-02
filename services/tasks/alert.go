package tasks

import (
	"bytes"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

const emailTemplate = "Subject: Task '{{ .Name }}' failed\r\n" +
	"From: {{ .From }}\r\n" +
	"\r\n" +
	"Task {{ .TaskID }} with template '{{ .Name }}' has failed!`\n" +
	"Task Log: {{ .TaskURL }}"

const telegramTemplate = `{"chat_id": "{{ .ChatID }}","parse_mode":"HTML","text":"<code>{{ .Name }}</code>\n#{{ .TaskID }} <b>{{ .TaskResult }}</b> <code>{{ .TaskVersion }}</code> {{ .TaskDescription }}\nby {{ .Author }}\n{{ .TaskURL }}"}`

const slackTemplate = `{ "attachments": [ { "title": "Task: {{ .Name }}", "title_link": "{{ .TaskURL }}", "text": "execution ID #{{ .TaskID }}, status: {{ .TaskResult }}!", "color": "{{ .Color }}", "mrkdwn_in": ["text"], "fields": [ { "title": "Author", "value": "{{ .Author }}", "short": true }] } ]}`

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
	Color           string
	From            string
}

func (t *TaskRunner) sendMailAlert() {
	if !util.Config.EmailAlert || !t.alert {
		return
	}

	mailHost := util.Config.EmailHost + ":" + util.Config.EmailPort

	var mailBuffer bytes.Buffer
	alert := Alert{
		TaskID: strconv.Itoa(t.task.ID),
		Name:   t.template.Name,
		TaskURL: util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID) +
			"/templates/" + strconv.Itoa(t.template.ID) +
			"?t=" + strconv.Itoa(t.task.ID),
		From: util.Config.EmailSender,
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
			err = util.SendSecureMail(util.Config.EmailHost, util.Config.EmailPort,
				util.Config.EmailSender, util.Config.EmailUsername, util.Config.EmailPassword,
				userObj.Email, mailBuffer)
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

	if t.template.SuppressSuccessAlerts && t.task.Status == db.TaskSuccessStatus {
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
		TaskResult:      strings.ToUpper(string(t.task.Status)),
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
		t.Log("Can't send telegram alert! Error: " + err.Error())
	} else if resp.StatusCode != 200 {
		t.Log("Can't send telegram alert! Response code: " + strconv.Itoa(resp.StatusCode))
	}
}

func (t *TaskRunner) sendSlackAlert() {
	if !util.Config.SlackAlert || !t.alert {
		return
	}

	if t.template.SuppressSuccessAlerts && t.task.Status == db.TaskSuccessStatus {
		return
	}

	slackUrl := util.Config.SlackUrl

	var slackBuffer bytes.Buffer

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

	var color string
	if t.task.Status == db.TaskSuccessStatus {
		color = "good"
	} else if t.task.Status == db.TaskFailStatus {
		color = "bad"
	} else if t.task.Status == db.TaskRunningStatus {
		color = "#333CFF"
	} else if t.task.Status == db.TaskWaitingStatus {
		color = "#FFFC33"
	} else if t.task.Status == db.TaskStoppingStatus {
		color = "#BEBEBE"
	} else if t.task.Status == db.TaskStoppedStatus {
		color = "#5B5B5B"
	}
	alert := Alert{
		TaskID:          strconv.Itoa(t.task.ID),
		Name:            t.template.Name,
		TaskURL:         util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID) + "/templates/" + strconv.Itoa(t.template.ID) + "?t=" + strconv.Itoa(t.task.ID),
		TaskResult:      strings.ToUpper(string(t.task.Status)),
		TaskVersion:     version,
		TaskDescription: message,
		Author:          author,
		Color:           color,
	}

	tpl := template.New("slack body template")

	tpl, err := tpl.Parse(slackTemplate)
	if err != nil {
		t.Log("Can't parse slack template!")
		panic(err)
	}

	err = tpl.Execute(&slackBuffer, alert)
	if err != nil {
		t.Log("Can't generate alert template!")
		panic(err)
	}
	resp, err := http.Post(slackUrl, "application/json", &slackBuffer)

	if err != nil {
		t.Log("Can't send slack alert! Error: " + err.Error())
	} else if resp.StatusCode != 200 {
		t.Log("Can't send slack alert! Response code: " + strconv.Itoa(resp.StatusCode))
	}
}
