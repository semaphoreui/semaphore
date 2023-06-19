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

type job struct {
	job             *tasks.LocalAnsibleJob
	Status          db.TaskStatus
	kind            jobType
	args            []string
	environmentVars []string
	id              int
}

type jobType int

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

const (
	playbook jobType = iota
	galaxy
)

func (j *job) run() {
	switch j.kind {
	case playbook:
		j.job.RunPlaybook(j.args, &j.environmentVars, nil)
	case galaxy:
		j.job.RunGalaxy(j.args)
	default:
		panic("Unknown job type")
	}
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

		case <-ticker.C: // timer 5 seconds
			if len(p.queue) == 0 {
				break
			}

			//get TaskRunner from top of queue
			t := p.queue[0]
			if t.Status == db.TaskFailStatus {
				//delete failed TaskRunner from queue
				p.queue = p.queue[1:]
				log.Info("Task " + strconv.Itoa(t.id) + " removed from queue")
				break
			}

			//if p.blocks(t) {
			//	//move blocked TaskRunner to end of queue
			//	p.queue = append(p.queue[1:], t)
			//	break
			//}

			log.Info("Set resource locker with TaskRunner " + strconv.Itoa(t.id))
			p.resourceLocker <- &resourceLock{lock: true, holder: t}
			//if !t.prepared {
			//	go t.prepareRun()
			//	break
			//}

			go t.run()
			p.queue = p.queue[1:]
			log.Info("Task " + strconv.Itoa(t.id) + " removed from queue")
		}
	}
}

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

	taskRunner := job{}

	p.register <- &taskRunner
}
