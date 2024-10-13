package runners

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ansible-semaphore/semaphore/db"

	"github.com/ansible-semaphore/semaphore/db_lib"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
	"github.com/ansible-semaphore/semaphore/services/tasks"
	"github.com/ansible-semaphore/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type ContextLogger struct {
	Context string
}

func (e *ContextLogger) Error(err error, action string, message string) {
	util.LogErrorWithFields(err, log.Fields{
		"context": e.Context,
		"action":  action,
		"error":   message,
	})
}

func (e *ContextLogger) Info(message string) {
	log.WithFields(log.Fields{
		"context": e.Context,
	}).Info(message)
}

func (e *ContextLogger) Panic(err error, action string, message string) {
	log.WithFields(log.Fields{
		"context": e.Context,
	}).Panic(message)
}

type JobPool struct {
	// logger channel used to putting log records to database.
	logger chan jobLogRecord

	// register channel used to put tasks to queue.
	register chan *job

	runningJobs map[int]*runningJob

	queue []*job

	//token *string

	processing int32
}

func (p *JobPool) existsInQueue(taskID int) bool {
	for _, j := range p.queue {
		if j.job.Task.ID == taskID {
			return true
		}
	}

	return false
}

func (p *JobPool) hasRunningJobs() bool {
	for _, j := range p.runningJobs {
		if !j.status.IsFinished() {
			return true
		}
	}

	return false
}

func (p *JobPool) Register() (err error) {
	//if util.Config.Runner.RegistrationToken == "" {
	//	return fmt.Errorf("runner registration token required")
	//}

	if util.Config.Runner.TokenFile == "" {
		return fmt.Errorf("runner token file required")
	}

	ok := p.tryRegisterRunner()

	if !ok {
		return fmt.Errorf("runner registration failed")
	}

	return
}

func (p *JobPool) Unregister() (err error) {

	if util.Config.Runner.Token == "" {
		return fmt.Errorf("runner is not registered")
	}

	client := &http.Client{}

	url := util.Config.WebHost + "/api/internal/runners"

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		err = fmt.Errorf("encountered error while unregistering runner; server returned code %d", resp.StatusCode)
		return
	}

	if util.Config.Runner.TokenFile != "" {
		err = os.Remove(util.Config.Runner.TokenFile)
	}

	return
}

func (p *JobPool) Run() {
	logger := ContextLogger{Context: "running"}

	if util.Config.Runner.Token == "" {
		logger.Panic(fmt.Errorf("no token provided"), "read input", "can not retrieve runner token")
	}

	queueTicker := time.NewTicker(5 * time.Second)
	requestTimer := time.NewTicker(1 * time.Second)
	p.runningJobs = make(map[int]*runningJob)

	defer func() {
		queueTicker.Stop()
		requestTimer.Stop()
	}()

	for {
		select {

		case <-queueTicker.C: // timer 5 seconds: get task from queue and run it
			if len(p.queue) == 0 {
				break
			}

			t := p.queue[0]
			if t.status == task_logger.TaskFailStatus {
				//delete failed TaskRunner from queue
				p.queue = p.queue[1:]
				log.Info("Task " + strconv.Itoa(t.job.Task.ID) + " dequeued (failed)")
				break
			}

			p.runningJobs[t.job.Task.ID] = &runningJob{
				job: t.job,
			}

			t.job.Logger = t.job.App.SetLogger(p.runningJobs[t.job.Task.ID])

			go func(runningJob *runningJob) {
				runningJob.SetStatus(task_logger.TaskRunningStatus)

				err := runningJob.job.Run(t.username, t.incomingVersion)

				if runningJob.status.IsFinished() {
					return
				}

				if err != nil {
					if runningJob.status == task_logger.TaskStoppingStatus {
						runningJob.SetStatus(task_logger.TaskStoppedStatus)
					} else {
						runningJob.SetStatus(task_logger.TaskFailStatus)
					}
				} else {
					runningJob.SetStatus(task_logger.TaskSuccessStatus)
				}

				logger.Info("Task " + strconv.Itoa(runningJob.job.Task.ID) + " finished (" + string(runningJob.status) + ")")
			}(p.runningJobs[t.job.Task.ID])

			p.queue = p.queue[1:]
			logger.Info("Task " + strconv.Itoa(t.job.Task.ID) + " dequeued")
			logger.Info("Task " + strconv.Itoa(t.job.Task.ID) + " started")

		case <-requestTimer.C:

			go func() {

				if !atomic.CompareAndSwapInt32(&p.processing, 0, 1) {
					return
				}

				defer atomic.StoreInt32(&p.processing, 0)

				p.sendProgress()

				if util.Config.Runner.OneOff && len(p.runningJobs) > 0 && !p.hasRunningJobs() {
					os.Exit(0)
				}

				p.checkNewJobs()
			}()

		}
	}
}

func (p *JobPool) sendProgress() {

	logger := ContextLogger{Context: "sending_progress"}

	client := &http.Client{}

	url := util.Config.WebHost + "/api/internal/runners"

	body := RunnerProgress{
		Jobs: nil,
	}

	for id, j := range p.runningJobs {

		body.Jobs = append(body.Jobs, JobProgress{
			ID:         id,
			LogRecords: j.logRecords,
			Status:     j.status,
		})

		j.logRecords = make([]LogRecord, 0)

		if j.status.IsFinished() {
			logger.Info("Task " + strconv.Itoa(id) + " removed from running list")
			delete(p.runningJobs, id)
		}
	}

	jsonBytes, err := json.Marshal(body)

	if err != nil {
		logger.Error(err, "form request body", "can not marshal json")
		return
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		logger.Error(err, "create request", "can not create request to the server")
		return
	}

	req.Header.Set("X-Runner-Token", util.Config.Runner.Token)

	resp, err := client.Do(req)
	if err != nil {

		logger.Error(err, "send request", "the server returned error "+strconv.Itoa(resp.StatusCode))
		return
	}

	defer resp.Body.Close()
}

func (p *JobPool) tryRegisterRunner() bool {

	logger := ContextLogger{Context: "registration"}

	log.Info("Registering a new runner")

	if util.Config.Runner.RegistrationToken == "" {
		logger.Error(fmt.Errorf("registration token cannot be empty"), "read input", "can not retrieve registration token")
	}

	client := &http.Client{}

	url := util.Config.WebHost + "/api/internal/runners"

	jsonBytes, err := json.Marshal(RunnerRegistration{
		RegistrationToken: util.Config.Runner.RegistrationToken,
		Webhook:           util.Config.Runner.Webhook,
		MaxParallelTasks:  util.Config.Runner.MaxParallelTasks,
	})

	if err != nil {
		logger.Error(err, "form request", "can not marshal json")
		return false
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		logger.Error(err, "create request", "can not create request to the server")
		return false
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		logger.Error(err, "send request", "the server returned error "+strconv.Itoa(resp.StatusCode))
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {

		logger.Error(err, "read response body", "can not read server's response body")
		return false
	}

	var res struct {
		Token string `json:"token"`
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		logger.Error(err, "parsing result json", "server's response has invalid format")
		return false
	}

	err = os.WriteFile(util.Config.Runner.TokenFile, []byte(res.Token), 0644)

	if err != nil {
		logger.Error(err, "store token", "can not store token to the file")
		return false
	}

	defer resp.Body.Close()

	return true
}

// checkNewJobs tries to find runner to queued jobs
func (p *JobPool) checkNewJobs() {

	logger := ContextLogger{Context: "checking new jobs"}

	if util.Config.Runner.Token == "" {
		logger.Error(fmt.Errorf("no token provided"), "read input", "can not retrieve runner token")
		return
	}

	client := &http.Client{}

	url := util.Config.WebHost + "/api/internal/runners"

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		logger.Error(err, "create request", "can not create request to the server")
		return
	}

	req.Header.Set("X-Runner-Token", util.Config.Runner.Token)

	resp, err := client.Do(req)

	if err != nil {
		logger.Error(err, "send request", "the server returned an error"+strconv.Itoa(resp.StatusCode))
		return
	}

	if resp.StatusCode >= 400 {

		logger.Error(fmt.Errorf("error status code"), "send request", "the server returned an error"+strconv.Itoa(resp.StatusCode))
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err, "read response body", "can not read server's response body")
		return
	}

	var response RunnerState
	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.Error(err, "parsing result json", "server's response has invalid format")
		return
	}

	for _, currJob := range response.CurrentJobs {
		runJob, exists := p.runningJobs[currJob.ID]

		if !exists {
			continue
		}

		if runJob.status == task_logger.TaskStoppingStatus || runJob.status == task_logger.TaskStoppedStatus {
			p.runningJobs[currJob.ID].job.Kill()
		}

		if runJob.status.IsFinished() {
			continue
		}

		switch runJob.status {
		case task_logger.TaskRunningStatus:
			if currJob.Status == task_logger.TaskStartingStatus || currJob.Status == task_logger.TaskWaitingStatus {
				continue
			}
		case task_logger.TaskStoppingStatus:
			if !currJob.Status.IsFinished() {
				continue
			}
		case task_logger.TaskConfirmed:
			if currJob.Status == task_logger.TaskWaitingConfirmation {
				continue
			}
		}

		runJob.SetStatus(currJob.Status)
	}

	if util.Config.Runner.OneOff {
		if len(p.queue) > 0 || len(p.runningJobs) > 0 {
			return
		}
	}

	for _, newJob := range response.NewJobs {
		if _, exists := p.runningJobs[newJob.Task.ID]; exists {
			continue
		}

		if p.existsInQueue(newJob.Task.ID) {
			continue
		}

		newJob.Inventory.Repository = newJob.InventoryRepository

		taskRunner := job{
			username:        newJob.Username,
			incomingVersion: newJob.IncomingVersion,

			job: &tasks.LocalJob{
				Task:        newJob.Task,
				Template:    newJob.Template,
				Inventory:   newJob.Inventory,
				Repository:  newJob.Repository,
				Environment: newJob.Environment,
				App: db_lib.CreateApp(
					newJob.Template,
					newJob.Repository,
					newJob.Inventory,
					nil),
			},
		}

		taskRunner.job.Repository.SSHKey = response.AccessKeys[taskRunner.job.Repository.SSHKeyID]

		if taskRunner.job.Inventory.SSHKeyID != nil {
			taskRunner.job.Inventory.SSHKey = response.AccessKeys[*taskRunner.job.Inventory.SSHKeyID]
		}

		if taskRunner.job.Inventory.BecomeKeyID != nil {
			taskRunner.job.Inventory.BecomeKey = response.AccessKeys[*taskRunner.job.Inventory.BecomeKeyID]
		}

		var vaults []db.TemplateVault
		if taskRunner.job.Template.Vaults != nil {
			for _, vault := range taskRunner.job.Template.Vaults {
				vault := vault
				key := response.AccessKeys[vault.VaultKeyID]
				vault.Vault = &key
				vaults = append(vaults, vault)
			}
		}
		taskRunner.job.Template.Vaults = vaults

		if taskRunner.job.Inventory.RepositoryID != nil {
			taskRunner.job.Inventory.Repository.SSHKey = response.AccessKeys[taskRunner.job.Inventory.Repository.SSHKeyID]
		}

		p.queue = append(p.queue, &taskRunner)
		logger.Info("Task " + strconv.Itoa(taskRunner.job.Task.ID) + " enqueued")
	}
}
