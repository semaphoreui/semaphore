package schedules

import (
	"strconv"
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/semaphoreui/semaphore/util"

	"github.com/robfig/cron/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/services/tasks"
	log "github.com/sirupsen/logrus"
)

type ScheduleRunner struct {
	projectID         int
	scheduleID        int
	pool              *SchedulePool
	encryptionService server.AccessKeyEncryptionService
	keyInstaller      db_lib.AccessKeyInstaller
}

type oneTimeSchedule struct {
	runAt time.Time
	ran   bool
}

func (s *oneTimeSchedule) Next(t time.Time) time.Time {
	if s.ran {
		return time.Time{}
	}

	if !t.Before(s.runAt) {
		s.ran = true
		return time.Time{}
	}

	return s.runAt
}

func CreateScheduleRunner(
	projectID int,
	scheduleID int,
	pool *SchedulePool,
	encryptionService server.AccessKeyEncryptionService,
	keyInstaller db_lib.AccessKeyInstaller,
) ScheduleRunner {
	return ScheduleRunner{
		projectID:         projectID,
		scheduleID:        scheduleID,
		pool:              pool,
		encryptionService: encryptionService,
		keyInstaller:      keyInstaller,
	}
}

func (r ScheduleRunner) tryUpdateScheduleCommitHash(schedule db.Schedule) (updated bool, err error) {
	repo, err := r.pool.store.GetRepository(schedule.ProjectID, *schedule.RepositoryID)
	if err != nil {
		return
	}

	err = r.pool.encryptionService.DeserializeSecret(&repo.SSHKey)
	if err != nil {
		return
	}

	remoteHash, err := db_lib.GitRepository{
		Logger:     nil,
		TemplateID: schedule.TemplateID,
		Repository: repo,
		Client:     db_lib.CreateDefaultGitClient(r.keyInstaller),
	}.GetLastRemoteCommitHash()

	if err != nil {
		return
	}

	if schedule.LastCommitHash != nil && remoteHash == *schedule.LastCommitHash {
		return
	}

	err = r.pool.store.SetScheduleCommitHash(schedule.ProjectID, schedule.ID, remoteHash)
	if err != nil {
		return
	}

	updated = true
	return
}

func (r ScheduleRunner) Run() {
	if !r.pool.store.PermanentConnection() {
		r.pool.store.Connect("schedule " + strconv.Itoa(r.scheduleID))
		defer r.pool.store.Close("schedule " + strconv.Itoa(r.scheduleID))
	}

	schedule, err := r.pool.store.GetSchedule(r.projectID, r.scheduleID)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context":     common_errors.GetErrorContext(),
			"project_id":  r.projectID,
			"schedule_id": r.scheduleID,
		}).Error("failed to get schedule")
		return
	}

	scheduleType := schedule.Type
	if scheduleType == "" {
		scheduleType = db.ScheduleTypeCron
	}

	if schedule.RepositoryID != nil {
		var updated bool
		updated, err = r.tryUpdateScheduleCommitHash(schedule)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context":     common_errors.GetErrorContext(),
				"project_id":  r.projectID,
				"schedule_id": r.scheduleID,
			}).Error("failed to update schedule commit hash")
			return
		}
		if !updated {
			return
		}
	}

	tpl, err := r.pool.store.GetTemplate(schedule.ProjectID, schedule.TemplateID)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context":     common_errors.GetErrorContext(),
			"project_id":  schedule.ProjectID,
			"schedule_id": schedule.ID,
			"template_id": schedule.TemplateID,
		}).Error("failed to get template")
		return
	}

	// In HA mode, ensure only one node fires this schedule occurrence.
	if r.pool.dedup != nil && !r.pool.dedup.TryLockExecution(r.scheduleID) {
		log.WithFields(log.Fields{
			"project_id":  r.projectID,
			"schedule_id": r.scheduleID,
		}).Debug("schedule already executed by another node")
		// For one-time schedules the winning node deactivates/deletes
		// the schedule in the DB after execution. Refresh so this
		// node's cron picks up that change and drops the stale entry.
		if scheduleType == db.ScheduleTypeRunAt {
			r.pool.Refresh()
		}
		return
	}

	var task db.Task
	if schedule.TaskParams != nil {
		task = schedule.TaskParams.CreateTask(schedule.TemplateID)
	} else {
		task = db.Task{
			ProjectID:  schedule.ProjectID,
			TemplateID: schedule.TemplateID,
		}
	}
	task.ScheduleID = &schedule.ID

	_, err = r.pool.taskPool.AddTask(
		task,
		nil,
		"",
		schedule.ProjectID,
		tpl.App.NeedTaskAlias(),
	)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context":     common_errors.GetErrorContext(),
			"project_id":  schedule.ProjectID,
			"schedule_id": schedule.ID,
			"template_id": schedule.TemplateID,
		}).Error("failed to add task")
	}

	// For "RunAt" schedules, the schedule should only trigger once at the specified time and be deactivated afterwards.
	// Calling Refresh here ensures that after the job has fired, the pool reloads the active schedules
	// from the database (where this run-at schedule may now be disabled) so it is not executed again.
	if scheduleType == db.ScheduleTypeRunAt {
		r.pool.Refresh()
	}
}

// ScheduleDeduplicator prevents the same schedule from being executed on
// multiple nodes simultaneously in an HA cluster. When configured, each
// ScheduleRunner calls TryLockExecution before creating a task.
//
// The deduplication lock is intended to cover a *single execution attempt*
// of a schedule occurrence: a node should acquire the lock immediately
// before creating a task and release it once the attempt has either
// completed or failed. Implementations are free to choose the underlying
// mechanism (in‑memory, database, distributed store, etc.), but they should
// be robust to node failures and process restarts (for example by using
// leases with automatic expiry).
//
// Callers MUST treat the lock as advisory and best‑effort: if the
// implementation becomes unavailable or releases the lock early, at‑most‑once
// execution across the cluster is not guaranteed.
type ScheduleDeduplicator interface {
	// TryLockExecution attempts to acquire an execution lock for the given
	// schedule occurrence.
	//
	// Lock duration:
	//   - The lock is expected to remain held for the duration of the current
	//     schedule execution attempt (from just before task creation until
	//     the attempt finishes or fails).
	//   - Implementations will typically release the lock explicitly when the
	//     attempt ends and/or rely on a lease with automatic expiry to avoid
	//     permanent deadlocks.
	//
	// Timeouts and crash behavior:
	//   - If the node that acquired the lock crashes or loses connectivity,
	//     the behavior is implementation‑specific. Recommended practice is to
	//     use a finite TTL/lease so that the lock eventually expires and
	//     future executions can proceed.
	//
	// Idempotency:
	//   - TryLockExecution may be called multiple times for the same
	//     scheduleID (for example, after retries or rescheduling). The
	//     implementation SHOULD behave idempotently such that, for a single
	//     schedule occurrence, at most one call across the cluster returns
	//     true.
	//
	// Returns true if this node successfully acquired the lock and should
	// execute the schedule, and false otherwise.
	TryLockExecution(scheduleID int) bool
}

type SchedulePool struct {
	cron              *cron.Cron
	locker            sync.Locker
	dedup             ScheduleDeduplicator
	store             db.Store
	taskPool          *tasks.TaskPool
	encryptionService server.AccessKeyEncryptionService
	keyInstaller      db_lib.AccessKeyInstaller
}

// SetDeduplicator configures a distributed schedule deduplicator for HA mode.
// When set, only one node in the cluster fires each schedule occurrence.
func (p *SchedulePool) SetDeduplicator(d ScheduleDeduplicator) {
	p.dedup = d
}

func (p *SchedulePool) init() {
	loc, err := time.LoadLocation(util.Config.Schedule.Timezone)
	if err != nil {
		panic(err)
	}
	p.cron = cron.New(cron.WithLocation(loc))
	p.locker = &sync.Mutex{}
}

func (p *SchedulePool) Refresh() {

	schedules, err := p.store.GetSchedules()

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": common_errors.GetErrorContext(),
		}).Error("failed to get schedules")
		return
	}

	p.locker.Lock()
	defer p.locker.Unlock()

	p.clear()
	now := time.Now().In(p.cron.Location())
	for _, schedule := range schedules {
		scheduleType := schedule.Type
		if scheduleType == "" {
			scheduleType = db.ScheduleTypeCron
		}

		if schedule.RepositoryID == nil && !schedule.Active {
			continue
		}

		runner := CreateScheduleRunner(
			schedule.ProjectID,
			schedule.ID,
			p,
			p.encryptionService,
			p.keyInstaller,
		)

		switch scheduleType {
		case db.ScheduleTypeRunAt:
			if schedule.RunAt == nil {
				log.WithFields(log.Fields{
					"project_id":  schedule.ProjectID,
					"schedule_id": schedule.ID,
				}).Warn("run_at schedule has no run_at value")
				continue
			}

			runAt := schedule.RunAt.In(p.cron.Location())

			if !runAt.After(now) {
				if schedule.DeleteAfterRun {
					err = p.store.DeleteSchedule(schedule.ProjectID, schedule.ID)
					if err != nil {
						log.WithError(err).WithFields(log.Fields{
							"context":     common_errors.GetErrorContext(),
							"project_id":  schedule.ProjectID,
							"schedule_id": schedule.ID,
						}).Warn("failed to delete past run_at schedule")
					}
				} else if schedule.Active {
					err = p.store.SetScheduleActive(schedule.ProjectID, schedule.ID, false)
					if err != nil {
						log.WithError(err).WithFields(log.Fields{
							"context":     common_errors.GetErrorContext(),
							"project_id":  schedule.ProjectID,
							"schedule_id": schedule.ID,
						}).Warn("failed to deactivate past run_at schedule")
					}
				}
				continue
			}

			_, err = p.addOneTimeRunner(runner, runAt)
		case db.ScheduleTypeCron:
			if schedule.CronFormat == "" {
				continue
			}

			_, err = p.addRunner(runner, schedule.CronFormat)
		default:
			log.WithFields(log.Fields{
				"project_id":  schedule.ProjectID,
				"schedule_id": schedule.ID,
				"type":        schedule.Type,
			}).Warn("schedule has unsupported type")
			continue
		}

		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context":     common_errors.GetErrorContext(),
				"project_id":  schedule.ProjectID,
				"schedule_id": schedule.ID,
			}).Errorf("failed to add schedule")
		}
	}
}

func (p *SchedulePool) addRunner(runner ScheduleRunner, cronFormat string) (int, error) {
	id, err := p.cron.AddJob(cronFormat, runner)

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (p *SchedulePool) addOneTimeRunner(runner ScheduleRunner, runAt time.Time) (int, error) {
	id := p.cron.Schedule(&oneTimeSchedule{runAt: runAt}, runner)

	return int(id), nil
}

func (p *SchedulePool) Run() {
	p.cron.Run()
}

func (p *SchedulePool) clear() {
	runners := p.cron.Entries()
	for _, r := range runners {
		p.cron.Remove(r.ID)
	}
}

func (p *SchedulePool) Destroy() {
	p.locker.Lock()
	defer p.locker.Unlock()
	p.cron.Stop()
	p.clear()
	p.cron = nil
}

func CreateSchedulePool(
	store db.Store,
	taskPool *tasks.TaskPool,
	keyInstaller db_lib.AccessKeyInstaller,
	encryptionService server.AccessKeyEncryptionService,
) SchedulePool {
	pool := SchedulePool{
		store:             store,
		taskPool:          taskPool,
		keyInstaller:      keyInstaller,
		encryptionService: encryptionService,
	}
	pool.init()
	pool.Refresh()
	return pool
}

func ValidateCronFormat(cronFormat string) error {
	_, err := cron.ParseStandard(cronFormat)
	return err
}
