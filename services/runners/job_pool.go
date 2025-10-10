package runners

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
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

type JobLogger struct {
	Context string
}

func (e *JobLogger) ActionError(err error, action string, message string) {
	util.LogErrorF(err, log.Fields{
		"type":    "action",
		"context": e.Context,
		"action":  action,
		"error":   message,
	})
}

func (e *JobLogger) Info(message string) {
	log.WithFields(log.Fields{
		"context": e.Context,
	}).Info(message)
}

func (e *JobLogger) TaskInfo(message string, task int, status string) {
	log.WithFields(log.Fields{
		"type":    "task",
		"context": e.Context,
		"task":    task,
		"status":  status,
	}).Info(message)
}

func (e *JobLogger) Panic(err error, action string, message string) {
	log.WithFields(log.Fields{
		"context": e.Context,
	}).Panic(message)
}

func (e *JobLogger) Debug(message string) {
	log.WithFields(log.Fields{
		"context": e.Context,
	}).Debug(message)
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
	logger := JobLogger{Context: "running"}

	launched := false

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
			logger.Debug("Checking queue")

			if len(p.queue) == 0 {
				break
			}

			t := p.queue[0]
			if t.status == task_logger.TaskFailStatus {
				//delete failed TaskRunner from queue
				p.queue = p.queue[1:]
				logger.TaskInfo("Task dequeued", t.job.Task.ID, "failed")
				break
			}

			p.runningJobs[t.job.Task.ID] = &runningJob{
				job: t.job,
			}

			t.job.Logger = t.job.App.SetLogger(p.runningJobs[t.job.Task.ID])

			go func(runningJob *runningJob) {
				runningJob.SetStatus(task_logger.TaskRunningStatus)

				err := runningJob.job.Run(t.username, t.incomingVersion, t.alias)

				if runningJob.status.IsFinished() {
					return
				}

				if err != nil {
					logger.ActionError(err, "launch job", "job failed")
					t.job.Logger.Log("Unable to launch the application. Please contact your system administrator for assistance.")

					if runningJob.status == task_logger.TaskStoppingStatus {
						runningJob.SetStatus(task_logger.TaskStoppedStatus)
					} else {
						runningJob.SetStatus(task_logger.TaskFailStatus)
					}
				} else {
					runningJob.SetStatus(task_logger.TaskSuccessStatus)
				}

				logger.TaskInfo("Task finished", runningJob.job.Task.ID, string(runningJob.status))
			}(p.runningJobs[t.job.Task.ID])

			p.queue = p.queue[1:]
			logger.TaskInfo("Task dequeued", t.job.Task.ID, string(t.job.Task.Status))
			logger.TaskInfo("Task started", t.job.Task.ID, string(t.job.Task.Status))

		case <-requestTimer.C:

			go func() {

				if !atomic.CompareAndSwapInt32(&p.processing, 0, 1) {
					return
				}

				defer atomic.StoreInt32(&p.processing, 0)

				ok := p.sendProgress()

				if ok && !launched {
					launched = true
					fmt.Println("Runner connected")
				}

				if util.Config.Runner.OneOff && len(p.runningJobs) > 0 && !p.hasRunningJobs() {
					os.Exit(0)
				}

				p.checkNewJobs()
			}()

		}
	}
}

func (p *JobPool) sendProgress() (ok bool) {

	logger := JobLogger{Context: "sending_progress"}

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
			Commit:     j.commit,
		})

		j.logRecords = make([]LogRecord, 0)

		if j.status.IsFinished() {
			logger.TaskInfo("Task removed from running list", id, string(j.status))
			delete(p.runningJobs, id)
		}
	}

	jsonBytes, err := json.Marshal(body)

	if err != nil {
		logger.ActionError(err, "form request body", "can not marshal json")
		return
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		logger.ActionError(err, "create request", "can not create request to the server")
		return
	}

	req.Header.Set("X-Runner-Token", util.Config.Runner.Token)

	resp, err := client.Do(req)
	if err != nil {
		logger.ActionError(err, "send request", "the server returned error")
		return
	}

	if resp.StatusCode >= 400 {
		logger.ActionError(fmt.Errorf("invalid status code"), "send request", "the server returned error "+strconv.Itoa(resp.StatusCode))
	} else {
		ok = true
	}

	defer resp.Body.Close() //nolint:errcheck

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

	logger := JobLogger{Context: "registration"}

	log.Info("Registering a new runner")

	if util.Config.Runner.RegistrationToken == "" {
		logger.ActionError(fmt.Errorf("registration token cannot be empty"), "read input", "can not retrieve registration token")
		return
	}

	var err error
	publicKey := ""

	if util.Config.Runner.PrivateKeyFile != "" {
		publicKey, err = generatePrivateKey(util.Config.Runner.PrivateKeyFile)
	}

	if err != nil {
		logger.ActionError(err, "read input", "can not generate private key file")
		return
	}

	client := &http.Client{}

	url := util.Config.WebHost + "/api/internal/runners"

	jsonBytes, err := json.Marshal(RunnerRegistration{
		RegistrationToken: util.Config.Runner.RegistrationToken,
		Webhook:           util.Config.Runner.Webhook,
		MaxParallelTasks:  util.Config.Runner.MaxParallelTasks,
		PublicKey:         &publicKey,
	})

	if err != nil {
		logger.ActionError(err, "form request", "can not marshal json")
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		logger.ActionError(err, "create request", "can not create request to the server")
		return
	}

	resp, err := client.Do(req)

	if err != nil {
		logger.ActionError(err, "send request", "unexpected error")
		return
	}

	if resp.StatusCode != 200 {
		logger.ActionError(fmt.Errorf("invalid status code"), "send request", p.getResponseErrorMessage(resp))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ActionError(err, "read response body", "can not read server's response body")
		return
	}

	var res struct {
		Token string `json:"token"`
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		logger.ActionError(err, "parsing result json", "server's response has invalid format")
		return
	}

	if util.Config.Runner.TokenFile != "" {
		err = os.WriteFile(util.Config.Runner.TokenFile, []byte(res.Token), 0644)
	} else {
		if configFilePath == nil {
			logger.ActionError(fmt.Errorf("config file path required"), "read input", "can not retrieve config file path")
			return
		}

		var configFileBuffer []byte
		configFileBuffer, err = os.ReadFile(*configFilePath)
		if err != nil {
			logger.ActionError(err, "read config file", "can not read config file")
			return
		}

		config := util.ConfigType{}
		err = json.Unmarshal(configFileBuffer, &config)
		if err != nil {
			logger.ActionError(err, "parse config file", "can not parse config file")
			return
		}

		config.Runner.Token = res.Token
		configFileBuffer, err = json.MarshalIndent(&config, " ", "\t")
		if err != nil {
			logger.ActionError(err, "marshal config file", "can not marshal config file")
			return
		}

		err = os.WriteFile(*configFilePath, configFileBuffer, 0644)
		if err != nil {
			logger.ActionError(err, "write config file", "can not write config file")
			return
		}
	}

	defer resp.Body.Close() //nolint:errcheck

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

	logger := JobLogger{Context: "checking new jobs"}

	if util.Config.Runner.Token == "" {
		logger.ActionError(fmt.Errorf("no token provided"), "read input", "can not retrieve runner token")
		return
	}

	client := &http.Client{}

	url := util.Config.WebHost + "/api/internal/runners"

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		logger.ActionError(err, "create request", "can not create request to the server")
		return
	}

	req.Header.Set("X-Runner-Token", util.Config.Runner.Token)

	resp, err := client.Do(req)

	if err != nil {
		logger.ActionError(err, "send request", "unexpected error")
		return
	}

	if resp.StatusCode >= 400 {

		logger.ActionError(fmt.Errorf("error status code"), "send request", p.getResponseErrorMessage(resp))
		return
	}

	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ActionError(err, "read response body", "can not read server's response body")
		return
	}

	if util.Config.Runner.PrivateKeyFile != "" {
		var pk *rsa.PrivateKey

		pk, err = loadPrivateKey(util.Config.Runner.PrivateKeyFile)
		if err != nil {
			logger.ActionError(err, "decrypt response body", "can not read private key")
			return
		}

		body, err = decryptChunkedBytes(body, pk)

		if err != nil {
			logger.ActionError(err, "decrypt response body", "can not decrypt server's response body")
			return
		}
	}

	var response RunnerState
	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.ActionError(err, "parsing result json", "server's response has invalid format")
		return
	}

	if response.ClearCache {
		if response.CacheCleanProjectID == nil {
			if err2 := util.Config.ClearTmpDir(); err2 != nil {
				logger.ActionError(
					err2,
					"cleaning cache",
					"cannot clear tmp directory",
				)
			}
		} else {
			if err2 := util.Config.ClearProjectTmpDir(*response.CacheCleanProjectID); err2 != nil {
				logger.ActionError(
					err2,
					"cleaning cache",
					"cannot clear project "+strconv.Itoa(*response.CacheCleanProjectID)+" tmp directory",
				)
			}
		}
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

		logger.TaskInfo("Task enqueued", taskRunner.job.Task.ID, string(taskRunner.job.Task.Status))
	}
}
