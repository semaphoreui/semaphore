package tasks

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/semaphoreui/semaphore/util/mailer"
)

//go:embed templates/*.tmpl
var templates embed.FS

// Alert represents an alert that will be templated and sent to the appropriate service
type Alert struct {
	Name   string
	Author string
	Color  string
	Task   alertTask
	Chat   alertChat
}

type alertTask struct {
	ID      string
	URL     string
	Result  string
	Desc    string
	Version string
}

type alertChat struct {
	ID       string
	ThreadID string
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func normalizeTelegramThreadID(threadID string) string {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return ""
	}
	for _, r := range threadID {
		if r < '0' || r > '9' {
			return ""
		}
	}
	return threadID
}

func telegramChatAndThread(globalChat, globalThread string, projectChat, projectThread *string) (chatID, rawThread string) {
	projChat := stringValue(projectChat)
	projThread := stringValue(projectThread)
	globalChat = strings.TrimSpace(globalChat)
	globalThread = strings.TrimSpace(globalThread)

	if projChat != "" {
		return projChat, projThread
	}
	if projThread != "" {
		return globalChat, projThread
	}
	return globalChat, globalThread
}

func resolveTelegramDestination(globalChat, globalThread string, projectChat, projectThread *string) (chatID, threadID string) {
	chatID, rawThread := telegramChatAndThread(globalChat, globalThread, projectChat, projectThread)
	return chatID, normalizeTelegramThreadID(rawThread)
}

func renderTelegramAlert(alert Alert) ([]byte, error) {
	tpl, err := template.ParseFS(templates, "templates/telegram.tmpl")
	if err != nil {
		return nil, err
	}

	body := bytes.NewBuffer(nil)
	if err := tpl.Execute(body, alert); err != nil {
		return nil, err
	}

	return body.Bytes(), nil
}

func (t *TaskRunner) shouldSkipStatusAlert() bool {
	if t.Template.SuppressSuccessAlerts && t.Task.Status == task_logger.TaskSuccessStatus {
		return true
	}
	if t.Template.SuppressErrorAlerts && t.Task.Status == task_logger.TaskFailStatus {
		return true
	}
	return false
}

func (t *TaskRunner) sendMailAlert() {
	if !util.Config.EmailAlert || !t.alert {
		return
	}

	if t.shouldSkipStatusAlert() {
		return
	}

	body := bytes.NewBufferString("")
	author, version := t.alertInfos()

	alert := Alert{
		Name:   t.Template.Name,
		Author: author,
		Color:  t.alertColor("email"),
		Task: alertTask{
			ID:      strconv.Itoa(t.Task.ID),
			URL:     t.taskLink(),
			Result:  t.Task.Status.Format(),
			Version: version,
			Desc:    t.Task.Message,
		},
	}

	tpl, err := htmltemplate.ParseFS(templates, "templates/email.tmpl")

	if err != nil {
		t.Log("Can't parse email alert template!")
		panic(err)
	}

	if err := tpl.Execute(body, alert); err != nil {
		t.Log("Can't generate email alert template!")
		panic(err)
	}

	if body.Len() == 0 {
		t.Log("Buffer for email alert is empty")
		return
	}

	for _, uid := range t.users {
		user, err := t.pool.store.GetUser(uid)

		if err != nil {
			util.LogError(err)
			continue
		}

		if !user.Alert {
			continue
		}

		t.Logf("Attempting to send email alert to %s", user.Email)

		str := body.String()
		if err := mailer.Send(
			util.Config.EmailSecure,
			util.Config.EmailTls,
			util.Config.EmailHost,
			util.Config.EmailPort,
			util.Config.EmailUsername,
			util.Config.EmailPassword,
			util.Config.EmailSender,
			user.Email,
			fmt.Sprintf("Task '%s' failed", t.Template.Name),
			str,
		); err != nil {
			util.LogError(err)
			continue
		}

		t.Logf("Sent successfully email alert to %s", user.Email)
	}
}

func (t *TaskRunner) sendTelegramAlert() {
	if !util.Config.TelegramAlert || !t.alert {
		return
	}

	if t.shouldSkipStatusAlert() {
		return
	}

	chatID, rawThread := telegramChatAndThread(
		util.Config.TelegramChat,
		util.Config.TelegramThread,
		t.alertChat,
		t.alertThread,
	)
	threadID := normalizeTelegramThreadID(rawThread)
	if rawThread != "" && threadID == "" {
		t.Log("Invalid Telegram thread ID, sending without message_thread_id")
	}

	if chatID == "" {
		return
	}

	author, version := t.alertInfos()

	alert := Alert{
		Name:   t.Template.Name,
		Author: author,
		Color:  t.alertColor("telegram"),
		Task: alertTask{
			ID:      strconv.Itoa(t.Task.ID),
			URL:     t.taskLink(),
			Result:  t.Task.Status.Format(),
			Version: version,
			Desc:    t.Task.Message,
		},
		Chat: alertChat{
			ID:       chatID,
			ThreadID: threadID,
		},
	}

	payload, err := renderTelegramAlert(alert)
	if err != nil {
		t.Log("Can't generate telegram alert template!")
		panic(err)
	}

	if len(payload) == 0 {
		t.Log("Buffer for telegram alert is empty")
		return
	}

	body := bytes.NewReader(payload)

	t.Log("Attempting to send telegram alert")

	resp, err := http.Post(
		fmt.Sprintf(
			"https://api.telegram.org/bot%s/sendMessage",
			util.Config.TelegramToken,
		),
		"application/json",
		body,
	)

	if err != nil {
		t.Log("Can't send telegram alert! Error: " + err.Error())
	} else if resp.StatusCode != 200 {
		t.Log("Can't send telegram alert! Response code: " + strconv.Itoa(resp.StatusCode))
	} else {
		t.Log("Sent successfully telegram alert")
	}

	if resp != nil {
		defer resp.Body.Close() //nolint:errcheck
	}
}

func (t *TaskRunner) sendSlackAlert() {
	if !util.Config.SlackAlert || !t.alert {
		return
	}

	if t.shouldSkipStatusAlert() {
		return
	}

	body := bytes.NewBufferString("")
	author, version := t.alertInfos()

	alert := Alert{
		Name:   t.Template.Name,
		Author: author,
		Color:  t.alertColor("slack"),
		Task: alertTask{
			ID:      strconv.Itoa(t.Task.ID),
			URL:     t.taskLink(),
			Result:  t.Task.Status.Format(),
			Version: version,
			Desc:    t.Task.Message,
		},
	}

	tpl, err := template.ParseFS(templates, "templates/slack.tmpl")

	if err != nil {
		t.Log("Can't parse slack alert template!")
		panic(err)
	}

	if err := tpl.Execute(body, alert); err != nil {
		t.Log("Can't generate slack alert template!")
		panic(err)
	}

	if body.Len() == 0 {
		t.Log("Buffer for slack alert is empty")
		return
	}

	t.Log("Attempting to send slack alert")

	resp, err := http.Post(
		util.Config.SlackUrl,
		"application/json",
		body,
	)

	if err != nil {
		t.Log("Can't send slack alert! Error: " + err.Error())
	} else if resp.StatusCode != 200 {
		t.Log("Can't send slack alert! Response code: " + strconv.Itoa(resp.StatusCode))
	} else {
		t.Log("Sent successfully slack alert")
	}

	if resp != nil {
		defer resp.Body.Close() //nolint:errcheck
	}
}

func (t *TaskRunner) sendRocketChatAlert() {
	if !util.Config.RocketChatAlert || !t.alert {
		return
	}

	if t.shouldSkipStatusAlert() {
		return
	}

	body := bytes.NewBufferString("")
	author, version := t.alertInfos()

	alert := Alert{
		Name:   t.Template.Name,
		Author: author,
		Color:  t.alertColor("rocketchat"),
		Task: alertTask{
			ID:      strconv.Itoa(t.Task.ID),
			URL:     t.taskLink(),
			Result:  t.Task.Status.Format(),
			Version: version,
			Desc:    t.Task.Message,
		},
	}

	tpl, err := template.ParseFS(templates, "templates/rocketchat.tmpl")

	if err != nil {
		t.Log("Can't parse rocketchat alert template!")
		panic(err)
	}

	if err := tpl.Execute(body, alert); err != nil {
		t.Log("Can't generate rocketchat alert template!")
		panic(err)
	}

	if body.Len() == 0 {
		t.Log("Buffer for rocketchat alert is empty")
		return
	}

	t.Log("Attempting to send rocketchat alert")

	resp, err := http.Post(
		util.Config.RocketChatUrl,
		"application/json",
		body,
	)

	if err != nil {
		t.Log("Can't send rocketchat alert! Error: " + err.Error())
	} else if resp.StatusCode != 200 {
		t.Log("Can't send rocketchat alert! Response code: " + strconv.Itoa(resp.StatusCode))
	} else {
		t.Log("Sent successfully rocketchat alert")
	}
	if resp != nil {
		defer resp.Body.Close() //nolint:errcheck
	}
}

func (t *TaskRunner) sendMicrosoftTeamsAlert() {
	if !util.Config.MicrosoftTeamsAlert || !t.alert {
		return
	}

	if t.shouldSkipStatusAlert() {
		return
	}

	body := bytes.NewBufferString("")
	author, version := t.alertInfos()

	alert := Alert{
		Name:   t.Template.Name,
		Author: author,
		Color:  t.alertColor("microsoft-teams"),
		Task: alertTask{
			ID:      strconv.Itoa(t.Task.ID),
			URL:     t.taskLink(),
			Result:  t.Task.Status.Format(),
			Version: version,
			Desc:    t.Task.Message,
		},
	}

	tpl, err := template.ParseFS(templates, "templates/microsoft-teams.tmpl")

	if err != nil {
		t.Log("Can't parse microsoft teams alert template!")
		panic(err)
	}

	if err := tpl.Execute(body, alert); err != nil {
		t.Log("Can't generate microsoft teams alert template!")
		panic(err)
	}

	if body.Len() == 0 {
		t.Log("Buffer for microsoft teams alert is empty")
		return
	}

	t.Log("Attempting to send microsoft teams alert")

	resp, err := http.Post(
		util.Config.MicrosoftTeamsUrl,
		"application/json",
		body,
	)

	if err != nil {
		t.Log("Can't send microsoft teams alert! Error: " + err.Error())
	} else if resp.StatusCode != 200 && resp.StatusCode != 202 {
		t.Log("Can't send microsoft teams alert! Response code: " + strconv.Itoa(resp.StatusCode))
	} else {
		t.Log("Sent successfully microsoft teams alert")
	}
	if resp != nil {
		defer resp.Body.Close() //nolint:errcheck
	}
}

func (t *TaskRunner) sendDingTalkAlert() {
	if !util.Config.DingTalkAlert || !t.alert {
		return
	}

	if t.shouldSkipStatusAlert() {
		return
	}

	body := bytes.NewBufferString("")
	author, version := t.alertInfos()

	alert := Alert{
		Name:   t.Template.Name,
		Author: author,
		Color:  t.alertColor("dingtalk"),
		Task: alertTask{
			ID:      strconv.Itoa(t.Task.ID),
			URL:     t.taskLink(),
			Result:  t.Task.Status.Format(),
			Version: version,
			Desc:    t.Task.Message,
		},
	}

	tpl, err := template.ParseFS(templates, "templates/dingtalk.tmpl")

	if err != nil {
		t.Log("Can't parse dingtalk alert template!")
		panic(err)
	}

	if err := tpl.Execute(body, alert); err != nil {
		t.Log("Can't generate dingtalk alert template!")
		panic(err)
	}

	if body.Len() == 0 {
		t.Log("Buffer for dingtalk alert is empty")
		return
	}

	t.Log("Attempting to send dingtalk alert")

	resp, err := http.Post(
		util.Config.DingTalkUrl,
		"application/json",
		body,
	)

	if err != nil {
		t.Log("Can't send dingtalk alert! Error: " + err.Error())
	} else if resp.StatusCode != 200 {
		t.Log("Can't send dingtalk alert! Response code: " + strconv.Itoa(resp.StatusCode))
	} else {
		t.Log("Sent successfully dingtalk alert")
	}

	if resp != nil {
		defer resp.Body.Close() //nolint:errcheck
	}
}

func (t *TaskRunner) sendGotifyAlert() {
	if !util.Config.GotifyAlert || !t.alert {
		return
	}

	if t.shouldSkipStatusAlert() {
		return
	}

	body := bytes.NewBufferString("")
	author, version := t.alertInfos()

	alert := Alert{
		Name:   t.Template.Name,
		Author: author,
		Color:  t.alertColor("gotify"),
		Task: alertTask{
			ID:      strconv.Itoa(t.Task.ID),
			URL:     t.taskLink(),
			Result:  t.Task.Status.Format(),
			Version: version,
			Desc:    t.Task.Message,
		},
	}

	tpl, err := template.ParseFS(templates, "templates/gotify.tmpl")

	if err != nil {
		t.Log("Can't parse gotify alert template!")
		panic(err)
	}

	if err := tpl.Execute(body, alert); err != nil {
		t.Log("Can't generate gotify alert template!")
		panic(err)
	}

	if body.Len() == 0 {
		t.Log("Buffer for gotify alert is empty")
		return
	}

	t.Log("Attempting to send gotify alert")

	resp, err := http.Post(
		fmt.Sprintf(
			"%s/message?token=%s",
			util.Config.GotifyUrl,
			util.Config.GotifyToken),
		"application/json",
		body,
	)

	if err != nil {
		t.Log("Can't send gotify alert! Error: " + err.Error())
	} else if resp.StatusCode != 200 {
		t.Log("Can't send gotify alert! Response code: " + strconv.Itoa(resp.StatusCode))
	} else {
		t.Log("Sent successfully gotify alert")
	}

	if resp != nil {
		defer resp.Body.Close() //nolint:errcheck
	}
}

func (t *TaskRunner) alertInfos() (string, string) {
	version := ""

	if t.Task.Version != nil {
		version = *t.Task.Version
	} else if t.Template.Type != db.TemplateTask {
		v := t.Task.GetIncomingVersion(t.pool.store)
		if v != nil {
			version = "build " + *v
		} else {
			version = ""
		}
	} else {
		version = ""
	}

	author := "—"

	if t.Task.UserID != nil {
		user, err := t.pool.store.GetUser(*t.Task.UserID)

		if err != nil {
			panic(err)
		}

		author = user.Name
	}

	return author, version
}

func (t *TaskRunner) alertColor(kind string) string {
	switch kind {
	case "slack":
		switch t.Task.Status {
		case task_logger.TaskSuccessStatus:
			return "good"
		case task_logger.TaskFailStatus:
			return "danger"
		case task_logger.TaskRunningStatus:
			return "#333CFF"
		case task_logger.TaskWaitingStatus:
			return "#FFFC33"
		case task_logger.TaskStoppingStatus:
			return "#BEBEBE"
		case task_logger.TaskStoppedStatus:
			return "#5B5B5B"
		}
	case "rocketchat":
		switch t.Task.Status {
		case task_logger.TaskSuccessStatus:
			return "#00EE00"
		case task_logger.TaskFailStatus:
			return "#EE0000"
		case task_logger.TaskRunningStatus:
			return "#333CFF"
		case task_logger.TaskWaitingStatus:
			return "#FFFC33"
		case task_logger.TaskStoppingStatus:
			return "#BEBEBE"
		case task_logger.TaskStoppedStatus:
			return "#5B5B5B"
		}
	}

	return ""
}

func (t *TaskRunner) taskLink() string {
	return fmt.Sprintf(
		"%s/project/%d/templates/%d?t=%d",
		util.Config.WebHost,
		t.Template.ProjectID,
		t.Template.ID,
		t.Task.ID,
	)
}
