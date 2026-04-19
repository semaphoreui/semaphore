package server

import (
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	log "github.com/sirupsen/logrus"
)

const secretStorageSyncTickInterval = 60 * time.Second

// SecretStorageSyncScheduler periodically runs SyncSecrets on every
// sync-enabled secret storage whose configured interval has elapsed.
type SecretStorageSyncScheduler struct {
	secretStorageRepo    db.SecretStorageRepository
	secretStorageService SecretStorageService

	stop chan struct{}
	wg   sync.WaitGroup
}

func NewSecretStorageSyncScheduler(
	secretStorageRepo db.SecretStorageRepository,
	secretStorageService SecretStorageService,
) *SecretStorageSyncScheduler {
	return &SecretStorageSyncScheduler{
		secretStorageRepo:    secretStorageRepo,
		secretStorageService: secretStorageService,
		stop:                 make(chan struct{}),
	}
}

func (s *SecretStorageSyncScheduler) Start() {
	s.wg.Add(1)
	go s.run()
}

func (s *SecretStorageSyncScheduler) Stop() {
	close(s.stop)
	s.wg.Wait()
}

func (s *SecretStorageSyncScheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(secretStorageSyncTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *SecretStorageSyncScheduler) tick() {
	storages, err := s.secretStorageRepo.GetSyncEnabledSecretStorages()
	if err != nil {
		log.WithError(err).Warn("secret storage sync: failed to list sync-enabled storages")
		return
	}

	now := tz.Now()
	for _, storage := range storages {
		if !secretStorageSyncDue(storage, now) {
			continue
		}

		syncErr := s.secretStorageService.SyncSecrets(storage)
		markTime := tz.Now()
		success := syncErr == nil

		if syncErr != nil {
			log.WithError(syncErr).
				WithField("storage_id", storage.ID).
				WithField("project_id", storage.ProjectID).
				Warn("secret storage sync failed")
		}

		if err := s.secretStorageRepo.MarkSecretStorageSynced(storage.ID, success, markTime); err != nil {
			log.WithError(err).
				WithField("storage_id", storage.ID).
				Warn("secret storage sync: failed to record sync timestamp")
		}
	}
}

func secretStorageSyncDue(storage db.SecretStorage, now time.Time) bool {
	if storage.SyncInterval <= 0 {
		return false
	}

	var lastAttempt time.Time
	if storage.LastSyncedAt != nil {
		lastAttempt = *storage.LastSyncedAt
	}
	if storage.LastSyncFailedAt != nil && storage.LastSyncFailedAt.After(lastAttempt) {
		lastAttempt = *storage.LastSyncFailedAt
	}

	if lastAttempt.IsZero() {
		return true
	}

	return now.Sub(lastAttempt) >= time.Duration(storage.SyncInterval)*time.Minute
}
