package runners

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/semaphoreui/semaphore/db"

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
	if util.Config.Runner.Connection.SkipTLSVerify {
		tlsConfig.InsecureSkipVerify = true
	}
	if util.Config.Runner.Connection.ServerCACertFile != "" {
		caCert, err := os.ReadFile(util.Config.Runner.Connection.ServerCACertFile)
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
	runningJobs map[int]*runningJob

	queue []*job

	processing int32

	keyInstaller db_lib.AccessKeyInstaller
}

func NewJobPool(keyInstaller db_lib.AccessKeyInstaller) *JobPool {
	return &JobPool{
		runningJobs:  make(map[int]*runningJob),
		queue:        make([]*job, 0),
		processing:   0,
		keyInstaller: keyInstaller,
	}
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

	client := newHTTPClient()

	url := util.Config.WebHost + "/api/internal/runners"

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}

	log.WithFields(log.Fields{
		"context": "unregistration",
		"url":     url,
	}).Debug("Sending unregistration request to the server")

	resp, err := client.Do(req)
	if err != nil {
		return
	}

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
				log.WithFields(log.Fields{
					"context": "job_running",
					"task_id": t.job.Task.ID,
					"status":  "failed",
				}).Info("Task dequeued")
				break
			}

			log.WithFields(log.Fields{
				"context":      "job_running",
				"task_id":      t.job.Task.ID,
				"queue_length": len(p.queue),
				"running_jobs": len(p.runningJobs),
			}).Debug("Dequeuing task for execution")

			// Default to starting so sendProgress never emits an empty status (invalid JSON)
			// before the job goroutine's first SetStatus(running). A rejected PUT fails the
			// whole batch and can leave the server stuck on "starting" forever.
			p.runningJobs[t.job.Task.ID] = &runningJob{
				job:    t.job,
				status: task_logger.TaskStartingStatus,
			}

			t.job.Logger = t.job.App.SetLogger(p.runningJobs[t.job.Task.ID])

			go func(runningJob *runningJob) {
				runningJob.SetStatus(task_logger.TaskRunningStatus)

				log.WithFields(log.Fields{
					"context":          "job_running",
					"task_id":          runningJob.job.Task.ID,
					"username":         t.username,
					"alias":            t.alias,
					"incoming_version": derefString(t.incomingVersion),
				}).Debug("Running job")

				err := runningJob.job.Run(t.username, t.incomingVersion, t.alias)

				log.WithError(err).WithFields(log.Fields{
					"context": "job_running",
					"task_id": runningJob.job.Task.ID,
					"status":  string(runningJob.status),
				}).Debug("Job run returned")

				if runningJob.status.IsFinished() {
					return
				}

				if err != nil {

					log.WithFields(log.Fields{
						"context":     "job_running",
						"task_id":     t.job.Task.ID,
						"task_status": t.job.Task.Status,
					}).WithError(err).Error("launch job failed")

					t.job.Logger.Log("Unable to launch the application. Please contact your system administrator for assistance.")

					if runningJob.status == task_logger.TaskStoppingStatus {
						runningJob.SetStatus(task_logger.TaskStoppedStatus)
					} else {
						runningJob.SetStatus(task_logger.TaskFailStatus)
					}
				} else {
					if runningJob.status == task_logger.TaskStoppingStatus {
						runningJob.SetStatus(task_logger.TaskStoppedStatus)
					} else {
						runningJob.SetStatus(task_logger.TaskSuccessStatus)
					}
				}

				log.WithFields(log.Fields{
					"context": "job_running",
					"task_id": runningJob.job.Task.ID,
					"status":  string(runningJob.status),
				}).Info("Task finished")
			}(p.runningJobs[t.job.Task.ID])

			p.queue = p.queue[1:]
			log.WithFields(log.Fields{
				"context": "job_running",
				"task_id": t.job.Task.ID,
				"status":  string(t.job.Task.Status),
			}).Info("Task dequeued")
			log.WithFields(log.Fields{
				"context": "job_running",
				"task_id": t.job.Task.ID,
				"status":  string(t.job.Task.Status),
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

				if util.Config.Runner.OneOff && ok && len(p.runningJobs) > 0 && !p.hasRunningJobs() {
					os.Exit(0)
				}

				p.checkNewJobs()
			}()

		}
	}
}

func (p *JobPool) sendProgress() (ok bool) {

	client := newHTTPClient()

	url := util.Config.WebHost + "/api/internal/runners"

	body := RunnerProgress{
		Jobs: nil,
	}

	for id, j := range p.runningJobs {

		body.Jobs = append(body.Jobs, JobProgress{
			ID:         id,
			LogRecords: j.logRecords,
			Status:     j.status,
			Commit:     j.commit,
		})

		log.WithFields(log.Fields{
			"context":     "sending_progress",
			"task_id":     id,
			"status":      string(j.status),
			"log_records": len(j.logRecords),
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

	req.Header.Set("X-Runner-Token", util.Config.Runner.Token)

	resp, err := client.Do(req)
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

	ok = true

	log.WithFields(log.Fields{
		"context":     "sending_progress",
		"jobs":        len(body.Jobs),
		"status_code": resp.StatusCode,
	}).Debug("Job progress accepted by the server")

	for _, jp := range body.Jobs {
		j := p.runningJobs[jp.ID]
		if j == nil {
			continue
		}
		sent := len(jp.LogRecords)
		if sent > 0 {
			if sent <= len(j.logRecords) {
				j.logRecords = j.logRecords[sent:]
			} else {
				j.logRecords = nil
			}

			log.WithFields(log.Fields{
				"context":      "sending_progress",
				"task_id":      jp.ID,
				"acknowledged": sent,
				"pending":      len(j.logRecords),
			}).Debug("Trimmed acknowledged log records")
		}
		if jp.Status.IsFinished() {
			log.WithFields(log.Fields{
				"context": "sending_progress",
				"task_id": jp.ID,
				"status":  string(jp.Status),
			}).Info("Task removed from running list")
			delete(p.runningJobs, jp.ID)
		}
	}

	return
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

	var err error
	publicKey := ""

	if util.Config.Runner.PrivateKeyFile != "" {
		publicKey, err = generatePrivateKey(util.Config.Runner.PrivateKeyFile)
	}

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "registration",
		}).Error("failed to generate private key file")
		return
	}

	client := newHTTPClient()

	url := util.Config.WebHost + "/api/internal/runners"

	jsonBytes, err := json.Marshal(RunnerRegistration{
		RegistrationToken: util.Config.Runner.RegistrationToken,
		Webhook:           util.Config.Runner.Webhook,
		Name:              util.Config.Runner.Name,
		Tags:              util.Config.Runner.Tags,
		MaxParallelTasks:  util.Config.Runner.MaxParallelTasks,
		Enabled:           util.Config.Runner.Enabled,
		PublicKey:         &publicKey,
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

	resp, err := client.Do(req)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "registration",
		}).Error("failed to send registration request to the server")
		return
	}

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

		config := util.ConfigType{}
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

	defer resp.Body.Close() //nolint:errcheck

	log.WithFields(log.Fields{
		"context":     "registration",
		"runner_name": util.Config.Runner.Name,
	}).Debug("Runner registered successfully")

	ok = true
	return
}

func loadPrivateKey(privateKeyFilePath string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(privateKeyFilePath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func generatePrivateKey(privateKeyFilePath string) (publicKey string, err error) {

	privateKeyFile, err := os.Create(privateKeyFilePath)
	if err != nil {
		return
	}
	defer privateKeyFile.Close() //nolint:errcheck

	return util.GeneratePrivateKey(privateKeyFile)
}

func decryptChunkedBytes(combinedCiphertext []byte, privateKey *rsa.PrivateKey) (fullPlaintext []byte, err error) {

	rsaBlockSize := privateKey.N.BitLen() / 8 // e.g. 256 for 2048-bit key

	// 3. Decrypt all chunks
	for i := 0; i < len(combinedCiphertext); i += rsaBlockSize {
		end := i + rsaBlockSize
		if end > len(combinedCiphertext) {
			// In case of partial/corrupted data
			end = len(combinedCiphertext)
		}
		chunk := combinedCiphertext[i:end]

		var decryptedChunk []byte
		decryptedChunk, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, chunk)
		if err != nil {
			return
		}

		// 4. Append decrypted chunk to our full plaintext buffer
		fullPlaintext = append(fullPlaintext, decryptedChunk...)
	}

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

	client := newHTTPClient()

	url := util.Config.WebHost + "/api/internal/runners"

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "checking_new_jobs",
		}).Error("failed to build new jobs request")
		return
	}

	req.Header.Set("X-Runner-Token", util.Config.Runner.Token)

	log.WithFields(log.Fields{
		"context":      "checking_new_jobs",
		"running_jobs": len(p.runningJobs),
		"queued_jobs":  len(p.queue),
	}).Debug("Fetching new jobs from the server")

	resp, err := client.Do(req)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "checking_new_jobs",
		}).Error("failed to fetch new jobs from the server")
		return
	}

	if resp.StatusCode >= 400 {

		log.WithError(fmt.Errorf("error status code")).WithFields(log.Fields{
			"context":     "checking_new_jobs",
			"status_code": resp.StatusCode,
		}).Error("server returned an error while fetching new jobs: " + p.getResponseErrorMessage(resp))
		return
	}

	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "checking_new_jobs",
		}).Error("failed to read new jobs response body")
		return
	}

	if util.Config.Runner.PrivateKeyFile != "" {
		var pk *rsa.PrivateKey

		pk, err = loadPrivateKey(util.Config.Runner.PrivateKeyFile)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context": "checking_new_jobs",
			}).Error("failed to load private key")
			return
		}

		body, err = decryptChunkedBytes(body, pk)

		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context": "checking_new_jobs",
			}).Error("failed to decrypt new jobs response body")
			return
		}

		log.WithFields(log.Fields{
			"context": "checking_new_jobs",
			"bytes":   len(body),
		}).Debug("Decrypted new jobs response body")
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

	for _, currJob := range response.CurrentJobs {
		runJob, exists := p.runningJobs[currJob.ID]

		if !exists {
			continue
		}

		if runJob.status == task_logger.TaskStoppingStatus || runJob.status == task_logger.TaskStoppedStatus {
			log.WithFields(log.Fields{
				"context": "checking_new_jobs",
				"task_id": currJob.ID,
				"status":  string(runJob.status),
			}).Debug("Killing job because it is stopping or stopped")
			p.runningJobs[currJob.ID].job.Kill()
		}

		if runJob.status.IsFinished() {
			continue
		}

		switch runJob.status {
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
			"old_status": string(runJob.status),
			"new_status": string(currJob.Status),
		}).Debug("Applying job status reported by the server")

		runJob.SetStatus(currJob.Status)
	}

	if util.Config.Runner.OneOff {
		if len(p.queue) > 0 || len(p.runningJobs) > 0 {
			return
		}
	}

	for _, newJob := range response.NewJobs {
		if _, exists := p.runningJobs[newJob.Task.ID]; exists {
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

		taskRunner := job{
			username:        newJob.Username,
			incomingVersion: newJob.IncomingVersion,
			alias:           newJob.Alias,

			job: &tasks.LocalJob{
				Task:         newJob.Task,
				Template:     newJob.Template,
				Inventory:    newJob.Inventory,
				Repository:   newJob.Repository,
				Environment:  newJob.Environment,
				KeyInstaller: p.keyInstaller,
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
				vault2 := vault
				if vault2.VaultKeyID != nil {
					key := response.AccessKeys[*vault2.VaultKeyID]
					vault2.Vault = &key
				}
				vaults = append(vaults, vault2)
			}
		}
		taskRunner.job.Template.Vaults = vaults

		if taskRunner.job.Inventory.RepositoryID != nil {
			taskRunner.job.Inventory.Repository.SSHKey = response.AccessKeys[taskRunner.job.Inventory.Repository.SSHKeyID]
		}

		p.queue = append(p.queue, &taskRunner)

		log.WithFields(log.Fields{
			"context":     "checking_new_jobs",
			"task_id":     taskRunner.job.Task.ID,
			"task_status": string(taskRunner.job.Task.Status),
		}).Info("Task enqueued")
	}
}
