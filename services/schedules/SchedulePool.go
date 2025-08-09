package schedules

import (
	"strconv"
	"sync"
	"time"

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

	tpl, err := r.pool.store.GetTemplate(schedule.ProjectID, schedule.TemplateID)
	if err != nil {
		log.Error(err)
		return
	}

	task := schedule.TaskParams.CreateTask(schedule.TemplateID)
	task.ScheduleID = &schedule.ID

	_, err = r.pool.taskPool.AddTask(
		task,
		nil,
		"",
		schedule.ProjectID,
		tpl.App.NeedTaskAlias(),
	)

	if err != nil {
		log.Error(err)
	}
}

type SchedulePool struct {
	cron              *cron.Cron
	locker            sync.Locker
	store             db.Store
	taskPool          *tasks.TaskPool
	encryptionService server.AccessKeyEncryptionService
	keyInstaller      db_lib.AccessKeyInstaller
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
		log.Error(err)
		return
	}

	p.locker.Lock()
	defer p.locker.Unlock()

	p.clear()
	for _, schedule := range schedules {
		if schedule.RepositoryID == nil && !schedule.Active {
			continue
		}

		_, err = p.addRunner(CreateScheduleRunner(
			schedule.ProjectID,
			schedule.ID,
			p,
			p.encryptionService,
			p.keyInstaller,
		), schedule.CronFormat)

		if err != nil {
			log.WithError(err).WithFields(log.Fields{
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
