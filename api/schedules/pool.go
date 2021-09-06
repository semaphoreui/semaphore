package schedules

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/tasks"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/robfig/cron/v3"
)

type ScheduleRunner struct {
	Store    db.Store
	Schedule db.Schedule
}

func (r ScheduleRunner) Run() {
	_, err := tasks.AddTaskToPool(r.Store, db.Task{}, nil, r.Schedule.ProjectID)
	if err != nil {
		log.Error(err)
	}
}

type SchedulePool struct {
	cron *cron.Cron
}

func (p *SchedulePool) init(d db.Store) {
	p.cron = cron.New()

	schedules, err := d.GetSchedules()

	if err != nil {
		log.Error(err)
		return
	}

	for _, schedule := range schedules {
		err := p.AddRunner(ScheduleRunner{
			Store:    d,
			Schedule: schedule,
		})
		if err != nil {
			log.Error(err)
		}
	}
}

func (p *SchedulePool) AddRunner(runner ScheduleRunner) error {
	_, err := p.cron.AddJob(runner.Schedule.CronFormat, runner)
	if err != nil {
		return err
	}
	return nil
}

func (p *SchedulePool) Run() {
	p.cron.Run()
}

func (p *SchedulePool) Destroy() {
	p.cron.Stop()
	runners := p.cron.Entries()
	for _, r := range runners {
		p.cron.Remove(r.ID)
	}
	p.cron = nil
}

func CreateSchedulePool(d db.Store) (pool SchedulePool) {
	pool.init(d)
	return
}
