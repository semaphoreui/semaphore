package server

import (
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	log "github.com/sirupsen/logrus"
)

const secretStorageSyncTickInterval = 60 * time.Second

// SecretStorageSyncScheduler walks every sync-enabled SecretSync row
// (storage-level and env-scoped) and runs SyncSecrets when the configured
// interval has elapsed.
type SecretStorageSyncScheduler struct {
	secretSyncRepo       db.SecretSyncRepository
	secretStorageService SecretStorageService

	stop chan struct{}
	wg   sync.WaitGroup
}

func NewSecretStorageSyncScheduler(
	secretSyncRepo db.SecretSyncRepository,
	secretStorageService SecretStorageService,
) *SecretStorageSyncScheduler {
	return &SecretStorageSyncScheduler{
		secretSyncRepo:       secretSyncRepo,
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
	syncs, err := s.secretSyncRepo.GetSyncEnabledSecretSyncs()
	if err != nil {
		log.WithError(err).Warn("secret sync: failed to list sync-enabled configs")
		return
	}

	now := tz.Now()
	for _, sync := range syncs {
		if !secretSyncDue(sync, now) {
			continue
		}

		syncErr := s.secretStorageService.SyncSecrets(sync)
		markTime := tz.Now()
		success := syncErr == nil

		if syncErr != nil {
			log.WithError(syncErr).
				WithField("sync_id", sync.ID).
				WithField("storage_id", sync.StorageID).
				Warn("secret sync failed")
		}

		if err := s.secretSyncRepo.MarkSecretSyncSynced(sync.ID, success, markTime); err != nil {
			log.WithError(err).
				WithField("sync_id", sync.ID).
				Warn("secret sync: failed to record sync timestamp")
		}
	}
}

func secretSyncDue(sync db.SecretSync, now time.Time) bool {
	if sync.SyncInterval <= 0 {
		return false
	}

	var lastAttempt time.Time
	if sync.LastSyncedAt != nil {
		lastAttempt = *sync.LastSyncedAt
	}
	if sync.LastSyncFailedAt != nil && sync.LastSyncFailedAt.After(lastAttempt) {
		lastAttempt = *sync.LastSyncFailedAt
	}

	if lastAttempt.IsZero() {
		return true
	}

	return now.Sub(lastAttempt) >= time.Duration(sync.SyncInterval)*time.Minute
}
