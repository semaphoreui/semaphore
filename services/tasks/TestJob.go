package tasks

import "time"

type TestJob struct {
	DurationMs int
	Fail       error
	Done       chan bool
}

func (job *TestJob) Run(username string, incomingVersion *string, alias string) error {
	defer close(job.Done)

	if job.DurationMs > 0 {
		time.Sleep(time.Duration(job.DurationMs) * time.Millisecond)
	}

	return job.Fail
}

func (t *TestJob) Kill() {
}
