package tasks

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/tasks/hooks"

	"github.com/semaphoreui/semaphore/api/sockets"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type Job interface {
	Run(username string, incomingVersion *string, alias string) error
	Kill()
	IsKilled() bool
}

type TaskRunner struct {
	Task        db.Task
	Template    db.Template
	Inventory   db.Inventory
	Repository  db.Repository
	Environment db.Environment

	currentStage  *db.TaskStage
	currentOutput *db.TaskOutput
	currentState  any

	users        []int
	alert        bool
	alertChat    *string
	pool         *TaskPool
	keyInstaller db_lib.AccessKeyInstaller

	// job executes Ansible and returns stdout to Semaphore logs
	job Job

	RunnerID        int
	Username        string
	IncomingVersion *string

	statusListeners []task_logger.StatusListener
	logListeners    []task_logger.LogListener

	// Alias uses if task require an alias for run.
	// For example, terraform task require an alias for run.
	Alias string

	logWG sync.WaitGroup
}

func NewTaskRunner(
	newTask db.Task,
	p *TaskPool,
	username string,
	keyInstaller db_lib.AccessKeyInstaller,
) *TaskRunner {
	return &TaskRunner{
		Task:         newTask,
		pool:         p,
		Username:     username,
		keyInstaller: keyInstaller,
	}
}

func (t *TaskRunner) AddStatusListener(l task_logger.StatusListener) {
	t.statusListeners = append(t.statusListeners, l)
}

func (t *TaskRunner) AddLogListener(l task_logger.LogListener) {
	t.logListeners = append(t.logListeners, l)
}

func (t *TaskRunner) saveStatus() {
	for _, user := range t.users {
		b, err := json.Marshal(&map[string]any{
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
	// persist runtime fields in HA store
	if t.pool != nil && t.pool.state != nil {
		t.pool.state.UpdateRuntimeFields(t)
	}
}

func (t *TaskRunner) kill() {
	t.job.Kill()
}

func (t *TaskRunner) createTaskEvent() {

	desc := "Task ID " + strconv.Itoa(t.Task.ID) + " (" + t.Template.Name + ")"

	if t.Task.Status.IsFinished() {
		desc += " finished with status " + strings.ToUpper(string(t.Task.Status))

		hook := hooks.GetHook(t.Template.App)
		if hook != nil {
			go hook.End(t.pool.store, t.Task.ProjectID, t.Task.ID)
		}
	} else {
		desc += " " + strings.ToUpper(string(t.Task.Status))
	}

	objType := db.EventTask
	event := db.Event{
		UserID:      t.Task.UserID,
		ProjectID:   &t.Task.ProjectID,
		ObjectType:  &objType,
		ObjectID:    &t.Task.ID,
		Description: &desc,
	}

	var runnerID *int
	if t.RunnerID > 0 {
		runnerID = &t.RunnerID
	}

	if err := t.pool.logWriteService.WriteTaskLog(pro_interfaces.TaskLogRecord{
		ProjectID:    t.Task.ProjectID,
		TemplateID:   t.Template.ID,
		TemplateName: t.Template.Name,
		TaskID:       t.Task.ID,
		UserID:       t.Task.UserID,
		Description:  &desc,
		Username:     t.Username,
		RunnerID:     runnerID,
		Status:       t.Task.Status,
	}); err != nil {
		log.Error(err)
	}

	_, err := t.pool.store.CreateEvent(event)

	if err != nil {
		msg := "Fatal error inserting an event"
		t.Log(msg)
		log.WithError(err).Error(msg)
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

		now := tz.Now()
		t.Task.End = &now
		t.saveStatus()
		t.createTaskEvent()
		t.pool.queueEvents <- PoolEvent{EventTypeFinished, t}
	}()

	// Mark task as stopped if user stopped task during preparation (before task run).
	if t.Task.Status == task_logger.TaskStoppingStatus {
		t.SetStatus(task_logger.TaskStoppedStatus)
		return
	}

	t.SetStatus(task_logger.TaskStartingStatus)
	t.createTaskEvent()

	t.Log("Started: " + strconv.Itoa(t.Task.ID))
	t.Log("Run TaskRunner with template: " + t.Template.Name + "\n")

	var err error
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

	err = t.job.Run(username, incomingVersion, t.Alias)

	if err != nil {
		if t.job.IsKilled() {
			t.SetStatus(task_logger.TaskStoppedStatus)
		} else {
			log.WithError(err).WithFields(log.Fields{
				"task_id":     t.Task.ID,
				"context":     "task_runner",
				"task_status": t.Task.Status,
			}).Warn("Failed to run task")
			t.Log("Failed to run task: " + err.Error())
			t.SetStatus(task_logger.TaskFailStatus)
		}
		return
	}

	if t.Task.Status == task_logger.TaskRunningStatus {
		t.SetStatus(task_logger.TaskSuccessStatus)
	}

	tpls, err := t.pool.store.GetTemplates(t.Task.ProjectID, db.TemplateFilter{
		BuildTemplateID: &t.Task.TemplateID,
		AutorunOnly:     true,
	}, db.RetrieveQueryParams{})

	if err != nil {
		t.Log("Running app failed: " + err.Error())
		return
	}

	for _, tpl := range tpls {
		task := db.Task{
			TemplateID:  tpl.ID,
			ProjectID:   tpl.ProjectID,
			BuildTaskID: &t.Task.ID,
		}
		_, err = t.pool.AddTask(
			task,
			nil,
			"",
			tpl.ProjectID,
			tpl.App.NeedTaskAlias(),
		)
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

func (t *TaskRunner) populateTaskEnvironment() (err error) {

	if t.Task.Environment == "" {
		return

	}

	tplEnvironment := make(map[string]any)
	err = json.Unmarshal([]byte(t.Environment.JSON), &tplEnvironment)
	if err != nil {
		return
	}

	taskEnvironment := make(map[string]any)
	err = json.Unmarshal([]byte(t.Task.Environment), &taskEnvironment)
	if err != nil {
		return
	}

	for k, v := range taskEnvironment {
		tplEnvironment[k] = v
	}

	var ev []byte
	ev, err = json.Marshal(tplEnvironment)
	if err != nil {
		return err
	}

	t.Environment.JSON = string(ev)

	return
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
	canOverrideInventory, err := t.Template.CanOverrideInventory()
	if err != nil {
		return err
	}

	if canOverrideInventory && t.Task.InventoryID != nil {
		t.Inventory, err = t.pool.inventoryService.GetInventory(t.Template.ProjectID, *t.Task.InventoryID)
		if err != nil {
			if t.Template.InventoryID != nil {
				t.Inventory, err = t.pool.inventoryService.GetInventory(t.Template.ProjectID, *t.Template.InventoryID)
				if err != nil {
					return t.prepareError(err, "Template Inventory not found!")
				}
			}
		}
	} else {
		if t.Template.InventoryID != nil {
			t.Inventory, err = t.pool.inventoryService.GetInventory(t.Template.ProjectID, *t.Template.InventoryID)
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

	if err = t.pool.encryptionService.DeserializeSecret(&t.Repository.SSHKey); err != nil {
		return err
	}

	// get environment
	if t.Template.EnvironmentID != nil {
		t.Environment, err = t.pool.store.GetEnvironment(t.Template.ProjectID, *t.Template.EnvironmentID)
		if err != nil {
			return err
		}

		err = t.pool.encryptionService.FillEnvironmentSecrets(&t.Environment, true)
		if err != nil {
			return err
		}
	}

	err = t.populateTaskEnvironment()

	return err
}

// checkTmpDir checks to see if the temporary directory exists
// and if it does not attempts to create it
func checkTmpDir(path string) error {
	var err error
	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, 0755)
		}
	}
	return err
}
