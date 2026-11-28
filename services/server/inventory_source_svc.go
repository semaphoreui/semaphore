package server

import (
	"context"
	"errors"
	"runtime/debug"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/tz"
	proServer "github.com/semaphoreui/semaphore/pro/services/server"
	log "github.com/sirupsen/logrus"
)

type InventorySourceFetcher interface {
	Fetch(ctx context.Context, source db.InventorySource) (db.InventoryDocument, error)
}

// InventorySourceService imports the document of an inventory source in the background
type InventorySourceService struct {
	store   db.Store
	git     InventorySourceFetcher
	secrets db_lib.SecretDeserializer
	now     func() time.Time
	slots   chan struct{}
	// timeout is how long a run may take, after it the sync is shown as interrupted and another one may start
	timeout time.Duration
}

// maxInventorySyncsAtOnce keeps the memory of a server bounded, a parsed inventory may take hundreds of megabytes
const maxInventorySyncsAtOnce = 4

func NewInventorySourceService(store db.Store, git InventorySourceFetcher, secrets db_lib.SecretDeserializer) *InventorySourceService {
	return &InventorySourceService{store: store, git: git, secrets: secrets, now: tz.Now, slots: make(chan struct{}, maxInventorySyncsAtOnce), timeout: db.InventorySourceStaleAfter}
}

func (s *InventorySourceService) Validate(source db.InventorySource) error {
	if err := source.Validate(); err != nil {
		return err
	}

	if source.Type != db.InventorySourceGit {
		return proServer.ValidateInventorySource(source)
	}

	if source.RepositoryID == nil {
		return common_errors.NewValidationError("source repository must not be empty")
	}

	return nil
}

// Sync returns false when the source is already syncing, done is called once when the sync ends
func (s *InventorySourceService) Sync(source db.InventorySource, done func(db.InventorySyncResult, error)) (bool, error) {
	now := s.now()
	run, err := s.store.StartInventorySourceSync(source.ProjectID, source.InventoryID, source.ID, now, now.Add(-db.InventorySourceStaleAfter))
	if err != nil || run == 0 {
		return false, err
	}

	go s.run(source, run, done)
	return true, nil
}

// Present shows a sync that has run longer than InventorySourceStaleAfter as failed
func (s *InventorySourceService) Present(source db.InventorySource) db.InventorySource {
	if source.Status == db.InventorySourceSyncing && source.SyncStarted != nil &&
		source.SyncStarted.Before(s.now().Add(-db.InventorySourceStaleAfter)) {
		source.Status = db.InventorySourceFailed
		source.LastError = db.InventorySourceInterrupted
	}

	return source
}

func (s *InventorySourceService) run(source db.InventorySource, run int, done func(db.InventorySyncResult, error)) {
	var result db.InventorySyncResult
	var err error

	defer func() {
		if r := recover(); r != nil {
			log.WithField("source_id", source.ID).Errorf("inventory source sync panicked: %v\n%s", r, debug.Stack())
			result, err = db.InventorySyncResult{}, errors.New("sync failed unexpectedly")
		}

		// A success is stored with the hosts
		if err != nil {
			failErr := s.store.FailInventorySourceSync(source.ProjectID, source.InventoryID, source.ID, run, s.now(), sanitizeSourceError(err.Error()))
			if failErr != nil {
				log.WithError(failErr).WithField("source_id", source.ID).Error("failed to record the inventory source sync")
			}
		}

		if done != nil {
			done(result, err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	select {
	case s.slots <- struct{}{}:
		defer func() { <-s.slots }()
	case <-ctx.Done():
		err = errors.New("too many syncs are running, try again later")
		return
	}

	// An edit between the request and the start changed the path or the repository, later edits wait for the end
	saved, readErr := s.store.GetInventorySource(source.ProjectID, source.InventoryID, source.ID)
	if err = readErr; err != nil {
		return
	}
	source = saved

	var doc db.InventoryDocument
	if source.Type == db.InventorySourceGit {
		doc, err = s.git.Fetch(ctx, source)
	} else {
		doc, err = proServer.FetchInventorySource(ctx, s.store, s.secrets, source)
	}
	if err != nil {
		return
	}

	// The sync is shown as interrupted by now, it must not change the inventory after that
	if ctx.Err() != nil {
		err = errors.New(db.InventorySourceInterrupted)
		return
	}

	result, err = s.store.ApplyInventoryDocument(source.ProjectID, source.InventoryID, source.ID, run, doc)
}
