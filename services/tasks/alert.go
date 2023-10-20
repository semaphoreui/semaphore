package tasks

import (
	"bytes"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ansible-semaphore/semaphore/lib"
	"github.com/ansible-semaphore/semaphore/util"
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
		TaskID: strconv.Itoa(t.Task.ID),
		Name:   t.Template.Name,
		TaskURL: util.Config.WebHost + "/project/" + strconv.Itoa(t.Template.ProjectID) +
			"/templates/" + strconv.Itoa(t.Template.ID) +
			"?t=" + strconv.Itoa(t.Task.ID),
		From: util.Config.EmailSender,
	}
	tpl := template.New("mail body template")
	tpl, err := tpl.Parse(emailTemplate)
	util.LogError(err)

	t.panicOnError(tpl.Execute(&mailBuffer, alert), "Can't generate alert template!")

	for _, user := range t.users {
		userObj, err2 := t.pool.store.GetUser(user)

		if !userObj.Alert {
			continue
		}

		if err2 != nil {
			util.LogError(err2)
			continue
		}

		if util.Config.EmailSecure {
			err2 = util.SendSecureMail(util.Config.EmailHost, util.Config.EmailPort,
				util.Config.EmailSender, util.Config.EmailUsername, util.Config.EmailPassword,
				userObj.Email, mailBuffer)
		} else {
			err2 = util.SendMail(mailHost, util.Config.EmailSender, userObj.Email, mailBuffer)
		}

		if err2 != nil {
			util.LogError(err2)
		}
	}
}

func (t *TaskRunner) sendTelegramAlert() {
	if !util.Config.TelegramAlert || !t.alert {
		return
	}

	if t.Template.SuppressSuccessAlerts && t.Task.Status == lib.TaskSuccessStatus {
		return
	}

	chatID := util.Config.TelegramChat
	if t.alertChat != nil && *t.alertChat != "" {
		chatID = *t.alertChat
	}

	if chatID == "" {
		return
	}

	var telegramBuffer bytes.Buffer

	var version string
	if t.Task.Version != nil {
		version = *t.Task.Version
	} else if t.Task.BuildTaskID != nil {
		buildVer := t.Task.GetIncomingVersion(t.pool.store)
		if buildVer != nil {
			version = *buildVer
		}
	} else {
		version = ""
	}

	var message string
	if t.Task.Message != "" {
		message = "- " + t.Task.Message
	}

	var author string
	if t.Task.UserID != nil {
		user, err := t.pool.store.GetUser(*t.Task.UserID)
		if err != nil {
			panic(err)
		}
		author = user.Name
	}

	alert := Alert{
		TaskID:          strconv.Itoa(t.Task.ID),
		Name:            t.Template.Name,
		TaskURL:         util.Config.WebHost + "/project/" + strconv.Itoa(t.Template.ProjectID) + "/templates/" + strconv.Itoa(t.Template.ID) + "?t=" + strconv.Itoa(t.Task.ID),
		ChatID:          chatID,
		TaskResult:      strings.ToUpper(string(t.Task.Status)),
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

	httpTransport := &http.Transport{}
	if len(util.Config.AlertUrlProxy) != 0 { // Set the proxy only if the proxy param is specified
		alertUrlProxy, proxyErr := url.Parse(util.Config.AlertUrlProxy)
		if proxyErr == nil {
			httpTransport.Proxy = http.ProxyURL(alertUrlProxy)
		}
		if proxyErr != nil {
			t.Log("Can't send slack alert! Error: " + proxyErr.Error())
		}
	}
	http := http.Client{Transport: httpTransport}

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

	if t.Template.SuppressSuccessAlerts && t.Task.Status == lib.TaskSuccessStatus {
		return
	}

	slackUrl := util.Config.SlackUrl

	httpTransport := &http.Transport{}
	if len(util.Config.AlertUrlProxy) != 0 { // Set the proxy only if the proxy param is specified
		alertUrlProxy, proxyErr := url.Parse(util.Config.AlertUrlProxy)
		if proxyErr == nil {
			httpTransport.Proxy = http.ProxyURL(alertUrlProxy)
		}
		if proxyErr != nil {
			t.Log("Can't send slack alert! Error: " + proxyErr.Error())
		}
	}
	http := http.Client{Transport: httpTransport}

	var slackBuffer bytes.Buffer

	var version string
	if t.Task.Version != nil {
		version = *t.Task.Version
	} else if t.Task.BuildTaskID != nil {
		version = "build " + strconv.Itoa(*t.Task.BuildTaskID)
	} else {
		version = ""
	}

	var message string
	if t.Task.Message != "" {
		message = "- " + t.Task.Message
	}

	var author string
	if t.Task.UserID != nil {
		user, err := t.pool.store.GetUser(*t.Task.UserID)
		if err != nil {
			panic(err)
		}
		author = user.Name
	}

	var color string
	if t.Task.Status == lib.TaskSuccessStatus {
		color = "good"
	} else if t.Task.Status == lib.TaskFailStatus {
		color = "bad"
	} else if t.Task.Status == lib.TaskRunningStatus {
		color = "#333CFF"
	} else if t.Task.Status == lib.TaskWaitingStatus {
		color = "#FFFC33"
	} else if t.Task.Status == lib.TaskStoppingStatus {
		color = "#BEBEBE"
	} else if t.Task.Status == lib.TaskStoppedStatus {
		color = "#5B5B5B"
	}
	alert := Alert{
		TaskID:          strconv.Itoa(t.Task.ID),
		Name:            t.Template.Name,
		TaskURL:         util.Config.WebHost + "/project/" + strconv.Itoa(t.Template.ProjectID) + "/templates/" + strconv.Itoa(t.Template.ID) + "?t=" + strconv.Itoa(t.Task.ID),
		TaskResult:      strings.ToUpper(string(t.Task.Status)),
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
