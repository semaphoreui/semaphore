package schedules

import (
	log "github.com/Sirupsen/logrus"
	"github.com/neo1908/semaphore/api/tasks"
	"github.com/neo1908/semaphore/db"
	"github.com/robfig/cron/v3"
	"sync"
)

type ScheduleRunner struct {
	Store    db.Store
	Schedule db.Schedule
}

func (r ScheduleRunner) Run() {
	_, err := tasks.AddTaskToPool(r.Store, db.Task{
		TemplateID: r.Schedule.TemplateID,
		ProjectID:  r.Schedule.ProjectID,
	}, nil, r.Schedule.ProjectID)
	if err != nil {
		log.Error(err)
	}
}

type SchedulePool struct {
	cron   *cron.Cron
	locker sync.Locker
}

func (p *SchedulePool) init() {
	p.cron = cron.New()
	p.locker = &sync.Mutex{}
}

func (p *SchedulePool) Refresh(d db.Store) {
	defer p.locker.Unlock()

	schedules, err := d.GetSchedules()

	if err != nil {
		log.Error(err)
		return
	}

	p.locker.Lock()
	p.clear()
	for _, schedule := range schedules {
		_, err := p.addRunner(ScheduleRunner{
			Store:    d,
			Schedule: schedule,
		})
		if err != nil {
			log.Error(err)
		}
	}
}

func (p *SchedulePool) addRunner(runner ScheduleRunner) (int, error) {
	id, err := p.cron.AddJob(runner.Schedule.CronFormat, runner)

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

func CreateSchedulePool(d db.Store) (pool SchedulePool) {
	pool.init()
	pool.Refresh(d)
	return
}

func ValidateCronFormat(cronFormat string) error {
	_, err := cron.ParseStandard(cronFormat)
	return err
}