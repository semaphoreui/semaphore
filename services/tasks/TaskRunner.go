package tasks

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
)

type TaskRunner struct {
	task        db.Task
	template    db.Template
	inventory   db.Inventory
	repository  db.Repository
	environment db.Environment

	users     []int
	alert     bool
	alertChat *string
	//prepared  bool

	pool *TaskPool

	// job executes Ansible and returns stdout to Semaphore logs
	job AnsibleJobRunner
}

func getMD5Hash(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (t *TaskRunner) setStatus(status db.TaskStatus) {
	if t.task.Status == db.TaskStoppingStatus {
		switch status {
		case db.TaskFailStatus:
			status = db.TaskStoppedStatus
		case db.TaskStoppedStatus:
		default:
			panic("stopping TaskRunner cannot be " + status)
		}
	}

	t.task.Status = status

	t.updateStatus()

	if status == db.TaskFailStatus {
		t.sendMailAlert()
	}

	if status == db.TaskSuccessStatus || status == db.TaskFailStatus {
		t.sendTelegramAlert()
		t.sendSlackAlert()
	}
}

func (t *TaskRunner) updateStatus() {
	for _, user := range t.users {
		b, err := json.Marshal(&map[string]interface{}{
			"type":        "update",
			"start":       t.task.Start,
			"end":         t.task.End,
			"status":      t.task.Status,
			"task_id":     t.task.ID,
			"template_id": t.task.TemplateID,
			"project_id":  t.task.ProjectID,
			"version":     t.task.Version,
		})

		util.LogPanic(err)

		sockets.Message(user, b)
	}

	if err := t.pool.store.UpdateTask(t.task); err != nil {
		t.panicOnError(err, "Failed to update TaskRunner status")
	}
}

func (t *TaskRunner) fail() {
	t.setStatus(db.TaskFailStatus)
}

func (t *TaskRunner) createTaskEvent() {
	objType := db.EventTask
	desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Name + ")" + " finished - " + strings.ToUpper(string(t.task.Status))

	_, err := t.pool.store.CreateEvent(db.Event{
		UserID:      t.task.UserID,
		ProjectID:   &t.task.ProjectID,
		ObjectType:  &objType,
		ObjectID:    &t.task.ID,
		Description: &desc,
	})

	if err != nil {
		t.panicOnError(err, "Fatal error inserting an event")
	}
}

func (t *TaskRunner) run() {
	if !t.pool.store.PermanentConnection() {
		t.pool.store.Connect("run task " + strconv.Itoa(t.task.ID))
		defer t.pool.store.Close("run task " + strconv.Itoa(t.task.ID))
	}

	defer func() {
		log.Info("Stopped running TaskRunner " + strconv.Itoa(t.task.ID))
		log.Info("Release resource locker with TaskRunner " + strconv.Itoa(t.task.ID))
		t.pool.resourceLocker <- &resourceLock{lock: false, holder: t}

		now := time.Now()
		t.task.End = &now
		t.updateStatus()
		t.createTaskEvent()
	}()

	// TODO: more details
	if t.task.Status == db.TaskStoppingStatus {
		t.setStatus(db.TaskStoppedStatus)
		return
	}

	now := time.Now()
	t.task.Start = &now
	t.setStatus(db.TaskRunningStatus)

	objType := db.EventTask
	desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Name + ")" + " is running"

	_, err := t.pool.store.CreateEvent(db.Event{
		UserID:      t.task.UserID,
		ProjectID:   &t.task.ProjectID,
		ObjectType:  &objType,
		ObjectID:    &t.task.ID,
		Description: &desc,
	})

	if err != nil {
		t.Log("Fatal error inserting an event")
		panic(err)
	}

	t.Log("Started: " + strconv.Itoa(t.task.ID))
	t.Log("Run TaskRunner with template: " + t.template.Name + "\n")

	// Mark task as stopped if user stops task during preparation (before task run).
	if t.task.Status == db.TaskStoppingStatus {
		t.setStatus(db.TaskStoppedStatus)
		return
	}

	err = t.job.Run()

	if err != nil {
		t.Log("Running playbook failed: " + err.Error())
		t.fail()
		return
	}

	t.setStatus(db.TaskSuccessStatus)

	templates, err := t.pool.store.GetTemplates(t.task.ProjectID, db.TemplateFilter{
		BuildTemplateID: &t.task.TemplateID,
		AutorunOnly:     true,
	}, db.RetrieveQueryParams{})
	if err != nil {
		t.Log("Running playbook failed: " + err.Error())
		return
	}

	for _, tpl := range templates {
		_, err = t.pool.AddTask(db.Task{
			TemplateID:  tpl.ID,
			ProjectID:   tpl.ProjectID,
			BuildTaskID: &t.task.ID,
		}, nil, tpl.ProjectID)
		if err != nil {
			t.Log("Running playbook failed: " + err.Error())
			continue
		}
	}
}

func (t *TaskRunner) prepareError(err error, errMsg string) error {
	if err == db.ErrNotFound {
		t.Log(errMsg)
		return err
	}

	if err != nil {
		t.fail()
		panic(err)
	}

	return nil
}

// nolint: gocyclo
func (t *TaskRunner) populateDetails() error {
	// get template
	var err error

	t.template, err = t.pool.store.GetTemplate(t.task.ProjectID, t.task.TemplateID)
	if err != nil {
		return t.prepareError(err, "Template not found!")
	}

	// get project alert setting
	project, err := t.pool.store.GetProject(t.template.ProjectID)
	if err != nil {
		return t.prepareError(err, "Project not found!")
	}

	t.alert = project.Alert
	t.alertChat = project.AlertChat

	// get project users
	users, err := t.pool.store.GetProjectUsers(t.template.ProjectID, db.RetrieveQueryParams{})
	if err != nil {
		return t.prepareError(err, "Users not found!")
	}

	t.users = []int{}
	for _, user := range users {
		t.users = append(t.users, user.ID)
	}

	// get inventory
	t.inventory, err = t.pool.store.GetInventory(t.template.ProjectID, t.template.InventoryID)
	if err != nil {
		return t.prepareError(err, "Template Inventory not found!")
	}

	// get repository
	t.repository, err = t.pool.store.GetRepository(t.template.ProjectID, t.template.RepositoryID)

	if err != nil {
		return err
	}

	err = t.repository.SSHKey.DeserializeSecret()
	if err != nil {
		return err
	}

	// get environment
	if t.template.EnvironmentID != nil {
		t.environment, err = t.pool.store.GetEnvironment(t.template.ProjectID, *t.template.EnvironmentID)
		if err != nil {
			return err
		}
	}

	if t.task.Environment != "" {
		environment := make(map[string]interface{})
		if t.environment.JSON != "" {
			err = json.Unmarshal([]byte(t.task.Environment), &environment)
			if err != nil {
				return err
			}
		}

		taskEnvironment := make(map[string]interface{})
		err = json.Unmarshal([]byte(t.environment.JSON), &taskEnvironment)
		if err != nil {
			return err
		}

		for k, v := range taskEnvironment {
			environment[k] = v
		}

		var ev []byte
		ev, err = json.Marshal(environment)
		if err != nil {
			return err
		}

		t.environment.JSON = string(ev)
	}

	return nil
}

func hasRequirementsChanges(requirementsFilePath string, requirementsHashFilePath string) bool {
	oldFileMD5HashBytes, err := ioutil.ReadFile(requirementsHashFilePath)
	if err != nil {
		return true
	}

	newFileMD5Hash, err := getMD5Hash(requirementsFilePath)
	if err != nil {
		return true
	}

	return string(oldFileMD5HashBytes) != newFileMD5Hash
}

func writeMD5Hash(requirementsFile string, requirementsHashFile string) error {
	newFileMD5Hash, err := getMD5Hash(requirementsFile)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(requirementsHashFile, []byte(newFileMD5Hash), 0644)
}

// checkTmpDir checks to see if the temporary directory exists
// and if it does not attempts to create it
func checkTmpDir(path string) error {
	var err error
	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, 0700)
		}
	}
	return err
}
