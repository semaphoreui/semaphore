package schedules

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/robfig/cron/v3"
)

type templateRunner struct {
	schedule db.Schedule
}

func (r templateRunner) Run() {
	// TODO: add task to tasks pool
}

type schedulePool struct {
	cron *cron.Cron
	jobs []cron.EntryID
}

func (p *schedulePool) init() {
	p.cron = cron.New()
}

func (p *schedulePool) loadData(d db.Store) {
	schedules, err := d.GetSchedules()

	if err != nil {
		// TODO: log error
		return
	}

	for _, schedule := range schedules {
		err := p.addSchedule(schedule)
		if err != nil {
			// TODO: log error
		}
	}
}

func (p *schedulePool) addSchedule(schedule db.Schedule) error {
	id, err := p.cron.AddJob(schedule.CronFormat, templateRunner{
		schedule: schedule,
	})
	if err != nil {
		return err
	}
	p.jobs = append(p.jobs, id)
	return nil
}

func (p *schedulePool) run() {
	p.cron.Run()
}

var pool = schedulePool{}

// StartRunner begins the schedule pool, used as a goroutine
func StartRunner(d db.Store) {
	pool.init()
	pool.loadData(d)
	pool.run()
}
