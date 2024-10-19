package schedules

import (
	"strconv"
	"sync"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db_lib"
	"github.com/ansible-semaphore/semaphore/services/tasks"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

type ScheduleRunner struct {
	projectID  int
	scheduleID int
	pool       *SchedulePool
}

func (r ScheduleRunner) tryUpdateScheduleCommitHash(schedule db.Schedule) (updated bool, err error) {
	repo, err := r.pool.store.GetRepository(schedule.ProjectID, *schedule.RepositoryID)
	if err != nil {
		return
	}

	err = repo.SSHKey.DeserializeSecret()
	if err != nil {
		return
	}

	remoteHash, err := db_lib.GitRepository{
		Logger:     nil,
		TemplateID: schedule.TemplateID,
		Repository: repo,
		Client:     db_lib.CreateDefaultGitClient(),
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
		log.Error(err)
		return
	}

	if schedule.RepositoryID != nil {
		var updated bool
		updated, err = r.tryUpdateScheduleCommitHash(schedule)
		if err != nil {
			log.Error(err)
			return
		}
		if !updated {
			return
		}
	}

	_, err = r.pool.taskPool.AddTask(db.Task{
		TemplateID: schedule.TemplateID,
		ProjectID:  schedule.ProjectID,
	}, nil, schedule.ProjectID)

	if err != nil {
		log.Error(err)
	}
}

type SchedulePool struct {
	cron     *cron.Cron
	locker   sync.Locker
	store    db.Store
	taskPool *tasks.TaskPool
}

func (p *SchedulePool) init() {
	p.cron = cron.New()
	p.locker = &sync.Mutex{}
}

func (p *SchedulePool) Refresh() {
	defer p.locker.Unlock()

	schedules, err := p.store.GetSchedules()

	if err != nil {
		log.Error(err)
		return
	}

	p.locker.Lock()
	p.clear()
	for _, schedule := range schedules {
		if schedule.RepositoryID == nil && !schedule.Active {
			continue
		}

		_, err := p.addRunner(ScheduleRunner{
			projectID:  schedule.ProjectID,
			scheduleID: schedule.ID,
			pool:       p,
		}, schedule.CronFormat)
		if err != nil {
			log.Error(err)
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
	defer p.locker.Unlock()
	p.locker.Lock()
	p.cron.Stop()
	p.clear()
	p.cron = nil
}

func CreateSchedulePool(store db.Store, taskPool *tasks.TaskPool) SchedulePool {
	pool := SchedulePool{
		store:    store,
		taskPool: taskPool,
	}
	pool.init()
	pool.Refresh()
	return pool
}

func ValidateCronFormat(cronFormat string) error {
	_, err := cron.ParseStandard(cronFormat)
	return err
}
