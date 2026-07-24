package runners

import (
	"errors"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/jwt"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/runners"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

func RunnerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token := r.Header.Get("X-Runner-Token")

		if token == "" {
			helpers.WriteJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid token",
			})
			return
		}

		store := helpers.Store(r)

		runner, err := store.GetRunnerByToken(token)

		if err != nil {
			helpers.WriteJSON(w, http.StatusNotFound, map[string]string{
				"error": "Runner not found",
			})
			return
		}

		if runner.Token != token {
			helpers.WriteJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid token",
			})
			return
		}

		r = helpers.SetContextValue(r, "runner", runner)
		next.ServeHTTP(w, r)
	})
}

type RunnerController struct {
	runnerRepo        db.RunnerManager
	taskPool          *tasks.TaskPool
	encryptionService server.AccessKeyEncryptionService
	signer            jwt.Signer
}

func NewRunnerController(runnerRepo db.RunnerManager, taskPool *tasks.TaskPool, encryptionService server.AccessKeyEncryptionService, signer jwt.Signer) *RunnerController {
	return &RunnerController{
		runnerRepo:        runnerRepo,
		taskPool:          taskPool,
		encryptionService: encryptionService,
		signer:            signer,
	}
}

func (c *RunnerController) GetRunner(w http.ResponseWriter, r *http.Request) {
	runner := helpers.GetFromContext(r, "runner").(db.Runner)

	clearCache := false

	// The runner reports its process start time on every poll. It changes on
	// every restart and is persisted next to "touched", so the task reconciler
	// can detect runners that restarted and lost their in-memory job pool.
	if v := r.Header.Get("X-Runner-Started-At"); v != "" {
		if startedAt, parseErr := time.Parse(time.RFC3339, v); parseErr == nil {
			runner.StartedAt = &startedAt
		} else {
			log.WithFields(log.Fields{
				"runner_id": runner.ID,
				"context":   "runner",
			}).WithError(parseErr).Warn("invalid X-Runner-Started-At header")
		}
	}

	if err := c.runnerRepo.TouchRunner(runner); err != nil {
		log.WithFields(log.Fields{
			"runner_id": runner.ID,
			"context":   "runner",
		}).WithError(err).Error("runner touch failed")
		helpers.WriteError(w, err)
		return
	}

	if runner.CleaningRequested != nil && (runner.Touched == nil || runner.CleaningRequested.After(*runner.Touched)) {
		clearCache = true
	}

	data := runners.RunnerState{
		AccessKeys: make(map[int]db.AccessKey),
		ClearCache: clearCache,
	}

	if clearCache {
		data.CacheCleanProjectID = runner.ProjectID
	}

	runningTasks := c.taskPool.GetRunningTasks()

	for _, tsk := range runningTasks {
		if tsk.Task.RunnerID == nil || *tsk.Task.RunnerID != runner.ID {
			continue
		}

		if tsk.Task.Status == task_logger.TaskWaitingStatus || tsk.Task.Status == task_logger.TaskStartingStatus {

			c.prepareRemoteJob(tsk, &runner, &data)

		} else {
			data.CurrentJobs = append(data.CurrentJobs, runners.JobState{
				ID:     tsk.Task.ID,
				Status: tsk.Task.Status,
			})
		}
	}

	helpers.WriteJSON(w, http.StatusOK, data)
}

// prepareRemoteJob builds the job for a waiting/starting task and stages it,
// together with the task's decrypted access keys, into data. On any failure it
// fails and finalizes the task in place, leaving data untouched, so a single
// bad task does not abort the poll for the whole runner.
func (c *RunnerController) prepareRemoteJob(tsk *tasks.TaskRunner, runner *db.Runner, data *runners.RunnerState) {
	// Survey secret variables are stored as a task-bound encrypted
	// access key in the shared DB, so any HA node serving this poll can
	// deliver them. An unreadable (e.g. expired) secret fails the task
	// instead of dispatching it with silently empty variables.
	surveySecrets, err := c.encryptionService.GetTaskSurveySecrets(tsk.Task.ProjectID, tsk.Task.ID)
	if err != nil {
		logger := log.WithError(err).WithFields(log.Fields{
			"runner_id": runner.ID,
			"task_id":   tsk.Task.ID,
			"context":   "survey_secrets",
		})
		if errors.Is(err, server.ErrAccessKeyExpired) {
			logger.Warn("task survey secrets expired before dispatch")
			tsk.Log("Survey secrets expired before the task started. Please run the task again.")
		} else {
			logger.Error("failed to read task survey secrets")
			tsk.Log("Failed to read survey secrets. More details in the server logs.")
		}
		tsk.SetStatus(task_logger.TaskFailStatus)
		c.taskPool.FinalizeRemoteTask(tsk, runner)
		return
	}

	jobData := runners.JobData{
		Username:            tsk.Username,
		IncomingVersion:     tsk.IncomingVersion,
		Alias:               tsk.Alias,
		Task:                tsk.Task,
		Template:            tsk.Template,
		Inventory:           tsk.Inventory,
		InventoryRepository: tsk.Inventory.Repository,
		Repository:          tsk.Repository,
		Environment:         tsk.Environment,
	}

	// Always overwrite: the dispatched Secret must be exactly the
	// DB-derived value, never whatever the in-memory task carries
	jobData.Task.Secret = surveySecrets

	if c.signer != nil && tsk.Template.JWTParams != nil && tsk.Template.JWTParams.Enabled {
		ttl, terr := tsk.Template.JWTParams.ParsedTTL()
		if terr != nil {
			log.WithError(terr).WithFields(log.Fields{
				"task_id":     tsk.Task.ID,
				"template_id": tsk.Template.ID,
				"context":     "jwt",
			}).Warn("invalid template jwt_params.ttl")
			tsk.Log("Invalid JWT token lifetime in the template settings: " + terr.Error())
			tsk.SetStatus(task_logger.TaskFailStatus)
			c.taskPool.FinalizeRemoteTask(tsk, runner)
			return
		}

		token, jerr := c.signer.Sign(jwt.TaskInfo{
			TaskID:     tsk.Task.ID,
			ProjectID:  tsk.Task.ProjectID,
			TemplateID: tsk.Template.ID,
			UserID:     tsk.Task.UserID,
			Audience:   tsk.Template.JWTParams.Audience,
			TTL:        ttl,
		})
		if jerr != nil {
			log.WithError(jerr).WithFields(log.Fields{
				"task_id": tsk.Task.ID,
				"context": "jwt",
			}).Error("failed to sign task JWT")
			tsk.Log("Failed to sign the task JWT. More details in the server logs.")
			tsk.SetStatus(task_logger.TaskFailStatus)
			c.taskPool.FinalizeRemoteTask(tsk, runner)
			return
		}

		jobData.JWT = token
	}

	// Decrypt all keys of the task before publishing anything, so a
	// task with an undecryptable key fails on its own instead of
	// aborting the poll for the whole runner, and none of its keys
	// leak into the response of a job that is not dispatched.
	taskKeys := make(map[int]db.AccessKey)
	if kerr := c.collectTaskAccessKeys(tsk, runner.ID, taskKeys); kerr != nil {
		tsk.Log("Failed to decrypt access keys of the task. More details in the server logs.")
		tsk.SetStatus(task_logger.TaskFailStatus)
		c.taskPool.FinalizeRemoteTask(tsk, runner)
		return
	}

	maps.Copy(data.AccessKeys, taskKeys)
	data.NewJobs = append(data.NewJobs, jobData)
}

// collectTaskAccessKeys decrypts every access key the dispatched task needs
// and stages them into keys, so the caller publishes either all of them or
// none. It returns the first decryption error.
func (c *RunnerController) collectTaskAccessKeys(tsk *tasks.TaskRunner, runnerID int, keys map[int]db.AccessKey) error {
	if tsk.Inventory.SSHKeyID != nil {
		if err := c.encryptionService.DeserializeSecret(&tsk.Inventory.SSHKey); err != nil {
			log.WithFields(log.Fields{
				"runner_id":     runnerID,
				"task_id":       tsk.Task.ID,
				"inventory_id":  tsk.Inventory.ID,
				"access_key_id": tsk.Inventory.SSHKey.ID,
				"context":       "runner",
			}).WithError(err).Error("Failed to decrypt inventory key")
			return err
		}
		keys[*tsk.Inventory.SSHKeyID] = tsk.Inventory.SSHKey
	}

	if tsk.Inventory.BecomeKeyID != nil {
		if err := c.encryptionService.DeserializeSecret(&tsk.Inventory.BecomeKey); err != nil {
			log.WithFields(log.Fields{
				"runner_id":     runnerID,
				"task_id":       tsk.Task.ID,
				"inventory_id":  tsk.Inventory.ID,
				"access_key_id": tsk.Inventory.BecomeKey.ID,
				"context":       "runner",
			}).WithError(err).Error("Failed to decrypt become key")
			return err
		}
		keys[*tsk.Inventory.BecomeKeyID] = tsk.Inventory.BecomeKey
	}

	for _, vault := range tsk.Template.Vaults {
		if vault.VaultKeyID == nil {
			continue
		}
		if err := c.encryptionService.DeserializeSecret(vault.Vault); err != nil {
			log.WithFields(log.Fields{
				"runner_id":     runnerID,
				"task_id":       tsk.Task.ID,
				"access_key_id": vault.Vault.ID,
				"context":       "runner",
			}).WithError(err).Error("Failed to decrypt vault")
			return err
		}
		keys[*vault.VaultKeyID] = *vault.Vault
	}

	if tsk.Inventory.RepositoryID != nil {
		if err := c.encryptionService.DeserializeSecret(&tsk.Inventory.Repository.SSHKey); err != nil {
			log.WithFields(log.Fields{
				"runner_id":     runnerID,
				"task_id":       tsk.Task.ID,
				"repository_id": tsk.Inventory.Repository.ID,
				"access_key_id": tsk.Inventory.Repository.SSHKey.ID,
				"context":       "runner",
			}).WithError(err).Error("Failed to decrypt repository key")
			return err
		}
		keys[tsk.Inventory.Repository.SSHKeyID] = tsk.Inventory.Repository.SSHKey
	}

	// Decrypt the task repository key here rather than relying on it having
	// been decrypted earlier (e.g. in TaskRunner.populateDetails). Decryption
	// reads key.Secret (the ciphertext, left intact) and refills the plaintext
	// field, so this is idempotent even if the key is already decrypted.
	if err := c.encryptionService.DeserializeSecret(&tsk.Repository.SSHKey); err != nil {
		log.WithFields(log.Fields{
			"runner_id":     runnerID,
			"task_id":       tsk.Task.ID,
			"repository_id": tsk.Repository.ID,
			"access_key_id": tsk.Repository.SSHKey.ID,
			"context":       "runner",
		}).WithError(err).Error("Failed to decrypt repository key")
		return err
	}
	keys[tsk.Repository.SSHKeyID] = tsk.Repository.SSHKey

	return nil
}

func (c *RunnerController) UpdateRunner(w http.ResponseWriter, r *http.Request) {

	runner := helpers.GetFromContext(r, "runner").(db.Runner)

	var body runners.RunnerProgress

	if !helpers.Bind(w, r, &body) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid format",
		})
		return
	}

	taskPool := c.taskPool

	if body.Jobs == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var response runners.RunnerProgressResponse

	for _, job := range body.Jobs {
		tsk, err := taskPool.GetTask(job.ID)

		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				// The task no longer exists at all — the runner's job is orphaned.
				response.TerminatedJobs = append(response.TerminatedJobs, job.ID)
				continue
			}
			log.WithError(err).WithFields(log.Fields{
				"task_id":   job.ID,
				"runner_id": runner.ID,
				"context":   "runner",
			}).Warn("runner progress: task not in local pool and could not be loaded from database")
			continue
		}

		if tsk == nil {
			// Not in the pool. The task may have been stopped and finalized while
			// the runner was offline; check the persisted row so the runner can be
			// told to emergency-stop instead of running the job forever.
			row, rowErr := helpers.Store(r).GetTaskByID(job.ID)
			switch {
			case rowErr != nil && errors.Is(rowErr, db.ErrNotFound):
				// The task no longer exists at all — the runner's job is orphaned.
				response.TerminatedJobs = append(response.TerminatedJobs, job.ID)
			case rowErr == nil && row.Status.IsFinished():
				response.TerminatedJobs = append(response.TerminatedJobs, job.ID)
			case rowErr != nil:
				log.WithError(rowErr).WithFields(log.Fields{
					"task_id":   job.ID,
					"runner_id": runner.ID,
					"context":   "runner",
				}).Warn("runner progress: task not in pool and could not be loaded from database")
			default:
				log.WithFields(log.Fields{
					"task_id":   job.ID,
					"runner_id": runner.ID,
					"context":   "runner",
				}).Warn("runner progress: task not found in pool")
			}
			continue
		}

		if tsk.Task.RunnerID == nil || *tsk.Task.RunnerID != runner.ID {
			// The task was reassigned (e.g. reconciler requeued it off an offline
			// runner). Tell this runner to emergency-stop the job instead of
			// rejecting the whole progress batch with 400 — sendProgress treats
			// >=400 as total failure and never applies terminated_jobs, so the
			// old runner would keep executing alongside the new assignee.
			response.TerminatedJobs = append(response.TerminatedJobs, job.ID)
			continue
		}

		if !job.Status.IsValid() {
			helpers.WriteErrorStatus(w, "Invalid task status", http.StatusBadRequest)
			return
		}

		// The task already reached a terminal status on the server (e.g. force
		// stopped while the runner was offline, or failed by the reconciler).
		// Reject the report — logs included — and tell the runner to
		// emergency-stop the job.
		if tsk.Task.Status.IsFinished() {
			response.TerminatedJobs = append(response.TerminatedJobs, job.ID)
			continue
		}

		for _, logRecord := range job.LogRecords {
			tsk.LogWithTime(logRecord.Time, logRecord.Message)
		}

		tsk.SetStatus(job.Status)

		if job.Commit != nil {
			tsk.SetCommit(job.Commit.Hash, job.Commit.Message)
		}

		// When the runner reports a terminal status, finalize the task here:
		// finish webhook, autorun children, and pool/Redis state cleanup.
		// This is what completes a task independently of the node that
		// originally dispatched it.
		if tsk.Task.Status.IsFinished() {
			runner := runner
			go taskPool.FinalizeRemoteTask(tsk, &runner)
		}
	}

	helpers.WriteJSON(w, http.StatusOK, response)
}

func RegisterRunner(w http.ResponseWriter, r *http.Request) {
	var register runners.RunnerRegistration

	if !helpers.Bind(w, r, &register) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid format",
		})
		return
	}

	if register.RegistrationToken == "" {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid registration token",
		})
		return
	}

	store := helpers.Store(r)

	var runner db.Runner
	var err error

	if strings.HasPrefix(register.RegistrationToken, "smrs_") {
		// Otherwise the value is a one-time registration token issued for a specific
		// unregistered runner. The global token cannot be used to register it.
		runner, err = store.RegisterRunner(server.HashRunnerRegistrationToken(register.RegistrationToken), nil)

		if err != nil {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid registration token",
			})
			return
		}
	} else if util.Config.GetRunnerRegistrationToken() != "" && register.RegistrationToken == util.Config.GetRunnerRegistrationToken() {
		// The shared, global registration token creates a brand-new runner.
		runner, err = store.CreateRunner(db.Runner{
			Token:            db.GenerateRunnerToken(),
			Webhook:          register.Webhook,
			Name:             register.Name,
			Tags:             register.Tags,
			MaxParallelTasks: register.MaxParallelTasks,
			Active:           register.Enabled,
			ProjectID:        register.ProjectID,
		})

		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context": "runner",
			}).Error("Can't create runner")

			helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Unexpected error",
			})
			return
		}
	} else {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid registration token",
		})
		return
	}

	log.WithFields(log.Fields{
		"runner_id": runner.ID,
		"context":   "runner",
	}).Info("New runner registered")

	var res struct {
		Token string `json:"token"`
	}

	res.Token = runner.Token

	helpers.WriteJSON(w, http.StatusOK, res)
}

func UnregisterRunner(w http.ResponseWriter, r *http.Request) {

	runner := helpers.GetFromContext(r, "runner").(db.Runner)

	err := helpers.Store(r).DeleteGlobalRunner(runner.ID)

	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Unknown error",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
