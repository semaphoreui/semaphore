package runners

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/semaphoreui/semaphore/pkg/tz"

	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func newHTTPClient() *http.Client {
	tlsConfig := &tls.Config{}
	conn := util.Config.Runner.Connection
	if conn.SkipTLSVerify {
		tlsConfig.InsecureSkipVerify = true
	}
	if conn.ServerCACertFile != "" {
		caCert, err := os.ReadFile(conn.ServerCACertFile)
		if err == nil {
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = pool
		}
	}
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}
}

type JobPool struct {
	// mu guards runningJobs and queue. They are mutated from the Run loop
	// goroutine, the progress/poll goroutine and (indirectly) the per-job
	// goroutines, so every read and write must hold the lock. Go maps and
	// slices are not safe for concurrent access — without this lock the runtime
	// aborts the process with "concurrent map read and map write".
	mu          sync.Mutex
	runningJobs map[int]*runningJob

	queue []*job

	processing int32

	// provider is the strategy that produces per-task Executors. Built once at
	// startup from the runner's executor config. Nil when initialisation failed —
	// dispatch refuses cleanly in that case instead of panicking on first task.
	provider tasks.ExecutorProvider
	// startedAt is the process start time, sent to the server on every poll
	// (X-Runner-Started-At header). It changes on every restart, letting the
	// server detect that this runner lost its in-memory job pool.
	startedAt time.Time

	// client is the shared HTTP client for all runner→server requests. It is
	// created once so the transport's keep-alive pool reuses connections —
	// creating a client per request leaks one ESTABLISHED connection per poll
	// cycle (~2/sec) until the runner exhausts ephemeral ports (issue #3941).
	client *http.Client
}

// NewJobPool wires a runner-side job pool. The ExecutorProvider is materialised
// eagerly so config errors (bad kubeconfig, missing in-cluster credentials, etc.)
// surface in the runner logs at startup rather than at first-task dispatch.
func NewJobPool(keyInstaller db_lib.AccessKeyInstaller) *JobPool {
	pool := &JobPool{
		runningJobs: make(map[int]*runningJob),
		queue:       make([]*job, 0),
		processing:  0,
		startedAt:   tz.Now(),
		client:      newHTTPClient(),
	}

	provider, err := newExecutorProvider(util.Config.Runner.Executor, keyInstaller)
	if err != nil {
		log.WithError(err).Error("failed to initialise executor provider; runner will reject jobs until restarted with a valid config")
	} else {
		pool.provider = provider
	}

	return pool
}

// setCommonHeaders sets the headers every runner→server request carries:
// the auth token and the process start time (restart detection).
func (p *JobPool) setCommonHeaders(req *http.Request) {
	req.Header.Set("X-Runner-Token", util.Config.Runner.Token)
	req.Header.Set("X-Runner-Started-At", p.startedAt.UTC().Format(time.RFC3339))
}

// addRunningJob registers a running job under the lock.
func (p *JobPool) addRunningJob(id int, j *runningJob) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.runningJobs[id] = j
}

// getRunningJob returns the running job for id, or nil if absent.
func (p *JobPool) getRunningJob(id int) *runningJob {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.runningJobs[id]
}

// deleteRunningJob removes a running job under the lock.
func (p *JobPool) deleteRunningJob(id int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.runningJobs, id)
}

// snapshotRunningJobs returns a shallow copy of the running jobs map so callers
// can iterate without holding the lock across slow operations (HTTP, Kill).
func (p *JobPool) snapshotRunningJobs() map[int]*runningJob {
	p.mu.Lock()
	defer p.mu.Unlock()
	snapshot := make(map[int]*runningJob, len(p.runningJobs))
	for id, j := range p.runningJobs {
		snapshot[id] = j
	}
	return snapshot
}

// runningJobsCount returns the number of running jobs under the lock.
func (p *JobPool) runningJobsCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.runningJobs)
}

// resetRunningJobs replaces the running jobs map under the lock.
func (p *JobPool) resetRunningJobs() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.runningJobs = make(map[int]*runningJob)
}

// enqueue appends a job to the queue under the lock.
func (p *JobPool) enqueue(j *job) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.queue = append(p.queue, j)
}

// dequeue removes and returns the job at the front of the queue. The second
// return value is false when the queue is empty.
func (p *JobPool) dequeue() (*job, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.queue) == 0 {
		return nil, false
	}
	t := p.queue[0]
	p.queue = p.queue[1:]
	return t, true
}

// queueLen returns the queue length under the lock.
func (p *JobPool) queueLen() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.queue)
}

func (p *JobPool) existsInQueue(taskID int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, j := range p.queue {
		if j.taskID == taskID {
			return true
		}
	}

	return false
}

func (p *JobPool) hasRunningJobs() bool {
	for _, j := range p.snapshotRunningJobs() {
		if !j.getStatus().IsFinished() {
			return true
		}
	}

	return false
}

func (p *JobPool) Register(configFilePath *string) (err error) {

	ok := p.tryRegisterRunner(configFilePath)

	if !ok {
		err = fmt.Errorf("runner registration failed")
		return
	}

	return
}

func (p *JobPool) Unregister() (err error) {

	if util.Config.Runner.Token == "" {
		return fmt.Errorf("runner is not registered")
	}

	url := util.Config.WebHost + "/api/internal/runners"

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}

	log.WithFields(log.Fields{
		"context": "unregistration",
		"url":     url,
	}).Debug("Sending unregistration request to the server")

	resp, err := p.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		err = fmt.Errorf("encountered error while unregistering runner; server returned code %d", resp.StatusCode)
		return
	}

	log.WithFields(log.Fields{
		"context":     "unregistration",
		"status_code": resp.StatusCode,
	}).Debug("Runner unregistered on the server")

	if util.Config.Runner.TokenFile != "" {
		err = os.Remove(util.Config.Runner.TokenFile)
	}

	return
}

func (p *JobPool) Run() {
	launched := false

	if util.Config.Runner.Token == "" {
		log.WithFields(log.Fields{
			"context": "job_running",
		}).Panic("runner token is empty, cannot start the runner")
	}

	queueTicker := time.NewTicker(5 * time.Second)
	requestTimer := time.NewTicker(1 * time.Second)
	p.resetRunningJobs()

	defer func() {
		queueTicker.Stop()
		requestTimer.Stop()
	}()

	for {
		select {

		case <-queueTicker.C: // timer 5 seconds: get task from queue and run it

			t, ok := p.dequeue()
			if !ok {
				break
			}

			if t.status == task_logger.TaskFailStatus {
				//delete failed TaskRunner from queue
				log.WithFields(log.Fields{
					"context": "job_running",
					"task_id": t.taskID,
					"status":  "failed",
				}).Info("Task dequeued")
				break
			}

			log.WithFields(log.Fields{
				"context":      "job_running",
				"task_id":      t.taskID,
				"queue_length": p.queueLen(),
				"running_jobs": p.runningJobsCount(),
			}).Debug("Dequeuing task for execution")

			// Default to starting so sendProgress never emits an empty status (invalid JSON)
			// before the job goroutine's first SetStatus(running). A rejected PUT fails the
			// whole batch and can leave the server stuck on "starting" forever.
			rj := &runningJob{
				job:    t.job,
				taskID: t.taskID,
				status: task_logger.TaskStartingStatus,
			}
			p.addRunningJob(t.taskID, rj)

			t.job.SetLogger(rj)

			go func(running *runningJob) {
				running.SetStatus(task_logger.TaskRunningStatus)

				log.WithFields(log.Fields{
					"context":          "job_running",
					"task_id":          running.taskID,
					"username":         t.username,
					"alias":            t.alias,
					"incoming_version": derefString(t.incomingVersion),
				}).Debug("Running job")

				err := running.job.Run(t.username, t.incomingVersion, t.alias)

				if err != nil {

					log.WithFields(log.Fields{
						"context":     "job_running",
						"task_id":     t.taskID,
						"task_status": t.status,
					}).WithError(err).Error("launch job failed")

					running.Log("Unable to launch the application. Please contact your system administrator for assistance.")

					if running.getStatus() == task_logger.TaskStoppingStatus {
						running.SetStatus(task_logger.TaskStoppedStatus)
					} else {
						running.SetStatus(task_logger.TaskFailStatus)
					}
				} else {

					log.WithFields(log.Fields{
						"context": "job_running",
						"task_id": running.taskID,
						"status":  string(running.getStatus()),
					}).Debug("Job run returned")

					if running.getStatus().IsFinished() {
						return
					}

					if running.getStatus() == task_logger.TaskStoppingStatus {
						running.SetStatus(task_logger.TaskStoppedStatus)
					} else {
						running.SetStatus(task_logger.TaskSuccessStatus)
					}
				}

				log.WithFields(log.Fields{
					"context": "job_running",
					"task_id": running.taskID,
					"status":  string(running.getStatus()),
				}).Info("Task finished")
			}(rj)

			log.WithFields(log.Fields{
				"context": "job_running",
				"task_id": t.taskID,
				"status":  string(t.status),
			}).Info("Task dequeued")
			log.WithFields(log.Fields{
				"context": "job_running",
				"task_id": t.taskID,
				"status":  string(t.status),
			}).Info("Task started")

		case <-requestTimer.C:

			go func() {

				if !atomic.CompareAndSwapInt32(&p.processing, 0, 1) {
					log.WithFields(log.Fields{
						"context": "job_running",
					}).Debug("Skipping poll cycle, previous one is still in progress")
					return
				}

				defer atomic.StoreInt32(&p.processing, 0)

				ok := p.sendProgress()

				if ok && !launched {
					launched = true
					fmt.Println("Runner connected")
				}

				if util.Config.Runner.OneOff && ok && p.runningJobsCount() > 0 && !p.hasRunningJobs() {
					os.Exit(0)
				}

				p.checkNewJobs()
			}()

		}
	}
}

func (p *JobPool) sendProgress() (ok bool) {

	url := util.Config.WebHost + "/api/internal/runners"

	body := RunnerProgress{
		Jobs: nil,
	}

	for id, j := range p.snapshotRunningJobs() {

		status, logRecords, commit := j.getProgress()

		body.Jobs = append(body.Jobs, JobProgress{
			ID:         id,
			LogRecords: logRecords,
			Status:     status,
			Commit:     commit,
		})

		log.WithFields(log.Fields{
			"context":     "sending_progress",
			"task_id":     id,
			"status":      string(status),
			"log_records": len(logRecords),
		}).Debug("Including job in progress report")
	}

	log.WithFields(log.Fields{
		"context": "sending_progress",
		"jobs":    len(body.Jobs),
	}).Debug("Sending job progress to the server")

	jsonBytes, err := json.Marshal(body)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "sending_progress",
		}).Error("failed to marshal job progress request body")
		return
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "sending_progress",
		}).Error("failed to build job progress request")
		return
	}

	p.setCommonHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "sending_progress",
		}).Error("failed to send job progress to the server")
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		log.WithError(fmt.Errorf("invalid status code")).WithFields(log.Fields{
			"context":     "sending_progress",
			"jobs":        len(body.Jobs),
			"status_code": resp.StatusCode,
		}).Error("server rejected job progress")
		return
	}

	// New servers reply 200 with a body listing jobs they no longer accept
	// results for; old servers reply 204 with an empty body.
	var progressResp RunnerProgressResponse
	if respBody, readErr := io.ReadAll(resp.Body); readErr == nil && len(respBody) > 0 {
		if parseErr := json.Unmarshal(respBody, &progressResp); parseErr != nil {
			log.WithError(parseErr).WithFields(log.Fields{
				"context": "sending_progress",
			}).Warn("failed to parse job progress response from the server")
		}
	}

	ok = true

	log.WithFields(log.Fields{
		"context":     "sending_progress",
		"jobs":        len(body.Jobs),
		"status_code": resp.StatusCode,
	}).Debug("Job progress accepted by the server")

	for _, jp := range body.Jobs {
		j := p.getRunningJob(jp.ID)
		if j == nil {
			continue
		}
		sent := len(jp.LogRecords)
		if sent > 0 {
			pending := j.ackLogRecords(sent)

			log.WithFields(log.Fields{
				"context":      "sending_progress",
				"task_id":      jp.ID,
				"acknowledged": sent,
				"pending":      pending,
			}).Debug("Trimmed acknowledged log records")
		}
		if jp.Status.IsFinished() {
			log.WithFields(log.Fields{
				"context": "sending_progress",
				"task_id": jp.ID,
				"status":  string(jp.Status),
			}).Info("Task removed from running list")
			p.deleteRunningJob(jp.ID)
		}
	}

	p.applyTerminatedJobs(progressResp.TerminatedJobs)

	return
}

// applyTerminatedJobs emergency-stops jobs the server no longer accepts
// results for — the task reached a terminal status on the server (e.g. force
// stopped) while the runner was offline. The job's process is killed and the
// job is dropped from the running list without further progress reports.
func (p *JobPool) applyTerminatedJobs(taskIDs []int) {
	for _, id := range taskIDs {
		j := p.getRunningJob(id)
		if j == nil {
			continue
		}

		if !j.getStatus().IsFinished() {
			log.WithFields(log.Fields{
				"context": "sending_progress",
				"task_id": id,
				"status":  string(j.getStatus()),
			}).Warn("Server reported the task as terminated, emergency stopping the job")

			j.job.Kill()
			j.SetStatus(task_logger.TaskStoppedStatus)
		}

		log.WithFields(log.Fields{
			"context": "sending_progress",
			"task_id": id,
			"status":  string(j.getStatus()),
		}).Info("Task removed from running list")

		p.deleteRunningJob(id)
	}
}

func (p *JobPool) getResponseErrorMessage(resp *http.Response) (res string) {
	res = "the server returned error " + strconv.Itoa(resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var errRes struct {
		Error string `json:"error"`
	}

	err = json.Unmarshal(body, &errRes)
	if err != nil {
		return
	}

	res += ": " + errRes.Error

	return
}

func (p *JobPool) tryRegisterRunner(configFilePath *string) (ok bool) {

	log.Info("Registering a new runner")

	if util.Config.Runner.RegistrationToken == "" {
		log.WithError(fmt.Errorf("registration token cannot be empty")).WithFields(log.Fields{
			"context": "registration",
		}).Error("registration token is not configured")
		return
	}

	url := util.Config.WebHost + "/api/internal/runners"

	jsonBytes, err := json.Marshal(RunnerRegistration{
		RegistrationToken: util.Config.Runner.RegistrationToken,
		Webhook:           util.Config.Runner.Webhook,
		Name:              util.Config.Runner.Name,
		Tags:              util.Config.Runner.Tags,
		MaxParallelTasks:  util.Config.Runner.MaxParallelTasks,
		Enabled:           util.Config.Runner.Enabled,
		ProjectID:         util.Config.Runner.ProjectID,
	})

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "registration",
		}).Error("failed to marshal registration request body")
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "registration",
		}).Error("failed to build registration request")
		return
	}

	log.WithFields(log.Fields{
		"context":     "registration",
		"runner_name": util.Config.Runner.Name,
		"url":         url,
	}).Debug("Sending registration request to the server")

	resp, err := p.client.Do(req)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "registration",
		}).Error("failed to send registration request to the server")
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != 200 {
		log.WithError(fmt.Errorf("invalid status code")).WithFields(log.Fields{
			"context":     "registration",
			"runner_name": util.Config.Runner.Name,
			"status_code": resp.StatusCode,
		}).Error("server rejected runner registration: " + p.getResponseErrorMessage(resp))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "registration",
		}).Error("failed to read registration response body")
		return
	}

	var res struct {
		Token string `json:"token"`
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "registration",
		}).Error("failed to parse registration response from the server")
		return
	}

	if util.Config.Runner.TokenFile != "" {
		err = os.WriteFile(util.Config.Runner.TokenFile, []byte(res.Token), 0644)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context": "registration",
			}).Error("con't save runner token")
			return
		}
	} else {
		if configFilePath == nil {
			log.WithError(fmt.Errorf("config file path required")).WithFields(log.Fields{
				"context": "registration",
			}).Error("config file path is required to store the runner token")
			return
		}

		var configFileBuffer []byte
		configFileBuffer, err = os.ReadFile(*configFilePath)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context": "registration",
			}).Error("failed to read config file")
			return
		}

		config := util.ConfigType{
			Runner: &util.RunnerConfig{},
		}
		err = json.Unmarshal(configFileBuffer, &config)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context": "registration",
			}).Error("failed to parse config file")
			return
		}

		config.Runner.Token = res.Token
		configFileBuffer, err = json.MarshalIndent(&config, " ", "\t")
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context": "registration",
			}).Error("failed to marshal updated config file")
			return
		}

		err = os.WriteFile(*configFilePath, configFileBuffer, 0644)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context": "registration",
			}).Error("failed to write config file with the runner token")
			return
		}
	}

	log.WithFields(log.Fields{
		"context":     "registration",
		"runner_name": util.Config.Runner.Name,
	}).Debug("Runner registered successfully")

	ok = true
	return
}

// checkNewJobs tries to find runner to queued jobs
func (p *JobPool) checkNewJobs() {

	if util.Config.Runner.Token == "" {
		log.WithError(fmt.Errorf("no token provided")).WithFields(log.Fields{
			"context": "checking_new_jobs",
		}).Error("runner token is empty")
		return
	}

	url := util.Config.WebHost + "/api/internal/runners"

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "checking_new_jobs",
		}).Error("failed to build new jobs request")
		return
	}

	p.setCommonHeaders(req)

	log.WithFields(log.Fields{
		"context":      "checking_new_jobs",
		"running_jobs": p.runningJobsCount(),
		"queued_jobs":  p.queueLen(),
	}).Debug("Fetching new jobs from the server")

	resp, err := p.client.Do(req)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "checking_new_jobs",
		}).Error("failed to fetch new jobs from the server")
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {

		log.WithError(fmt.Errorf("error status code")).WithFields(log.Fields{
			"context":     "checking_new_jobs",
			"status_code": resp.StatusCode,
		}).Error("server returned an error while fetching new jobs: " + p.getResponseErrorMessage(resp))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "checking_new_jobs",
		}).Error("failed to read new jobs response body")
		return
	}

	var response RunnerState
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "checking_new_jobs",
		}).Error("failed to parse new jobs response from the server")
		return
	}

	log.WithFields(log.Fields{
		"context":      "checking_new_jobs",
		"current_jobs": len(response.CurrentJobs),
		"new_jobs":     len(response.NewJobs),
		"clear_cache":  response.ClearCache,
		"access_keys":  len(response.AccessKeys),
	}).Debug("Received runner state from the server")

	if response.ClearCache {
		if response.CacheCleanProjectID == nil {
			if err2 := util.Config.ClearTmpDir(); err2 != nil {
				log.WithError(err2).WithFields(log.Fields{
					"context": "checking_new_jobs",
				}).Error("failed to clear tmp directory")
			}
		} else {
			if err2 := util.Config.ClearProjectTmpDir(*response.CacheCleanProjectID); err2 != nil {
				log.WithError(err2).WithFields(log.Fields{
					"context":    "checking_new_jobs",
					"project_id": *response.CacheCleanProjectID,
				}).Error("failed to clear project tmp directory")
			}
		}
	}

	runningJobs := p.snapshotRunningJobs()

	for _, currJob := range response.CurrentJobs {
		runJob, exists := runningJobs[currJob.ID]

		if !exists {
			continue
		}

		status := runJob.getStatus()

		if status == task_logger.TaskStoppingStatus || status == task_logger.TaskStoppedStatus {
			log.WithFields(log.Fields{
				"context": "checking_new_jobs",
				"task_id": currJob.ID,
				"status":  string(status),
			}).Debug("Killing job because it is stopping or stopped")
			runJob.job.Kill()
		}

		if status.IsFinished() {
			continue
		}

		switch status {
		case task_logger.TaskRunningStatus:
			if currJob.Status == task_logger.TaskStartingStatus || currJob.Status == task_logger.TaskWaitingStatus || currJob.Status == task_logger.TaskConfirmed {
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
		case task_logger.TaskWaitingConfirmation:
			if currJob.Status == task_logger.TaskRunningStatus {
				continue
			}
		}

		log.WithFields(log.Fields{
			"context":    "checking_new_jobs",
			"task_id":    currJob.ID,
			"old_status": string(status),
			"new_status": string(currJob.Status),
		}).Debug("Applying job status reported by the server")

		runJob.SetStatus(currJob.Status)
	}

	if util.Config.Runner.OneOff {
		if p.queueLen() > 0 || p.runningJobsCount() > 0 {
			return
		}
	}

	for _, newJob := range response.NewJobs {
		if p.getRunningJob(newJob.Task.ID) != nil {
			log.WithFields(log.Fields{
				"context": "checking_new_jobs",
				"task_id": newJob.Task.ID,
			}).Debug("Skipping new job, already running")
			continue
		}

		if p.existsInQueue(newJob.Task.ID) {
			log.WithFields(log.Fields{
				"context": "checking_new_jobs",
				"task_id": newJob.Task.ID,
			}).Debug("Skipping new job, already queued")
			continue
		}

		log.WithFields(log.Fields{
			"context":     "checking_new_jobs",
			"task_id":     newJob.Task.ID,
			"template_id": newJob.Task.TemplateID,
			"project_id":  newJob.Task.ProjectID,
		}).Debug("Accepting new job from the server")

		newJob.Inventory.Repository = newJob.InventoryRepository

		executor, execErr := newExecutor(newJob, response.AccessKeys, p.provider)
		if execErr != nil {
			log.WithError(execErr).WithFields(log.Fields{
				"context":    "checking_new_jobs",
				"project_id": newJob.Task.ProjectID,
				"task_id":    newJob.Task.ID,
			}).Error("cannot construct executor for task")
			continue
		}

		taskRunner := job{
			username:        newJob.Username,
			incomingVersion: newJob.IncomingVersion,
			alias:           newJob.Alias,
			job:             executor,
			taskID:          newJob.Task.ID,
			status:          newJob.Task.Status,
		}

		p.enqueue(&taskRunner)

		log.WithFields(log.Fields{
			"context":     "checking_new_jobs",
			"task_id":     taskRunner.taskID,
			"task_status": string(taskRunner.status),
		}).Info("Task enqueued")
	}
}
