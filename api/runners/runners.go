package runners

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/ssh"
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

	tasks := c.taskPool.GetRunningTasks()

	for _, tsk := range tasks {
		if tsk.Task.RunnerID == nil || *tsk.Task.RunnerID != runner.ID {
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
					log.WithFields(log.Fields{
						"runner_id":     runner.ID,
						"task_id":       tsk.Task.ID,
						"inventory_id":  tsk.Inventory.ID,
						"access_key_id": tsk.Inventory.SSHKey.ID,
						"context":       "runner",
					}).WithError(err).Error("Failed to decrypt inventory key")
					helpers.WriteError(w, err)
					return
				}
				data.AccessKeys[*tsk.Inventory.SSHKeyID] = tsk.Inventory.SSHKey
			}

			if tsk.Inventory.BecomeKeyID != nil {
				err := c.encryptionService.DeserializeSecret(&tsk.Inventory.BecomeKey)
				if err != nil {
					log.WithFields(log.Fields{
						"runner_id":     runner.ID,
						"task_id":       tsk.Task.ID,
						"inventory_id":  tsk.Inventory.ID,
						"access_key_id": tsk.Inventory.BecomeKey.ID,
						"context":       "runner",
					}).WithError(err).Error("Failed to decrypt become key")
					helpers.WriteError(w, err)
					return
				}
				data.AccessKeys[*tsk.Inventory.BecomeKeyID] = tsk.Inventory.BecomeKey
			}

			if tsk.Template.Vaults != nil {
				for _, vault := range tsk.Template.Vaults {
					if vault.VaultKeyID != nil {
						err := c.encryptionService.DeserializeSecret(vault.Vault)
						if err != nil {
							log.WithFields(log.Fields{
								"runner_id":     runner.ID,
								"task_id":       tsk.Task.ID,
								"access_key_id": vault.Vault.ID,
								"context":       "runner",
							}).WithError(err).Error("Failed to decrypt vault")
							helpers.WriteError(w, err)
							return
						}
						data.AccessKeys[*vault.VaultKeyID] = *vault.Vault
					}
				}
			}

			if tsk.Inventory.RepositoryID != nil {
				err := c.encryptionService.DeserializeSecret(&tsk.Inventory.Repository.SSHKey)
				if err != nil {
					log.WithFields(log.Fields{
						"runner_id":     runner.ID,
						"task_id":       tsk.Task.ID,
						"repository_id": tsk.Inventory.Repository.ID,
						"access_key_id": tsk.Inventory.Repository.SSHKey.ID,
						"context":       "runner",
					}).WithError(err).Error("Failed to decrypt repository key")
					helpers.WriteError(w, err)
					return
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
		tsk, err := taskPool.GetTask(job.ID)

		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"task_id":   job.ID,
				"runner_id": runner.ID,
				"context":   "runner",
			}).Warn("runner progress: task not in local pool and could not be loaded from database")
			continue
		}

		if tsk == nil {
			log.WithFields(log.Fields{
				"task_id":   job.ID,
				"runner_id": runner.ID,
				"context":   "runner",
			}).Warn("runner progress: task not found in pool")
			continue
		}

		if tsk.Task.RunnerID == nil || *tsk.Task.RunnerID != runner.ID {
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

		if !tsk.Task.Status.IsFinished() {
			tsk.SetStatus(job.Status)

			if job.Commit != nil {
				tsk.SetCommit(job.Commit.Hash, job.Commit.Message)
			}
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

	if register.RegistrationToken == "" {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid registration token",
		})
		return
	}

	store := helpers.Store(r)

	var runner db.Runner
	var err error

	if util.Config.RunnerRegistrationToken != "" && register.RegistrationToken == util.Config.RunnerRegistrationToken {
		// The shared, global registration token creates a brand-new runner.
		runner, err = store.CreateRunner(db.Runner{
			Token:            db.GenerateRunnerToken(),
			Webhook:          register.Webhook,
			Name:             register.Name,
			Tags:             register.Tags,
			MaxParallelTasks: register.MaxParallelTasks,
			Active:           register.Enabled,
			PublicKey:        register.PublicKey,
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
		// Otherwise the value is a one-time registration token issued for a specific
		// unregistered runner. The global token cannot be used to register it.
		runner, err = store.RegisterRunner(
			server.HashRunnerRegistrationToken(register.RegistrationToken),
			register.PublicKey,
			register.Enabled,
		)

		if err != nil {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid registration token",
			})
			return
		}
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

type RepositoryRequest struct {
	GitURL    string `json:"git_url" binding:"required"`
	GitBranch string `json:"git_branch" binding:"required"`
	SSHKeyID  *int   `json:"ssh_key_id,omitempty"`
}

type RepositoryResponse struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Archive []byte `json:"archive"`
}

func GetRepositoryArchive(w http.ResponseWriter, r *http.Request) {
	runner := helpers.GetFromContext(r, "runner").(db.Runner)

	var req RepositoryRequest
	if !helpers.Bind(w, r, &req) {
		return
	}

	store := helpers.Store(r)

	// Create a temporary repository record for cloning
	tempRepo := db.Repository{
		GitURL:    req.GitURL,
		GitBranch: req.GitBranch,
	}

	// Set SSH key if provided
	if req.SSHKeyID != nil && runner.ProjectID != nil {
		accessKey, err := store.GetAccessKey(*runner.ProjectID, *req.SSHKeyID)
		if err != nil {
			helpers.WriteErrorStatus(w, "Access key not found", http.StatusNotFound)
			return
		}
		tempRepo.SSHKeyID = *req.SSHKeyID
		tempRepo.SSHKey = accessKey
	}

	// Create a temporary directory for cloning
	tempDir, err := os.MkdirTemp("", "semaphore-repo-*")
	if err != nil {
		log.WithError(err).Error("Failed to create temp directory")
		helpers.WriteErrorStatus(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	// Create a simple logger for the git operations
	logger := &simpleLogger{}

	// Create git repository with the appropriate git client (but not proxy to avoid recursion)
	var gitClient db_lib.GitClient
	switch util.Config.GitClientId {
	case util.GoGitClientId:
		gitClient = db_lib.CreateGoGitClient(&simpleKeyInstaller{})
	default:
		gitClient = db_lib.CreateCmdGitClient(&simpleKeyInstaller{})
	}

	gitRepo := db_lib.GitRepository{
		Repository: tempRepo,
		Logger:     logger,
		Client:     gitClient,
	}

	// Create a custom GitRepository that returns our temp directory
	customGitRepo := customGitRepository{
		GitRepository: gitRepo,
		customPath:    tempDir,
	}

	// Clone the repository
	err = customGitRepo.Clone()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"git_url":    req.GitURL,
			"git_branch": req.GitBranch,
		}).Error("Failed to clone repository")
		helpers.WriteErrorStatus(w, "Failed to clone repository: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get commit information
	hash, err := customGitRepo.GetLastCommitHash()
	if err != nil {
		log.WithError(err).Error("Failed to get commit hash")
		hash = "unknown"
	}

	message, err := customGitRepo.GetLastCommitMessage()
	if err != nil {
		log.WithError(err).Error("Failed to get commit message")
		message = "unknown"
	}

	// Create tar.gz archive of the repository
	archiveData, err := createRepositoryArchive(tempDir)
	if err != nil {
		log.WithError(err).Error("Failed to create repository archive")
		helpers.WriteErrorStatus(w, "Failed to create archive", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := RepositoryResponse{
		Hash:    hash,
		Message: message,
		Archive: archiveData,
	}

	log.WithFields(log.Fields{
		"git_url":     req.GitURL,
		"git_branch":  req.GitBranch,
		"commit_hash": hash,
		"runner_id":   runner.ID,
	}).Info("Repository archive served to runner")

	helpers.WriteJSON(w, http.StatusOK, response)
}

// Simple logger implementation for git operations
type simpleLogger struct {
	status task_logger.TaskStatus
}

type StatusListener = task_logger.StatusListener
type LogListener = task_logger.LogListener
type TaskStatus = task_logger.TaskStatus

func (l *simpleLogger) Log(message string) {
	log.Info(message)
}

func (l *simpleLogger) Logf(format string, a ...any) {
	log.Infof(format, a...)
}

func (l *simpleLogger) LogWithTime(time time.Time, message string) {
	log.Info(message)
}

func (l *simpleLogger) LogfWithTime(time time.Time, format string, a ...any) {
	log.Infof(format, a...)
}

func (l *simpleLogger) LogCmd(cmd *exec.Cmd) {
	log.Infof("Executing command: %v", cmd)
}

func (l *simpleLogger) SetStatus(status TaskStatus) {
	l.status = status
}

func (l *simpleLogger) AddStatusListener(listener StatusListener) {
	// No-op for simple implementation
}

func (l *simpleLogger) AddLogListener(listener LogListener) {
	// No-op for simple implementation
}

func (l *simpleLogger) SetCommit(hash, message string) {
	// No-op for simple implementation
}

func (l *simpleLogger) WaitLog() {
	// No-op for simple implementation
}

// Simple key installer implementation for git operations
type simpleKeyInstaller struct{}

func (k *simpleKeyInstaller) Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (ssh.AccessKeyInstallation, error) {
	// For now, return a simple implementation that doesn't install keys
	// This will work for public repositories or repositories that don't require authentication
	return ssh.AccessKeyInstallation{}, nil
}

// Custom GitRepository wrapper that overrides GetFullPath
type customGitRepository struct {
	db_lib.GitRepository
	customPath string
}

func (c customGitRepository) GetFullPath() string {
	return c.customPath
}

func createRepositoryArchive(repoPath string) ([]byte, error) {
	var buf bytes.Buffer

	// Create gzip writer
	gzWriter := gzip.NewWriter(&buf)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk through the repository directory
	err := filepath.Walk(repoPath, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(repoPath, file)
		if err != nil {
			return err
		}

		header.Name = relPath

		// Write header
		err = tarWriter.WriteHeader(header)
		if err != nil {
			return err
		}

		// If it's a regular file, write its contents
		if fi.Mode().IsRegular() {
			fileData, err := os.Open(file)
			if err != nil {
				return err
			}
			defer fileData.Close()

			_, err = io.Copy(tarWriter, fileData)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Close writers to flush data
	tarWriter.Close()
	gzWriter.Close()

	return buf.Bytes(), nil
}
