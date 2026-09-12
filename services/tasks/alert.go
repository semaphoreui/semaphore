package tasks

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	htmltemplate "html/template"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
	"github.com/semaphoreui/semaphore/util/mailer"
)

//go:embed templates/*.tmpl
var templates embed.FS

// Alert represents an alert that will be templated and sent to the appropriate service
type Alert struct {
	Name         string
	Author       string
	Color        string
	Task         alertTask
	Chat         alertChat
	Playbook     string
	Project      alertProjectInfo
	ScheduleName string
}

type alertTask struct {
	ID       string
	URL      string
	Result   string
	Desc     string
	Version  string
	Duration string
	Trigger  string
}

type alertProjectInfo struct {
	ID   int
	Name string
}

type alertChat struct {
	ID       string
	ThreadID string
}

func (t *TaskRunner) alertEvents() (onSuccess bool, onError bool) {
	if t.Task.AlertSnapshot != nil {
		return t.Task.AlertSnapshot.OnSuccess, t.Task.AlertSnapshot.OnError
	}
	return db.ResolveAlertOn(t.Template.AlertOnSuccess, t.Template.SuppressSuccessAlerts),
		db.ResolveAlertOn(t.Template.AlertOnError, t.Template.SuppressErrorAlerts)
}

func (t *TaskRunner) shouldSkipStatusAlert() bool {
	onSuccess, onError := t.alertEvents()
	return shouldSkipAlertForStatus(t.Task.Status, onSuccess, onError)
}

func shouldSkipAlertForStatus(status task_logger.TaskStatus, onSuccess bool, onError bool) bool {
	switch status {
	case task_logger.TaskSuccessStatus:
		return !onSuccess
	case task_logger.TaskFailStatus:
		return !onError
	default:
		return true
	}
}

func (t *TaskRunner) sendStatusAlerts() {
	if t.Task.AlertSnapshot == nil {
		return
	}
	snapshot := *t.Task.AlertSnapshot
	status := t.Task.Status
	go t.sendResolvedAlerts(snapshot, status)
}

func (t *TaskRunner) sendResolvedAlerts(snapshot db.AlertSnapshot, status task_logger.TaskStatus) {
	if shouldSkipAlertForStatus(status, snapshot.OnSuccess, snapshot.OnError) {
		return
	}
	if len(snapshot.AlertIDs) == 0 {
		return
	}
	if t.pool == nil || t.pool.store == nil {
		return
	}

	sent := make(map[int]bool)
	for _, alertID := range snapshot.AlertIDs {
		if sent[alertID] {
			continue
		}
		sent[alertID] = true

		alert, err := t.pool.store.GetAlert(t.Task.ProjectID, alertID)
		if err != nil {
			t.Logf("Can't load alert %d: %s", alertID, err.Error())
			continue
		}
		if !alert.Enabled {
			continue
		}
		if t.Task.ID > 0 {
			claimed, claimErr := t.pool.store.ClaimAlertSend(t.Task.ID, alertID, string(status))
			if claimErr != nil {
				t.Logf("Can't claim alert %d: %s", alertID, claimErr.Error())
				continue
			}
			if !claimed {
				continue
			}
		}
		if err := t.sendProjectAlertForStatus(alert, status); err != nil {
			t.Logf("%s", err.Error())
		}
	}
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
			t.emailSubject(),
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

	chatID := util.Config.TelegramChat
	if t.alertChat != nil && *t.alertChat != "" {
		chatID = *t.alertChat
	}

	if chatID == "" {
		return
	}

	body := bytes.NewBufferString("")
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
			ID: chatID,
		},
	}

	tpl, err := template.ParseFS(templates, "templates/telegram.tmpl")

	if err != nil {
		t.Log("Can't parse telegram alert template!")
		panic(err)
	}

	if err := tpl.Execute(body, alert); err != nil {
		t.Log("Can't generate telegram alert template!")
		panic(err)
	}

	if body.Len() == 0 {
		t.Log("Buffer for telegram alert is empty")
		return
	}

	payload, err := wrapTelegramMessage(chatID, "", body.String())
	if err != nil {
		t.Log("Can't wrap telegram alert! Error: " + err.Error())
		return
	}

	t.Log("Attempting to send telegram alert")

	resp, err := http.Post(
		fmt.Sprintf(
			"https://api.telegram.org/bot%s/sendMessage",
			util.Config.TelegramToken,
		),
		"application/json",
		strings.NewReader(payload),
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

	if t.Task.UserID != nil && t.pool != nil && t.pool.store != nil {
		user, err := t.pool.store.GetUser(*t.Task.UserID)
		if err == nil {
			author = user.Name
		}
	}

	return author, version
}

func (t *TaskRunner) alertColor(kind string) string {
	return alertColorFor(kind, t.Task.Status)
}

func alertColorFor(kind string, status task_logger.TaskStatus) string {
	switch kind {
	case "slack":
		switch status {
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
		switch status {
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

func (t *TaskRunner) emailSubject() string {
	return t.emailSubjectFor(t.Task.Status)
}

func (t *TaskRunner) emailSubjectFor(status task_logger.TaskStatus) string {
	result := status.Format()
	if result == "" {
		result = string(status)
	}
	return fmt.Sprintf("Task '%s' %s", t.Template.Name, result)
}

func (t *TaskRunner) alertTrigger() string {
	if t.Task.ScheduleID != nil {
		return "schedule"
	}
	if t.Task.IntegrationID != nil {
		return "integration"
	}
	if t.Task.UserID != nil {
		return "manual"
	}
	return "api"
}

func (t *TaskRunner) newAlertPayload(kind string, chatID string, threadID string) Alert {
	return t.newAlertPayloadAt(kind, chatID, threadID, t.Task.Status)
}

func (t *TaskRunner) newAlertPayloadAt(kind string, chatID string, threadID string, status task_logger.TaskStatus) Alert {
	author, version := t.alertInfos()
	duration := t.alertDuration()

	return Alert{
		Name:         t.Template.Name,
		Author:       author,
		Color:        alertColorFor(kind, status),
		ScheduleName: t.alertScheduleName(),
		Task: alertTask{
			ID:       strconv.Itoa(t.Task.ID),
			URL:      t.taskLink(),
			Result:   status.Format(),
			Version:  version,
			Desc:     t.Task.Message,
			Duration: duration,
			Trigger:  t.alertTrigger(),
		},
		Chat: alertChat{
			ID:       chatID,
			ThreadID: threadID,
		},
		Playbook: t.Template.Playbook,
		Project: alertProjectInfo{
			ID:   t.Template.ProjectID,
			Name: t.alertProjectName(),
		},
	}
}

func (t *TaskRunner) alertDuration() string {
	if t.Task.Start == nil {
		return ""
	}
	end := t.Task.End
	if end == nil {
		now := tz.Now()
		end = &now
	}
	return end.Sub(*t.Task.Start).String()
}

func (t *TaskRunner) alertScheduleName() string {
	if t.Task.ScheduleID == nil || t.pool == nil || t.pool.store == nil {
		return ""
	}
	schedule, err := t.pool.store.GetSchedule(t.Task.ProjectID, *t.Task.ScheduleID)
	if err != nil {
		return ""
	}
	return schedule.Name
}

func (t *TaskRunner) alertProjectName() string {
	if t.pool == nil || t.pool.store == nil {
		return ""
	}
	project, err := t.pool.store.GetProject(t.Template.ProjectID)
	if err != nil {
		return ""
	}
	return project.Name
}

// projectAlertReady reports whether a project alert can be sent.
// Email and Telegram always need the instance SMTP/bot config.
// Webhook destinations must be set on the alert. Gotify may use the
// instance URL/token pair only when both alert fields are empty.
func projectAlertReady(alert db.Alert) error {
	if util.Config == nil {
		return common_errors.NewValidationError("server config is not loaded")
	}

	switch alert.Type {
	case db.AlertTypeEmail:
		if !util.Config.EmailAlert || util.Config.EmailHost == "" {
			return common_errors.NewValidationError("email is not configured on the server")
		}
	case db.AlertTypeTelegram:
		if stringValue(alert.ChatID) == "" {
			return common_errors.NewValidationError("telegram chat id can not be empty")
		}
		if !util.Config.TelegramAlert || util.Config.TelegramToken == "" {
			return common_errors.NewValidationError("telegram is not configured on the server")
		}
	case db.AlertTypeSlack:
		return webhookDestinationReady(stringValue(alert.URL), "slack")
	case db.AlertTypeTeams:
		return webhookDestinationReady(stringValue(alert.URL), "microsoft teams")
	case db.AlertTypeRocketChat:
		return webhookDestinationReady(stringValue(alert.URL), "rocketchat")
	case db.AlertTypeDingTalk:
		return webhookDestinationReady(stringValue(alert.URL), "dingtalk")
	case db.AlertTypeGotify:
		return gotifyDestinationReady(stringValue(alert.URL), stringValue(alert.Token))
	default:
		return common_errors.NewValidationError("invalid alert type")
	}
	return nil
}

func webhookDestinationReady(alertURL, channel string) error {
	if alertURL != "" {
		return nil
	}
	return common_errors.NewValidationError(channel + " webhook URL is missing")
}

func gotifyDestinationReady(alertURL, alertToken string) error {
	if alertURL != "" {
		if alertToken == "" {
			return common_errors.NewValidationError("gotify token is missing for the alert URL")
		}
		return db.ValidateGotifyURL(alertURL)
	}
	if alertToken != "" {
		if util.Config.GotifyUrl == "" {
			return common_errors.NewValidationError("gotify URL is missing for the alert token")
		}
		return db.ValidateGotifyURL(util.Config.GotifyUrl)
	}
	if !util.Config.GotifyAlert || util.Config.GotifyUrl == "" || util.Config.GotifyToken == "" {
		return common_errors.NewValidationError("gotify is not configured on the server")
	}
	return db.ValidateGotifyURL(util.Config.GotifyUrl)
}

func resolveGotifyDestination(alert db.Alert) (url string, token string, err error) {
	url = stringValue(alert.URL)
	token = stringValue(alert.Token)
	if url != "" {
		if token == "" {
			return "", "", common_errors.NewValidationError("gotify token is missing for the alert URL")
		}
		return url, token, db.ValidateGotifyURL(url)
	}
	url = util.Config.GotifyUrl
	if token == "" {
		token = util.Config.GotifyToken
	}
	if url == "" || token == "" {
		return "", "", common_errors.NewValidationError("gotify URL or token is missing")
	}
	return url, token, db.ValidateGotifyURL(url)
}

func (t *TaskRunner) sendProjectAlert(alert db.Alert) error {
	return t.sendProjectAlertForStatus(alert, t.Task.Status)
}

func (t *TaskRunner) sendProjectAlertForStatus(alert db.Alert, status task_logger.TaskStatus) error {
	if err := projectAlertReady(alert); err != nil {
		return err
	}

	chatID := stringValue(alert.ChatID)
	payload := t.newAlertPayloadAt(string(alert.Type), chatID, stringValue(alert.ThreadID), status)
	body, err := t.renderAlertBody(alert, payload)
	if err != nil {
		t.Logf("Can't render alert %s: %s", alert.Name, err.Error())
		return err
	}
	if body == "" {
		t.Logf("Buffer for alert %s is empty", alert.Name)
		return common_errors.NewValidationError("alert message is empty")
	}

	switch alert.Type {
	case db.AlertTypeEmail:
		return t.sendProjectMailAlert(alert, body, status)
	case db.AlertTypeTelegram:
		wrapped, wrapErr := wrapTelegramMessage(chatID, stringValue(alert.ThreadID), body)
		if wrapErr != nil {
			return wrapErr
		}
		return t.postAlertJSON(
			alert.Name,
			fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", util.Config.TelegramToken),
			wrapped,
			[]int{200},
		)
	case db.AlertTypeSlack:
		return t.postAlertJSON(alert.Name, stringValue(alert.URL), body, []int{200})
	case db.AlertTypeRocketChat:
		return t.postAlertJSON(alert.Name, stringValue(alert.URL), body, []int{200})
	case db.AlertTypeTeams:
		return t.postAlertJSON(alert.Name, stringValue(alert.URL), body, []int{200, 202})
	case db.AlertTypeDingTalk:
		return t.postAlertJSON(alert.Name, stringValue(alert.URL), body, []int{200})
	case db.AlertTypeGotify:
		url, token, destErr := resolveGotifyDestination(alert)
		if destErr != nil {
			return destErr
		}
		return t.postAlertJSONWithHeaders(
			alert.Name,
			strings.TrimRight(url, "/")+"/message",
			body,
			[]int{200},
			map[string]string{"X-Gotify-Key": token},
		)
	default:
		return common_errors.NewValidationError("invalid alert type")
	}
}

func (t *TaskRunner) renderAlertBody(alert db.Alert, payload Alert) (string, error) {
	if alert.Body != nil && strings.TrimSpace(*alert.Body) != "" {
		if alert.Type == db.AlertTypeEmail {
			tpl, err := htmltemplate.New("alert").Parse(*alert.Body)
			if err != nil {
				return "", err
			}
			var buf bytes.Buffer
			if err = tpl.Execute(&buf, payload); err != nil {
				return "", err
			}
			return buf.String(), nil
		}
		tpl, err := template.New("alert").Parse(*alert.Body)
		if err != nil {
			return "", err
		}
		var buf bytes.Buffer
		if err = tpl.Execute(&buf, payload); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	tmplName := builtinAlertTemplate(alert.Type)
	if alert.Type == db.AlertTypeEmail {
		tpl, err := htmltemplate.ParseFS(templates, tmplName)
		if err != nil {
			return "", err
		}
		var buf bytes.Buffer
		if err = tpl.Execute(&buf, payload); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	tpl, err := template.ParseFS(templates, tmplName)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err = tpl.Execute(&buf, payload); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// DefaultAlertBodies returns the built-in message template for each channel
// so the UI can show the same text the sender uses when body is empty.
func DefaultAlertBodies() (map[db.AlertType]string, error) {
	types := []db.AlertType{
		db.AlertTypeEmail,
		db.AlertTypeTelegram,
		db.AlertTypeSlack,
		db.AlertTypeTeams,
		db.AlertTypeRocketChat,
		db.AlertTypeDingTalk,
		db.AlertTypeGotify,
	}
	out := make(map[db.AlertType]string, len(types))
	for _, alertType := range types {
		name := builtinAlertTemplate(alertType)
		body, err := templates.ReadFile(name)
		if err != nil {
			return nil, err
		}
		out[alertType] = string(body)
	}
	return out, nil
}

func wrapTelegramMessage(chatID, threadID, text string) (string, error) {
	msg := map[string]any{
		"chat_id":    chatID,
		"parse_mode": "HTML",
		"text":       text,
	}
	if err := db.ValidateTelegramThreadID(threadID); err != nil {
		return "", err
	}
	if n, err := strconv.Atoi(strings.TrimSpace(threadID)); err == nil {
		msg["message_thread_id"] = n
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func builtinAlertTemplate(alertType db.AlertType) string {
	switch alertType {
	case db.AlertTypeEmail:
		return "templates/email.tmpl"
	case db.AlertTypeTelegram:
		return "templates/telegram.tmpl"
	case db.AlertTypeSlack:
		return "templates/slack.tmpl"
	case db.AlertTypeTeams:
		return "templates/microsoft-teams.tmpl"
	case db.AlertTypeRocketChat:
		return "templates/rocketchat.tmpl"
	case db.AlertTypeDingTalk:
		return "templates/dingtalk.tmpl"
	case db.AlertTypeGotify:
		return "templates/gotify.tmpl"
	default:
		return ""
	}
}

func (t *TaskRunner) sendProjectMailAlert(alert db.Alert, body string, status task_logger.TaskStatus) error {
	recipients := t.mailRecipients(alert)
	if len(recipients) == 0 {
		return common_errors.NewValidationError("no email recipients")
	}
	var firstErr error
	for _, email := range recipients {
		t.Logf("Attempting to send email alert to %s", email)
		if err := mailer.Send(
			util.Config.EmailSecure,
			util.Config.EmailTls,
			util.Config.EmailHost,
			util.Config.EmailPort,
			util.Config.EmailUsername,
			util.Config.EmailPassword,
			util.Config.EmailSender,
			email,
			t.emailSubjectFor(status),
			body,
		); err != nil {
			util.LogError(err)
			if firstErr == nil {
				firstErr = common_errors.NewUserError(err)
			}
			continue
		}
		t.Logf("Sent successfully email alert to %s", email)
	}
	return firstErr
}

func (t *TaskRunner) mailRecipients(alert db.Alert) []string {
	if alert.Recipients != nil && strings.TrimSpace(*alert.Recipients) != "" {
		parts := strings.Split(*alert.Recipients, ",")
		var out []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}

	var emails []string
	for _, uid := range t.users {
		if t.pool == nil || t.pool.store == nil {
			continue
		}
		user, err := t.pool.store.GetUser(uid)
		if err != nil || !user.Alert || user.Email == "" {
			continue
		}
		emails = append(emails, user.Email)
	}
	return emails
}

func alertHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           alertDialContext,
			ForceAttemptHTTP2:     true,
			TLSHandshakeTimeout:   10 * time.Second,
			IdleConnTimeout:       30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func alertDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	allowed := db.AllowedAlertIPs(ips)
	if len(allowed) == 0 {
		return nil, common_errors.NewValidationError("alert URL host is not allowed")
	}
	dialer := net.Dialer{}
	var lastErr error
	for _, ip := range allowed {
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (t *TaskRunner) postAlertJSON(name, url, body string, okCodes []int) error {
	return t.postAlertJSONWithHeaders(name, url, body, okCodes, nil)
}

func (t *TaskRunner) postAlertJSONWithHeaders(name, url, body string, okCodes []int, headers map[string]string) error {
	if url == "" {
		t.Logf("Can't send alert %s: empty URL", name)
		return common_errors.NewValidationError("alert URL is empty")
	}
	if err := db.ValidateAlertURL(url); err != nil {
		t.Logf("Can't send alert %s: %s", name, err.Error())
		return err
	}

	t.Logf("Attempting to send alert %s", name)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Logf("Can't send alert %s: %s", name, err.Error())
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := alertHTTPClient().Do(req)
	if err != nil {
		t.Logf("Can't send alert %s: %s", name, err.Error())
		return common_errors.NewUserError(err)
	}
	defer resp.Body.Close() //nolint:errcheck
	_, _ = io.Copy(io.Discard, resp.Body)

	for _, code := range okCodes {
		if resp.StatusCode == code {
			t.Logf("Sent successfully alert %s", name)
			return nil
		}
	}
	t.Logf("Can't send alert %s! Response code: %d", name, resp.StatusCode)
	return common_errors.NewUserErrorS(fmt.Sprintf("alert %s: unexpected response code %d", name, resp.StatusCode))
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}
