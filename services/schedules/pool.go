package schedules

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/lib"
	"github.com/ansible-semaphore/semaphore/services/tasks"
	"github.com/robfig/cron/v3"
	"sync"
)

type ScheduleRunner struct {
	schedule db.Schedule
	pool     *SchedulePool
}

func (r ScheduleRunner) Run() {
	if r.schedule.RepositoryID != nil {
		repo, err := r.pool.store.GetRepository(r.schedule.ProjectID, *r.schedule.RepositoryID)
		if err != nil {
			log.Error(err)
			return
		}

		remoteHash, err := lib.GitRepository{
			Logger:     nil,
			TemplateID: r.schedule.TemplateID,
			Repository: repo,
		}.GetLastRemoteCommitHash()

		if err != nil {
			log.Error(err)
			return
		}

		if r.schedule.LastCommitHash != nil && remoteHash == *r.schedule.LastCommitHash {
			return
		}

		err = r.pool.store.SetScheduleCommitHash(r.schedule.ProjectID, r.schedule.ID, remoteHash)
		if err != nil {
			log.Error(err)
			return
		}
	}

	_, err := r.pool.taskPool.AddTask(db.Task{
		TemplateID: r.schedule.TemplateID,
		ProjectID:  r.schedule.ProjectID,
	}, nil, r.schedule.ProjectID)

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
		_, err := p.addRunner(ScheduleRunner{
			schedule: schedule,
			pool:     p,
		})
		if err != nil {
			log.Error(err)
		}
	}
}

func (p *SchedulePool) addRunner(runner ScheduleRunner) (int, error) {
	id, err := p.cron.AddJob(runner.schedule.CronFormat, runner)

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
