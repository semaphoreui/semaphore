package tasks

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
	"github.com/ansible-semaphore/semaphore/util"
	log "github.com/sirupsen/logrus"
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

	statusListeners []task_logger.StatusListener
	logListeners    []task_logger.LogListener
}

func (t *TaskRunner) AddStatusListener(l task_logger.StatusListener) {
	t.statusListeners = append(t.statusListeners, l)
}

func (t *TaskRunner) AddLogListener(l task_logger.LogListener) {
	t.logListeners = append(t.logListeners, l)
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
	if t.Task.Status == task_logger.TaskStoppingStatus {
		t.SetStatus(task_logger.TaskStoppedStatus)
		return
	}

	t.SetStatus(task_logger.TaskStartingStatus)

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
		t.Log("Running app failed: " + err.Error())
		t.SetStatus(task_logger.TaskFailStatus)
		return
	}

	if t.Task.Status == task_logger.TaskRunningStatus {
		t.SetStatus(task_logger.TaskSuccessStatus)
	}

	templates, err := t.pool.store.GetTemplates(t.Task.ProjectID, db.TemplateFilter{
		BuildTemplateID: &t.Task.TemplateID,
		AutorunOnly:     true,
	}, db.RetrieveQueryParams{})
	if err != nil {
		t.Log("Running app failed: " + err.Error())
		return
	}

	for _, tpl := range templates {
		_, err = t.pool.AddTask(db.Task{
			TemplateID:  tpl.ID,
			ProjectID:   tpl.ProjectID,
			BuildTaskID: &t.Task.ID,
		}, nil, tpl.ProjectID)
		if err != nil {
			t.Log("Running app failed: " + err.Error())
			continue
		}
	}
}

func (t *TaskRunner) prepareError(err error, errMsg string) error {
	if errors.Is(err, db.ErrNotFound) {
		t.Log(errMsg)
		return err
	}

	if err != nil {
		t.SetStatus(task_logger.TaskFailStatus)
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
	projectUsers, err := t.pool.store.GetProjectUsers(t.Template.ProjectID, db.RetrieveQueryParams{})
	if err != nil {
		return t.prepareError(err, "Users not found!")
	}

	users := make(map[int]bool)

	for _, user := range projectUsers {
		users[user.ID] = true
	}

	admins, err := t.pool.store.GetAllAdmins()
	if err != nil {
		return err
	}

	for _, admin := range admins {
		users[admin.ID] = true
	}

	t.users = []int{}
	for userID := range users {
		t.users = append(t.users, userID)
	}

	// get inventory
	if t.Task.InventoryID != nil {
		t.Inventory, err = t.pool.store.GetInventory(t.Template.ProjectID, *t.Task.InventoryID)
		if err != nil {
			if t.Template.InventoryID != nil {
				t.Inventory, err = t.pool.store.GetInventory(t.Template.ProjectID, *t.Template.InventoryID)
				if err != nil {
					return t.prepareError(err, "Template Inventory not found!")
				}
			}
		}
	} else {
		if t.Template.InventoryID != nil {
			t.Inventory, err = t.pool.store.GetInventory(t.Template.ProjectID, *t.Template.InventoryID)
			if err != nil {
				return t.prepareError(err, "Template Inventory not found!")
			}
		}
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

		if err = db.FillEnvironmentSecrets(t.pool.store, &t.Environment, true); err != nil {
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
