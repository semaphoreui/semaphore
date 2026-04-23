package server

import (
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	log "github.com/sirupsen/logrus"
)

const environmentSyncTickInterval = 60 * time.Second

// EnvironmentSyncScheduler periodically runs SyncSecrets on every
// sync-enabled environment whose configured interval has elapsed.
type EnvironmentSyncScheduler struct {
	environmentRepo    db.EnvironmentManager
	environmentService EnvironmentService

	stop chan struct{}
	wg   sync.WaitGroup
}

func NewEnvironmentSyncScheduler(
	environmentRepo db.EnvironmentManager,
	environmentService EnvironmentService,
) *EnvironmentSyncScheduler {
	return &EnvironmentSyncScheduler{
		environmentRepo:    environmentRepo,
		environmentService: environmentService,
		stop:               make(chan struct{}),
	}
}

func (s *EnvironmentSyncScheduler) Start() {
	s.wg.Add(1)
	go s.run()
}

func (s *EnvironmentSyncScheduler) Stop() {
	close(s.stop)
	s.wg.Wait()
}

func (s *EnvironmentSyncScheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(environmentSyncTickInterval)
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

func (s *EnvironmentSyncScheduler) tick() {
	environments, err := s.environmentRepo.GetSyncEnabledEnvironments()
	if err != nil {
		log.WithError(err).Warn("environment sync: failed to list sync-enabled environments")
		return
	}

	now := tz.Now()
	for _, env := range environments {
		if !environmentSyncDue(env, now) {
			continue
		}

		syncErr := s.environmentService.SyncSecrets(env)
		markTime := tz.Now()
		success := syncErr == nil

		if syncErr != nil {
			log.WithError(syncErr).
				WithField("environment_id", env.ID).
				WithField("project_id", env.ProjectID).
				Warn("environment sync failed")
		}

		if err := s.environmentRepo.MarkEnvironmentSynced(env.ID, success, markTime); err != nil {
			log.WithError(err).
				WithField("environment_id", env.ID).
				Warn("environment sync: failed to record sync timestamp")
		}
	}
}

func environmentSyncDue(env db.Environment, now time.Time) bool {
	if env.SyncInterval <= 0 {
		return false
	}

	var lastAttempt time.Time
	if env.LastSyncedAt != nil {
		lastAttempt = *env.LastSyncedAt
	}
	if env.LastSyncFailedAt != nil && env.LastSyncFailedAt.After(lastAttempt) {
		lastAttempt = *env.LastSyncFailedAt
	}

	if lastAttempt.IsZero() {
		return true
	}

	return now.Sub(lastAttempt) >= time.Duration(env.SyncInterval)*time.Minute
}
