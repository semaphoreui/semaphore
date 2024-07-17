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

	"github.com/ansible-semaphore/semaphore/db_lib"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
	"github.com/ansible-semaphore/semaphore/services/tasks"
	"github.com/ansible-semaphore/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type JobPool struct {
	// logger channel used to putting log records to database.
	logger chan jobLogRecord

	// register channel used to put tasks to queue.
	register chan *job

	runningJobs map[int]*runningJob

	queue []*job

	config *util.RunnerConfig

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

func (p *JobPool) Unregister() (err error) {

	config, err := util.LoadRunnerSettings(util.Config.Runner.ConfigFile)

	if err != nil {
		return
	}

	if config.Token == "" {
		return fmt.Errorf("runner is not registered")
	}

	client := &http.Client{}

	url := util.Config.Runner.ApiURL + "/runners"

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

	err = os.Remove(util.Config.Runner.ConfigFile)
	if err != nil {
		return
	}

	return
}

func (p *JobPool) Run() {
	queueTicker := time.NewTicker(5 * time.Second)
	requestTimer := time.NewTicker(1 * time.Second)
	p.runningJobs = make(map[int]*runningJob)

	defer func() {
		queueTicker.Stop()
		requestTimer.Stop()
	}()

	for {

		if p.tryRegisterRunner() {

			log.Info("The runner has been started")

			break
		}

		time.Sleep(5_000_000_000)
	}

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

				log.Info("Task " + strconv.Itoa(runningJob.job.Task.ID) + " finished (" + string(runningJob.status) + ")")
			}(p.runningJobs[t.job.Task.ID])

			p.queue = p.queue[1:]
			log.Info("Task " + strconv.Itoa(t.job.Task.ID) + " dequeued")
			log.Info("Task " + strconv.Itoa(t.job.Task.ID) + " started")

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

	client := &http.Client{}

	url := util.Config.Runner.ApiURL + "/runners/" + strconv.Itoa(p.config.RunnerID)

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
			log.Info("Task " + strconv.Itoa(id) + " removed from running list")
			delete(p.runningJobs, id)
		}
	}

	jsonBytes, err := json.Marshal(body)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("X-API-Token", p.config.Token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}

	defer resp.Body.Close()
}

func (p *JobPool) tryRegisterRunner() bool {
	if p.config != nil {
		return true
	}

	log.Info("Attempting to register on the server")

	config, err := util.LoadRunnerSettings(util.Config.Runner.ConfigFile)

	if err != nil {
		panic(err)
	}

	if config.Token != "" {
		p.config = &config
		return true
	}

	// Can not restore runner configuration. Register new runner on the server.

	if util.Config.Runner.RegistrationToken == "" {
		panic("registration token cannot be empty")
	}

	log.Info("Registering a new runner")

	client := &http.Client{}

	url := util.Config.Runner.ApiURL + "/runners"

	jsonBytes, err := json.Marshal(RunnerRegistration{
		RegistrationToken: util.Config.Runner.RegistrationToken,
		Webhook:           util.Config.Runner.Webhook,
		MaxParallelTasks:  util.Config.Runner.MaxParallelTasks,
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Error("Error creating request:", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		log.Error("Error making request:", err)
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return false
	}

	err = json.Unmarshal(body, &config)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return false
	}

	configBytes, err := json.Marshal(config)

	if err != nil {
		panic("cannot save runner config")
	}

	err = os.WriteFile(util.Config.Runner.ConfigFile, configBytes, 0644)

	p.config = &config

	defer resp.Body.Close()

	return true
}

// checkNewJobs tries to find runner to queued jobs
func (p *JobPool) checkNewJobs() {

	client := &http.Client{}

	url := util.Config.Runner.ApiURL + "/runners/" + strconv.Itoa(p.config.RunnerID)

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("X-API-Token", p.config.Token)

	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}

	if resp.StatusCode >= 400 {
		log.Error("Encountered error while checking for new jobs; server returned code ", resp.StatusCode)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Encountered error while checking for new jobs; unable to read response body:", err)
		return
	}

	var response RunnerState
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Error("Checking new jobs, parsing JSON error:", err)
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

		if taskRunner.job.Template.VaultKeyID != nil {
			taskRunner.job.Template.VaultKey = response.AccessKeys[*taskRunner.job.Template.VaultKeyID]
		}

		if taskRunner.job.Inventory.RepositoryID != nil {
			taskRunner.job.Inventory.Repository.SSHKey = response.AccessKeys[taskRunner.job.Inventory.Repository.SSHKeyID]
		}

		p.queue = append(p.queue, &taskRunner)
		log.Info("Task " + strconv.Itoa(taskRunner.job.Task.ID) + " enqueued")
	}
}
