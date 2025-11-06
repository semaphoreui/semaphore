package runners

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
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

func loadPublicKey(keyData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid public key")
	}
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pub, nil
}

func chunkRSAEncrypt(pub *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	// For a 2048-bit key, pub.Size() == 256 bytes
	// PKCS#1 v1.5 overhead = 11 bytes, so max plaintext per chunk = 256 - 11 = 245
	rsaBlockSize := pub.Size()        // 256 for 2048-bit
	maxChunkSize := rsaBlockSize - 11 // 245

	var encryptedBuffer bytes.Buffer

	for start := 0; start < len(plaintext); start += maxChunkSize {
		end := start + maxChunkSize
		if end > len(plaintext) {
			end = len(plaintext)
		}
		chunk := plaintext[start:end]

		encryptedChunk, err := rsa.EncryptPKCS1v15(rand.Reader, pub, chunk)
		if err != nil {
			return nil, fmt.Errorf("encrypt chunk failed: %w", err)
		}

		// Append the encrypted chunk (always 256 bytes for 2048-bit key)
		encryptedBuffer.Write(encryptedChunk)
	}

	return encryptedBuffer.Bytes(), nil
}

type RunnerController struct {
	runnerRepo        db.RunnerManager
	taskPool          *tasks.TaskPool
	encryptionService server.AccessKeyEncryptionService
}

func NewRunnerController(runnerRepo db.RunnerManager, taskPool *tasks.TaskPool, encryptionService server.AccessKeyEncryptionService) *RunnerController {
	return &RunnerController{
		runnerRepo:        runnerRepo,
		taskPool:          taskPool,
		encryptionService: encryptionService,
	}
}

func (c *RunnerController) GetRunner(w http.ResponseWriter, r *http.Request) {
	runner := helpers.GetFromContext(r, "runner").(db.Runner)

	clearCache := false

	err := c.runnerRepo.TouchRunner(runner)
	if err != nil {
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

	tasks := c.taskPool.GetRunningTasks()

	for _, tsk := range tasks {
		if tsk.RunnerID != runner.ID {
			continue
		}

		if tsk.Task.Status == task_logger.TaskStartingStatus {

			data.NewJobs = append(data.NewJobs, runners.JobData{
				Username:            tsk.Username,
				IncomingVersion:     tsk.IncomingVersion,
				Alias:               tsk.Alias,
				Task:                tsk.Task,
				Template:            tsk.Template,
				Inventory:           tsk.Inventory,
				InventoryRepository: tsk.Inventory.Repository,
				Repository:          tsk.Repository,
				Environment:         tsk.Environment,
			})

			if tsk.Inventory.SSHKeyID != nil {
				err := c.encryptionService.DeserializeSecret(&tsk.Inventory.SSHKey)
				if err != nil {
					// TODO: return error
				}
				data.AccessKeys[*tsk.Inventory.SSHKeyID] = tsk.Inventory.SSHKey
			}

			if tsk.Inventory.BecomeKeyID != nil {
				err := c.encryptionService.DeserializeSecret(&tsk.Inventory.BecomeKey)
				if err != nil {
					// TODO: return error
				}
				data.AccessKeys[*tsk.Inventory.BecomeKeyID] = tsk.Inventory.BecomeKey
			}

			if tsk.Template.Vaults != nil {
				for _, vault := range tsk.Template.Vaults {
					if vault.VaultKeyID != nil {
						err := c.encryptionService.DeserializeSecret(vault.Vault)
						if err != nil {
							// TODO: return error
						}
						data.AccessKeys[*vault.VaultKeyID] = *vault.Vault
					}
				}
			}

			if tsk.Inventory.RepositoryID != nil {
				err := c.encryptionService.DeserializeSecret(&tsk.Inventory.Repository.SSHKey)
				if err != nil {
					// TODO: return error
				}
				data.AccessKeys[tsk.Inventory.Repository.SSHKeyID] = tsk.Inventory.Repository.SSHKey
			}

			data.AccessKeys[tsk.Repository.SSHKeyID] = tsk.Repository.SSHKey

		} else {
			data.CurrentJobs = append(data.CurrentJobs, runners.JobState{
				ID:     tsk.Task.ID,
				Status: tsk.Task.Status,
			})
		}
	}

	if runner.PublicKey != nil {

		publicKey, err := loadPublicKey([]byte(*runner.PublicKey))
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		message, err := json.Marshal(data)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		encryptedBytes, err := chunkRSAEncrypt(publicKey, message)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")

		_, err = w.Write(encryptedBytes)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

	} else {
		helpers.WriteJSON(w, http.StatusOK, data)
	}

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

	for _, job := range body.Jobs {
		tsk := taskPool.GetTask(job.ID)

		if tsk == nil {
			continue
		}

		if tsk.RunnerID != runner.ID {
			helpers.WriteErrorStatus(w, "Task not assigned to this runner", http.StatusBadRequest)
			return
		}

		for _, logRecord := range job.LogRecords {
			tsk.LogWithTime(logRecord.Time, logRecord.Message)
		}

		if !job.Status.IsValid() {
			helpers.WriteErrorStatus(w, "Invalid task status", http.StatusBadRequest)
			return
		}

		tsk.SetStatus(job.Status)

		if job.Commit != nil {
			tsk.SetCommit(job.Commit.Hash, job.Commit.Message)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func RegisterRunner(w http.ResponseWriter, r *http.Request) {
	var register runners.RunnerRegistration

	if !helpers.Bind(w, r, &register) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid format",
		})
		return
	}

	if util.Config.RunnerRegistrationToken == "" || register.RegistrationToken != util.Config.RunnerRegistrationToken {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid registration token",
		})
		return
	}

	runner, err := helpers.Store(r).CreateRunner(db.Runner{
		Webhook:          register.Webhook,
		MaxParallelTasks: register.MaxParallelTasks,
		PublicKey:        register.PublicKey,
	})

	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Unexpected error",
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
