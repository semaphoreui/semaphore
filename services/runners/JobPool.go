//
// Runner's job pool. NOT SERVER!!!
// Runner gets jobs from the server and put them to this pool.
//

package runners

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/lib"
	"github.com/ansible-semaphore/semaphore/services/tasks"
	"github.com/ansible-semaphore/semaphore/util"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type logRecord struct {
	job    *job
	output string
	time   time.Time
}

type resourceLock struct {
	lock   bool
	holder *job
}

// job presents current job on semaphore server.
type job struct {

	// job presents remote or local job information
	job             *tasks.LocalJob
	Status          db.TaskStatus
	args            []string
	environmentVars []string
	id              int
}

type JobData struct {
	Task        db.Task        `json:"task"`
	Template    db.Template    `json:"template"`
	Inventory   db.Inventory   `json:"inventory"`
	Repository  db.Repository  `json:"repository"`
	Environment db.Environment `json:"environment"`
}

type RunnerState struct {
	CurrentJobs []JobState
	NewJobs     []JobData `json:"new_jobs"`
}

type JobState struct {
	ID     int           `json:"id"`
	Status db.TaskStatus `json:"status"`
}

type LogRecord struct {
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

type RunnerProgress struct {
	Jobs []JobProgress `json:"jobs"`
}

type JobProgress struct {
	ID         int
	Status     db.TaskStatus
	LogRecords []LogRecord
}

type JobPool struct {
	// logger channel used to putting log records to database.
	logger chan logRecord

	// register channel used to put tasks to queue.
	register chan *job

	resourceLocker chan *resourceLock

	logRecords []logRecord

	queue []*job
}

func (p *JobPool) Run() {
	ticker := time.NewTicker(5 * time.Second)

	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case record := <-p.logger: // new log message which should be put to database
			p.logRecords = append(p.logRecords, record)

		case job := <-p.register: // new task created by API or schedule
			p.queue = append(p.queue, job)

		case <-ticker.C: // timer 5 seconds: get task from queue and run it
			if len(p.queue) == 0 {
				break
			}

			t := p.queue[0]
			if t.Status == db.TaskFailStatus {
				//delete failed TaskRunner from queue
				p.queue = p.queue[1:]
				log.Info("Task " + strconv.Itoa(t.id) + " removed from queue")
				break
			}

			log.Info("Set resource locker with TaskRunner " + strconv.Itoa(t.id))
			p.resourceLocker <- &resourceLock{lock: true, holder: t}

			go t.job.Run("", nil)
			p.queue = p.queue[1:]
			log.Info("Task " + strconv.Itoa(t.id) + " removed from queue")
		}
	}
}

func (p *JobPool) sendProgress() {
	client := &http.Client{}

	runnerID := 0 // TODO: read from stored file

	url := util.Config.Runner.ApiURL + "/runners/" + strconv.Itoa(runnerID)

	body := RunnerProgress{
		Jobs: nil,
	}

	jsonBytes, err := json.Marshal(body)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()
}

// checkNewJobs tries to find runner to queued jobs
func (p *JobPool) checkNewJobs() {
	client := &http.Client{}

	runnerID := 0 // TODO: read from stored file

	url := util.Config.Runner.ApiURL + "/runners/" + strconv.Itoa(runnerID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var response RunnerState
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	for _, newJob := range response.NewJobs {
		taskRunner := job{
			job: &tasks.LocalJob{
				Task:        newJob.Task,
				Template:    newJob.Template,
				Inventory:   newJob.Inventory,
				Repository:  newJob.Repository,
				Environment: newJob.Environment,
				//logger:      &jobData,
				Playbook: &lib.AnsiblePlaybook{
					//Logger:     &jobData,
					TemplateID: newJob.Template.ID,
					Repository: newJob.Repository,
				},
			},
		}

		p.register <- &taskRunner
	}

}
