//
// Runner's job pool. NOT SERVER!!!
// Runner gets jobs from the server and put them to this pool.
//

package runners

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db_lib"
	"github.com/ansible-semaphore/semaphore/lib"
	"github.com/ansible-semaphore/semaphore/services/tasks"
	"github.com/ansible-semaphore/semaphore/util"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync/atomic"
	"time"
)

type jobLogRecord struct {
	taskID int
	record LogRecord
}

type resourceLock struct {
	lock   bool
	holder *job
}

// job presents current job on semaphore server.
type job struct {
	username        string
	incomingVersion *string

	// job presents remote or local job information
	job             *tasks.LocalJob
	status          lib.TaskStatus
	args            []string
	environmentVars []string
}

type RunnerConfig struct {
	RunnerID int    `json:"runner_id"`
	Token    string `json:"token"`
}

type JobData struct {
	Username        string
	IncomingVersion *string
	Task            db.Task        `json:"task" binding:"required"`
	Template        db.Template    `json:"template" binding:"required"`
	Inventory       db.Inventory   `json:"inventory" binding:"required"`
	Repository      db.Repository  `json:"repository" binding:"required"`
	Environment     db.Environment `json:"environment" binding:"required"`
}

type RunnerState struct {
	CurrentJobs []JobState
	NewJobs     []JobData            `json:"new_jobs" binding:"required"`
	AccessKeys  map[int]db.AccessKey `json:"access_keys" binding:"required"`
}

type JobState struct {
	ID     int            `json:"id" binding:"required"`
	Status lib.TaskStatus `json:"status" binding:"required"`
}

type LogRecord struct {
	Time    time.Time `json:"time" binding:"required"`
	Message string    `json:"message" binding:"required"`
}

type RunnerProgress struct {
	Jobs []JobProgress
}

type JobProgress struct {
	ID         int
	Status     lib.TaskStatus
	LogRecords []LogRecord
}

type runningJob struct {
	status     lib.TaskStatus
	logRecords []LogRecord
	job        *tasks.LocalJob
}

type JobPool struct {
	// logger channel used to putting log records to database.
	logger chan jobLogRecord

	// register channel used to put tasks to queue.
	register chan *job

	runningJobs map[int]*runningJob

	queue []*job

	config *RunnerConfig

	processing int32
}

type RunnerRegistration struct {
	RegistrationToken string `json:"registration_token" binding:"required"`
	Webhook           string `json:"webhook"`
	MaxParallelTasks  int    `db:"max_parallel_tasks" json:"max_parallel_tasks"`
}

func (p *runningJob) Log2(msg string, now time.Time) {
	p.logRecords = append(p.logRecords, LogRecord{Time: now, Message: msg})
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

func (p *runningJob) Log(msg string) {
	p.Log2(msg, time.Now())
}

func (p *runningJob) SetStatus(status lib.TaskStatus) {
	if p.status == status {
		return
	}

	p.status = status
	p.job.SetStatus(status)
}

func (p *runningJob) LogCmd(cmd *exec.Cmd) {
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	go p.logPipe(bufio.NewReader(stderr))
	go p.logPipe(bufio.NewReader(stdout))
}

func (p *runningJob) logPipe(reader *bufio.Reader) {

	line, err := tasks.Readln(reader)
	for err == nil {
		p.Log(line)
		line, err = tasks.Readln(reader)
	}

	if err != nil && err.Error() != "EOF" {
		//don't panic on these errors, sometimes it throws not dangerous "read |0: file already closed" error
		util.LogWarningWithFields(err, log.Fields{"error": "Failed to read TaskRunner output"})
	}

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

			log.Info("Runner registered on server")

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
			if t.status == lib.TaskFailStatus {
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
				runningJob.SetStatus(lib.TaskRunningStatus)

				err := runningJob.job.Run(t.username, t.incomingVersion)

				if runningJob.status.IsFinished() {
					return
				}

				if err != nil {
					if runningJob.status == lib.TaskStoppingStatus {
						runningJob.SetStatus(lib.TaskStoppedStatus)
					} else {
						runningJob.SetStatus(lib.TaskFailStatus)
					}
				} else {
					runningJob.SetStatus(lib.TaskSuccessStatus)
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

	log.Info("Trying to register on server")

	if os.Getenv("SEMAPHORE_RUNNER_ID") != "" {

		runnerId, err := strconv.Atoi(os.Getenv("SEMAPHORE_RUNNER_ID"))

		if err != nil {
			panic(err)
		}

		if os.Getenv("SEMAPHORE_RUNNER_TOKEN") == "" {
			panic(fmt.Errorf("runner token required"))
		}

		p.config = &RunnerConfig{
			RunnerID: runnerId,
			Token:    os.Getenv("SEMAPHORE_RUNNER_TOKEN"),
		}

		return true
	}

	_, err := os.Stat(util.Config.Runner.ConfigFile)

	if err == nil {
		configBytes, err2 := os.ReadFile(util.Config.Runner.ConfigFile)

		if err2 != nil {
			panic(err2)
		}

		var config RunnerConfig

		err2 = json.Unmarshal(configBytes, &config)

		if err2 != nil {
			panic(err2)
		}

		p.config = &config

		return true
	}

	if !os.IsNotExist(err) {
		panic(err)
	}

	if util.Config.Runner.RegistrationToken == "" {
		panic("registration token cannot be empty")
	}

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

	var config RunnerConfig
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
		log.Error("Checking new jobs error, server returns code ", resp.StatusCode)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Checking new jobs, error reading response body:", err)
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

		if runJob.status == lib.TaskStoppingStatus || runJob.status == lib.TaskStoppedStatus {
			p.runningJobs[currJob.ID].job.Kill()
		}

		if runJob.status.IsFinished() {
			continue
		}

		switch runJob.status {
		case lib.TaskRunningStatus:
			if currJob.Status == lib.TaskStartingStatus || currJob.Status == lib.TaskWaitingStatus {
				continue
			}
		case lib.TaskStoppingStatus:
			if !currJob.Status.IsFinished() {
				continue
			}
		case lib.TaskConfirmed:
			if currJob.Status == lib.TaskWaitingConfirmation {
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

		taskRunner := job{
			username:        newJob.Username,
			incomingVersion: newJob.IncomingVersion,

			job: &tasks.LocalJob{
				Task:        newJob.Task,
				Template:    newJob.Template,
				Inventory:   newJob.Inventory,
				Repository:  newJob.Repository,
				Environment: newJob.Environment,
				App:         db_lib.CreateApp(newJob.Template, newJob.Repository, nil),
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

		p.queue = append(p.queue, &taskRunner)
		log.Info("Task " + strconv.Itoa(taskRunner.job.Task.ID) + " enqueued")
	}
}
