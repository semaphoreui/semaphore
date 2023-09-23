package tasks

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/lib"
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

type Job interface {
	Run(username string, incomingVersion *string) error
	Kill()
}

type TaskRunner struct {
	Task        db.Task
	Template    db.Template
	Inventory   db.Inventory
	Repository  db.Repository
	Environment db.Environment

	users     []int
	alert     bool
	alertChat *string
	pool      *TaskPool

	// job executes Ansible and returns stdout to Semaphore logs
	job Job

	RunnerID        int
	Username        string
	IncomingVersion *string
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

func (t *TaskRunner) SetStatus(status lib.TaskStatus) {
	if status == t.Task.Status {
		return
	}

	switch t.Task.Status { // check old status
	case lib.TaskRunningStatus:
		if status == lib.TaskWaitingStatus {
			//panic("running TaskRunner cannot be " + status)
			return
		}
		break
	case lib.TaskStoppingStatus:
		if status == lib.TaskWaitingStatus || status == lib.TaskRunningStatus {
			//panic("stopping TaskRunner cannot be " + status)
			return
		}
		break
	case lib.TaskSuccessStatus:
	case lib.TaskFailStatus:
	case lib.TaskStoppedStatus:
		//panic("stopped TaskRunner cannot be " + status)
		return
	}

	t.Task.Status = status

	if status == lib.TaskRunningStatus {
		now := time.Now()
		t.Task.Start = &now
	}

	t.saveStatus()

	if status == lib.TaskFailStatus {
		t.sendMailAlert()
	}

	if status == lib.TaskSuccessStatus || status == lib.TaskFailStatus {
		t.sendTelegramAlert()
		t.sendSlackAlert()
	}
}

func (t *TaskRunner) saveStatus() {
	for _, user := range t.users {
		b, err := json.Marshal(&map[string]interface{}{
			"type":        "update",
			"start":       t.Task.Start,
			"end":         t.Task.End,
			"status":      t.Task.Status,
			"task_id":     t.Task.ID,
			"template_id": t.Task.TemplateID,
			"project_id":  t.Task.ProjectID,
			"version":     t.Task.Version,
		})

		util.LogPanic(err)

		sockets.Message(user, b)
	}

	if err := t.pool.store.UpdateTask(t.Task); err != nil {
		t.panicOnError(err, "Failed to update TaskRunner status")
	}
}

func (t *TaskRunner) kill() {
	t.job.Kill()
}

func (t *TaskRunner) createTaskEvent() {
	objType := db.EventTask
	desc := "Task ID " + strconv.Itoa(t.Task.ID) + " (" + t.Template.Name + ")" + " finished - " + strings.ToUpper(string(t.Task.Status))

	_, err := t.pool.store.CreateEvent(db.Event{
		UserID:      t.Task.UserID,
		ProjectID:   &t.Task.ProjectID,
		ObjectType:  &objType,
		ObjectID:    &t.Task.ID,
		Description: &desc,
	})

	if err != nil {
		t.panicOnError(err, "Fatal error inserting an event")
	}
}

func (t *TaskRunner) run() {
	if !t.pool.store.PermanentConnection() {
		t.pool.store.Connect("run task " + strconv.Itoa(t.Task.ID))
		defer t.pool.store.Close("run task " + strconv.Itoa(t.Task.ID))
	}

	defer func() {
		log.Info("Stopped running TaskRunner " + strconv.Itoa(t.Task.ID))
		log.Info("Release resource locker with TaskRunner " + strconv.Itoa(t.Task.ID))
		t.pool.resourceLocker <- &resourceLock{lock: false, holder: t}

		now := time.Now()
		t.Task.End = &now
		t.saveStatus()
		t.createTaskEvent()
	}()

	// Mark task as stopped if user stopped task during preparation (before task run).
	if t.Task.Status == lib.TaskStoppingStatus {
		t.SetStatus(lib.TaskStoppedStatus)
		return
	}

	t.SetStatus(lib.TaskStartingStatus)

	objType := db.EventTask
	desc := "Task ID " + strconv.Itoa(t.Task.ID) + " (" + t.Template.Name + ")" + " is running"

	_, err := t.pool.store.CreateEvent(db.Event{
		UserID:      t.Task.UserID,
		ProjectID:   &t.Task.ProjectID,
		ObjectType:  &objType,
		ObjectID:    &t.Task.ID,
		Description: &desc,
	})

	if err != nil {
		t.Log("Fatal error inserting an event")
		panic(err)
	}

	t.Log("Started: " + strconv.Itoa(t.Task.ID))
	t.Log("Run TaskRunner with template: " + t.Template.Name + "\n")

	var username string
	var incomingVersion *string

	if t.Task.UserID != nil {
		var user db.User
		user, err = t.pool.store.GetUser(*t.Task.UserID)
		if err == nil {
			username = user.Username
		}
	}

	if t.Template.Type != db.TemplateTask {
		incomingVersion = t.Task.GetIncomingVersion(t.pool.store)

	}

	err = t.job.Run(username, incomingVersion)

	if err != nil {
		t.Log("Running playbook failed: " + err.Error())
		t.SetStatus(lib.TaskFailStatus)
		return
	}

	if t.Task.Status == lib.TaskRunningStatus {
		t.SetStatus(lib.TaskSuccessStatus)
	}

	templates, err := t.pool.store.GetTemplates(t.Task.ProjectID, db.TemplateFilter{
		BuildTemplateID: &t.Task.TemplateID,
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
			BuildTaskID: &t.Task.ID,
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
		t.SetStatus(lib.TaskFailStatus)
		panic(err)
	}

	return nil
}

// nolint: gocyclo
func (t *TaskRunner) populateDetails() error {
	// get template
	var err error

	t.Template, err = t.pool.store.GetTemplate(t.Task.ProjectID, t.Task.TemplateID)
	if err != nil {
		return t.prepareError(err, "Template not found!")
	}

	// get project alert setting
	project, err := t.pool.store.GetProject(t.Template.ProjectID)
	if err != nil {
		return t.prepareError(err, "Project not found!")
	}

	t.alert = project.Alert
	t.alertChat = project.AlertChat

	// get project users
	users, err := t.pool.store.GetProjectUsers(t.Template.ProjectID, db.RetrieveQueryParams{})
	if err != nil {
		return t.prepareError(err, "Users not found!")
	}

	t.users = []int{}
	for _, user := range users {
		t.users = append(t.users, user.ID)
	}

	// get inventory
	t.Inventory, err = t.pool.store.GetInventory(t.Template.ProjectID, t.Template.InventoryID)
	if err != nil {
		return t.prepareError(err, "Template Inventory not found!")
	}

	// get repository
	t.Repository, err = t.pool.store.GetRepository(t.Template.ProjectID, t.Template.RepositoryID)

	if err != nil {
		return err
	}

	err = t.Repository.SSHKey.DeserializeSecret()
	if err != nil {
		return err
	}

	// get environment
	if t.Template.EnvironmentID != nil {
		t.Environment, err = t.pool.store.GetEnvironment(t.Template.ProjectID, *t.Template.EnvironmentID)
		if err != nil {
			return err
		}
	}

	if t.Task.Environment != "" {
		environment := make(map[string]interface{})
		if t.Environment.JSON != "" {
			err = json.Unmarshal([]byte(t.Task.Environment), &environment)
			if err != nil {
				return err
			}
		}

		taskEnvironment := make(map[string]interface{})
		err = json.Unmarshal([]byte(t.Environment.JSON), &taskEnvironment)
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

		t.Environment.JSON = string(ev)
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
