package tasks

import (
	"bytes"
	"encoding/json"
	"github.com/ansible-semaphore/semaphore/util"
	"net/http"
	"strconv"
	"time"
)

func (t *task) sendAuditLog() {
	if !util.Config.AuditLog {
		return
	}
	url := util.Config.AuditLogURL
	method := "POST"
	payload, err := json.Marshal(map[string]interface{}{
		"project_id":  t.projectID,
		"template_id": t.task.TemplateID,
		"task_id":     t.task.ID,
		"playbook":    t.task.Playbook,
		"environment": t.task.Environment,
		"start":       t.task.Start,
		"end":         t.task.End,
		"status":      t.task.Status,
		"task_url":    util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID) + "/history/?t=" + strconv.Itoa(t.task.ID),
	})
	requestBody := bytes.NewBuffer(payload)
	if err != nil {
		util.LogError(err)
		return
	}
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(method, url, requestBody)

	if err != nil {
		util.LogError(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	_, err = client.Do(req)
	if err != nil {
		util.LogError(err)
		return
	}
}
