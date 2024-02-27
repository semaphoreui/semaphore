package tasks

import (
	"bytes"
	"embed"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/ansible-semaphore/semaphore/lib"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/ansible-semaphore/semaphore/util/mailer"
)

//go:embed templates/*.tmpl
var templates embed.FS

const microsoftTeamsTemplate = `{
	"type": "message",
	"attachments": [
		{
			"contentType": "application/vnd.microsoft.card.adaptive",
			"content": {
				"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
				"type": "AdaptiveCard",
				"version": "1.5",
				"body": [
					{
						"type": "TextBlock",
						"text": "Ansible Task Template Execution by: {{ .Author }}",
					},
					{
						"type": "FactSet",
						"facts": [
						  {
							"title": "Task:",
							"value": "{{ .Name }}"
						  },
						  {
							"title": "Status:",
							"value": "{{ .TaskResult }}"
						  },
						  {
							"title": "Task ID:",
							"value": "{{ .TaskID }}"
						  }
						],
						"separator": true
					}
				],
				"actions": [
					{
						"type": "Action.OpenUrl",
						"title": "Task URL",
						"url": "{{ .TaskURL }}"
					}
				],
				"msteams": {
					"width": "Full"
				},
				"backgroundImage": {
					"horizontalAlignment": "Center",
					"url": "data:image/jpg;base64,iVBORw0KGgoAAAANSUhEUgAABSgAAAAFCAYAAABGmwLHAAAARklEQVR4nO3YMQEAIBDEsANPSMC/AbzwMm5JJHTseuf+AAAAAAAUbNEBAAAAgBaDEgAAAACoMSgBAAAAgBqDEgAAAADoSDL8RAJfcbcsoQAAAABJRU5ErkJggg==",
					"fillMode": "RepeatHorizontally"
				}
			}
		}
	]
}`

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
	ID string
}

func (t *TaskRunner) sendMailAlert() {
	if !util.Config.EmailAlert || !t.alert {
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
			Result:  strings.ToUpper(string(t.Task.Status)),
			Version: version,
			Desc:    t.Task.Message,
		},
	}

	tpl, err := template.ParseFS(templates, "templates/email.tmpl")

	if err != nil {
		t.Log("Can't parse email alert template!")
		panic(err)
	}

	if err := tpl.Execute(body, alert); err != nil {
		t.Log("Can't generate email alert template!")
		panic(err)
	}

	for _, uid := range t.users {
		user, err := t.pool.store.GetUser(uid)

		if !user.Alert {
			continue
		}

		if err != nil {
			util.LogError(err)
			continue
		}

		if err := mailer.Send(
			util.Config.EmailSecure,
			util.Config.EmailHost,
			util.Config.EmailPort,
			util.Config.EmailUsername,
			util.Config.EmailPassword,
			util.Config.EmailSender,
			user.Email,
			fmt.Sprintf("Task '%s' failed", t.Template.Name),
			body.String(),
		); err != nil {
			util.LogError(err)
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

	body := bytes.NewBufferString("")
	author, version := t.alertInfos()

	alert := Alert{
		Name:   t.Template.Name,
		Author: author,
		Color:  t.alertColor("telegram"),
		Task: alertTask{
			ID:      strconv.Itoa(t.Task.ID),
			URL:     t.taskLink(),
			Result:  strings.ToUpper(string(t.Task.Status)),
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
	}
}

func (t *TaskRunner) sendSlackAlert() {
	if !util.Config.SlackAlert || !t.alert {
		return
	}

	if t.Template.SuppressSuccessAlerts && t.Task.Status == lib.TaskSuccessStatus {
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
			Result:  strings.ToUpper(string(t.Task.Status)),
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

	resp, err := http.Post(
		util.Config.SlackUrl,
		"application/json",
		body,
	)

	if err != nil {
		t.Log("Can't send slack alert! Error: " + err.Error())
	} else if resp.StatusCode != 200 {
		t.Log("Can't send slack alert! Response code: " + strconv.Itoa(resp.StatusCode))
	}
}

func (t *TaskRunner) alertInfos() (string, string) {
	version := ""

	if t.Task.Version != nil {
		version = *t.Task.Version
	} else if t.Task.BuildTaskID != nil {
		version = "build " + strconv.Itoa(*t.Task.BuildTaskID)
	} else {
		version = ""
	}

	author := ""

	if t.Task.UserID != nil {
		user, err := t.pool.store.GetUser(*t.Task.UserID)

		if err != nil {
			panic(err)
		}

		author = user.Name
	}

	return version, author
}

func (t *TaskRunner) alertColor(kind string) string {
	switch kind {
	case "slack":
		switch t.Task.Status {
		case lib.TaskSuccessStatus:
			return "good"
		case lib.TaskFailStatus:
			return "danger"
		case lib.TaskRunningStatus:
			return "#333CFF"
		case lib.TaskWaitingStatus:
			return "#FFFC33"
		case lib.TaskStoppingStatus:
			return "#BEBEBE"
		case lib.TaskStoppedStatus:
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

func (t *TaskRunner) sendMicrosoftTeamsAlert() {
	if !util.Config.MicrosoftTeamsAlert || !t.alert {
		return
	}

	if t.Template.SuppressSuccessAlerts && t.Task.Status == lib.TaskSuccessStatus {
		return
	}

	MicrosoftTeamsUrl := util.Config.MicrosoftTeamsUrl

	var microsoftTeamsBuffer bytes.Buffer

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

	// Instantiate an alert object
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

	tpl := template.New("MicrosoftTeams body template")

	tpl, err := tpl.Parse(microsoftTeamsTemplate)
	if err != nil {
		t.Log("Can't parse MicrosoftTeams template!")
		panic(err)
	}

	// The tpl.Execute(&microsoftTeamsBuffer, alert) line is used to apply the data from the alert struct to the template.
	// This operation fills in the placeholders in the template with the corresponding values from the alert struct
	// and writes the result to the microsoftTeamsBuffer. In essence, it generates a JSON message based on the template and the data in the alert struct.
	err = tpl.Execute(&microsoftTeamsBuffer, alert)
	if err != nil {
		t.Log("Can't generate alert template!")
		panic(err)
	}

	// test if buffer is empty
	if microsoftTeamsBuffer.Len() == 0 {
		t.Log("MicrosoftTeams buffer is empty!")
		return
	}

	t.Log("Attempting to send MicrosoftTeams alert")

	resp, err := http.Post(MicrosoftTeamsUrl, "application/json", &microsoftTeamsBuffer)

	if err != nil {
		t.Log("Can't send MicrosoftTeams alert! Error: " + err.Error())
	} else if resp.StatusCode != 200 {
		t.Log("Can't send MicrosoftTeams alert! Response code: " + strconv.Itoa(resp.StatusCode))
	}

	t.Log("MicrosoftTeams alert sent successfully")
}
