//
// Runner's job pool. NOT SERVER!!!
// Runner gets jobs from the server and put them to this pool.
//

package runners

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/services/tasks"
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
	job             *tasks.AnsibleJobRunner
	Status          db.TaskStatus
	args            []string
	environmentVars []string
	id              int
}

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
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

			go t.job.Run()
			p.queue = p.queue[1:]
			log.Info("Task " + strconv.Itoa(t.id) + " removed from queue")
		}
	}
}

// checkNewJobs tries to find runner to queued jobs
func (p *JobPool) checkNewJobs() {
	client := &http.Client{}
	url := "https://example.com"
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

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	taskRunner := job{
		job: &tasks.AnsibleJobRunner{
			// TODO: fields
		},
	}

	p.register <- &taskRunner
}
