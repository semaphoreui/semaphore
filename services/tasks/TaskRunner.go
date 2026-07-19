package tasks

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/jwt"
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
	// Async reports whether the job completes asynchronously after Run returns.
	// Remote runner jobs return true: Run only dispatches the task to a runner,
	// and completion is reported back via the runner API and finalized by
	// TaskPool.FinalizeRemoteTask. When true, TaskRunner.run must NOT finalize
	// the task when Run returns. Local jobs return false (synchronous).
	Async() bool
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

	Username        string
	IncomingVersion *string

	statusListeners []task_logger.StatusListener
	logListeners    []task_logger.LogListener

	// Alias uses if task require an alias for run.
	// For example, terraform task require an alias for run.
	Alias string

	logWG sync.WaitGroup

	// dispatching is true while this process owns a live goroutine that is
	// dispatching/running the task (set in runTask). A TaskRunner restored from
	// Redis after a node restart (getOrHydrate / RedisTaskStateStore.Start) is an
	// inert stub with no goroutine and leaves this false. The runner-task
	// reconciler uses it to tell a task it is actively dispatching from a stale
	// "starting" task whose dispatch goroutine died with a previous process.
	dispatching atomic.Bool
}

// isDispatching reports whether this process has a live goroutine
// dispatching/running the task. See the dispatching field.
func (t *TaskRunner) isDispatching() bool {
	return t.dispatching.Load()
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

	t.pool.state.UpdateRuntimeFields(t)
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

	if err := t.pool.logWriteService.WriteTaskLog(pro_interfaces.TaskLogRecord{
		ProjectID:    t.Task.ProjectID,
		TemplateID:   t.Template.ID,
		TemplateName: t.Template.Name,
		TaskID:       t.Task.ID,
		UserID:       t.Task.UserID,
		Description:  &desc,
		Username:     t.Username,
		RunnerID:     t.Task.RunnerID,
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

	// requeued indicates task should go back to waiting state (e.g., all runners busy)
	requeued := false

	// handedOff indicates the task was dispatched to a remote runner and will be
	// finalized asynchronously when the runner reports completion (see
	// TaskPool.FinalizeRemoteTask). In that case run() must not finalize it here.
	handedOff := false

	defer func() {
		if requeued {
			// Task is being re-queued, don't mark as finished
			log.Info("Task " + strconv.Itoa(t.Task.ID) + " re-queued (waiting for available runner)")
			t.pool.queueEvents <- PoolEvent{EventTypeRequeued, t}
			return
		}

		if handedOff {
			// Task is now running on a remote runner. Completion is driven by the
			// runner's status report, not by this goroutine, so do not finalize.
			l := log.WithField("task_id", t.Task.ID).WithField("status", t.Task.Status)

			if t.Task.RunnerID != nil {
				l = log.WithField("runner_id", *t.Task.RunnerID)
			}

			l.Info("Task dispatched to runner; awaiting remote completion")
			return
		}

		log.WithFields(log.Fields{
			"task_id": t.Task.ID,
		}).Info("Stopped running task " + t.Template.Name)

		t.finishRun()
	}()

	// Mark task as stopped if user stopped task during preparation (before task run).
	if t.Task.Status == task_logger.TaskStoppingStatus {
		t.SetStatus(task_logger.TaskStoppedStatus)
		return
	}

	t.SetStatus(task_logger.TaskStartingStatus)
	t.createTaskEvent()

	t.Log("Started task #" + strconv.Itoa(t.Task.ID) + " of template '" + t.Template.Name + "'\n")

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

	// For locally-executed tasks, mint a JWT and pass it to the LocalJob so it
	// can be exposed to the playbook as SEMAPHORE_JWT. Remote runners receive
	// the JWT inside the JobData payload returned by the API.
	if localJob, ok := t.job.(*LocalExecutor); ok {
		if t.pool.signer != nil && t.Template.JWTParams != nil && t.Template.JWTParams.Enabled {
			ttl, terr := t.Template.JWTParams.ParsedTTL()
			if terr != nil {
				log.WithError(terr).WithFields(log.Fields{
					"task_id":     t.Task.ID,
					"template_id": t.Template.ID,
					"context":     "jwt",
				}).Error("invalid template jwt_params.ttl; skipping token issuance")
			} else {
				token, jerr := t.pool.signer.Sign(jwt.TaskInfo{
					TaskID:     t.Task.ID,
					ProjectID:  t.Task.ProjectID,
					TemplateID: t.Template.ID,
					UserID:     t.Task.UserID,
					Audience:   jwt.Audience(t.Template.JWTParams.Audience),
					TTL:        ttl,
				})
				if jerr != nil {
					log.WithError(jerr).WithFields(log.Fields{
						"task_id": t.Task.ID,
						"context": "jwt",
					}).Error("failed to sign task JWT")
				} else {
					localJob.JWT = token
				}
			}
		}
	}

	err = t.job.Run(username, incomingVersion, t.Alias)

	if err != nil {
		if errors.Is(err, ErrAllRunnersBusy) {
			// No runners available right now, put task back in waiting state
			t.SetStatus(task_logger.TaskWaitingStatus)
			t.pool.state.Enqueue(t)
			requeued = true
			return
		}

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

	// Remote jobs only dispatch the task to a runner; their completion is
	// reported asynchronously via the runner API and finalized there. Hand off
	// and let the deferred cleanup skip finalization.
	if t.job.Async() {
		handedOff = true
		return
	}

	if t.Task.Status == task_logger.TaskRunningStatus {
		t.SetStatus(task_logger.TaskSuccessStatus)
	}

	t.startAutorunTasks()
}

// finishRun records the end of a task run, persists it, and notifies the pool
// to release the task's resources (EventTypeFinished -> onTaskStop). It is used
// by the synchronous local path and by FinalizeRemoteTask for remote tasks.
func (t *TaskRunner) finishRun() {
	now := tz.Now()
	t.Task.End = &now
	t.saveStatus()
	t.createTaskEvent()
	t.pool.queueEvents <- PoolEvent{EventTypeFinished, t}

	// Notify the workflow service that this task finished so it can progress the
	// run (launch downstream nodes, create the next approval, or finalize the
	// run). This is the single completion point shared by the local (deferred
	// run()) and remote (FinalizeRemoteTask) paths and fires for every terminal
	// status — success and failure drive different edges, so progression must
	// not be limited to success or to tasks that have autorun children. It is a
	// no-op for non-workflow tasks. Done after the EventTypeFinished event so the
	// finished task's pool slot is released before any downstream node is queued.
	if err := t.pool.HandleWorkflowTaskCompletion(t.Task); err != nil {
		t.Log("Workflow progression failed: " + err.Error())
	}
}

// startAutorunTasks queues the autorun child templates of a successfully
// finished build task. It is a no-op unless the task succeeded.
func (t *TaskRunner) startAutorunTasks() {
	if t.Task.Status != task_logger.TaskSuccessStatus {
		return
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

	if t.Environment.JSON != "" {
		err = json.Unmarshal([]byte(t.Environment.JSON), &tplEnvironment)
	}

	if err != nil {
		return
	}

	taskEnvironment := make(map[string]any)
	if t.Task.Environment != "" {
		err = json.Unmarshal([]byte(t.Task.Environment), &taskEnvironment)
	}

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

	// load and merge all configured environments
	err = t.loadEnvironments()
	if err != nil {
		return err
	}

	err = t.populateTaskEnvironment()

	return err
}

// loadEnvironments loads all Variable Groups configured on the template
// (Template.EnvironmentIDs) and merges their JSON, ENV vars, and secrets
// into t.Environment. Later entries override earlier ones on key conflicts.
// If the template allows overriding environments and the task specifies
// environment_ids, those are used instead of the template's list.
func (t *TaskRunner) loadEnvironments() error {
	envIDs := t.Template.EnvironmentIDs

	// Use task-level override when the template allows it
	if t.Template.AllowOverrideEnvInTask && len(t.Task.EnvironmentIDs) > 0 {
		envIDs = []int(t.Task.EnvironmentIDs)
	}

	if len(envIDs) == 0 {
		return nil
	}

	seen := make(map[int]bool)

	mergedJSON := make(map[string]any)
	mergedENV := make(map[string]string)
	var mergedSecrets []db.EnvironmentSecret
	secretIndex := make(map[string]int)

	var lastEnv db.Environment

	for _, envID := range envIDs {
		if seen[envID] {
			continue
		}
		seen[envID] = true

		env, err := t.pool.store.GetEnvironment(t.Template.ProjectID, envID)
		if err != nil {
			return err
		}

		err = t.pool.encryptionService.FillEnvironmentSecrets(&env, true)
		if err != nil {
			return err
		}

		if env.JSON != "" {
			partial := make(map[string]any)
			if err := json.Unmarshal([]byte(env.JSON), &partial); err != nil {
				return err
			}
			for k, v := range partial {
				mergedJSON[k] = v
			}
		}

		if env.ENV != nil && *env.ENV != "" {
			partial := make(map[string]string)
			if err := json.Unmarshal([]byte(*env.ENV), &partial); err != nil {
				return err
			}
			for k, v := range partial {
				mergedENV[k] = v
			}
		}

		for _, s := range env.Secrets {
			key := string(s.Type) + ":" + s.Name
			if idx, ok := secretIndex[key]; ok {
				mergedSecrets[idx] = s
			} else {
				mergedSecrets = append(mergedSecrets, s)
				secretIndex[key] = len(mergedSecrets) - 1
			}
		}

		lastEnv = env
	}

	t.Environment = lastEnv

	if len(mergedJSON) > 0 {
		b, err := json.Marshal(mergedJSON)
		if err != nil {
			return err
		}
		t.Environment.JSON = string(b)
	} else {
		t.Environment.JSON = ""
	}

	if len(mergedENV) > 0 {
		b, err := json.Marshal(mergedENV)
		if err != nil {
			return err
		}
		s := string(b)
		t.Environment.ENV = &s
	} else {
		t.Environment.ENV = nil
	}

	t.Environment.Secrets = mergedSecrets

	return nil
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
