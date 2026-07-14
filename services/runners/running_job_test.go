package runners

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/stretchr/testify/assert"
)

// newTestRunningJob wires a runningJob to a LocalJob whose Logger points back at
// the runningJob, mirroring production. This matters because LocalJob.SetStatus
// calls Logger.SetStatus, so runningJob.SetStatus re-enters itself; the test
// exercises that path.
func newTestRunningJob(id int) *runningJob {
	lj := &tasks.LocalExecutor{Task: db.Task{ID: id}}
	rj := &runningJob{job: lj}
	lj.Logger = rj
	return rj
}

func TestRunningJob_SetStatusReentrant(t *testing.T) {
	rj := newTestRunningJob(1)

	rj.SetStatus(task_logger.TaskRunningStatus)
	assert.Equal(t, task_logger.TaskRunningStatus, rj.getStatus())

	// Setting the same status again must be a no-op and must not deadlock.
	rj.SetStatus(task_logger.TaskRunningStatus)
	assert.Equal(t, task_logger.TaskRunningStatus, rj.getStatus())

	rj.SetStatus(task_logger.TaskSuccessStatus)
	assert.Equal(t, task_logger.TaskSuccessStatus, rj.getStatus())
}

func TestRunningJob_StatusListenerInvoked(t *testing.T) {
	rj := newTestRunningJob(1)

	var got []task_logger.TaskStatus
	rj.AddStatusListener(func(s task_logger.TaskStatus) {
		got = append(got, s)
	})

	rj.SetStatus(task_logger.TaskRunningStatus)
	rj.SetStatus(task_logger.TaskRunningStatus) // duplicate, must not notify
	rj.SetStatus(task_logger.TaskSuccessStatus)

	assert.Equal(t, []task_logger.TaskStatus{
		task_logger.TaskRunningStatus,
		task_logger.TaskSuccessStatus,
	}, got)
}

func TestRunningJob_AckLogRecords(t *testing.T) {
	tests := []struct {
		name        string
		records     int
		sent        int
		wantPending int
	}{
		{"ack subset", 5, 3, 2},
		{"ack all", 4, 4, 0},
		{"ack more than present", 2, 5, 0},
		{"ack none", 3, 0, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rj := newTestRunningJob(1)
			for i := 0; i < tt.records; i++ {
				rj.Log(fmt.Sprintf("line %d", i))
			}

			pending := rj.ackLogRecords(tt.sent)
			assert.Equal(t, tt.wantPending, pending)

			_, logs, _ := rj.getProgress()
			assert.Len(t, logs, tt.wantPending)
		})
	}
}

func TestRunningJob_GetProgressReturnsCopy(t *testing.T) {
	rj := newTestRunningJob(1)
	rj.Log("a")
	rj.Log("b")

	_, logs, _ := rj.getProgress()
	assert.Len(t, logs, 2)

	// Mutating the returned slice must not corrupt the internal state.
	logs[0].Message = "mutated"

	_, logs2, _ := rj.getProgress()
	assert.Equal(t, "a", logs2[0].Message)
}

// TestRunningJob_ConcurrentAccess hammers every mutator and reader of a single
// runningJob from many goroutines. Run with -race to catch data races on
// status, logRecords, commit and the listener slices.
func TestRunningJob_ConcurrentAccess(t *testing.T) {
	rj := newTestRunningJob(1)

	const workers = 8
	const iterations = 500

	statuses := []task_logger.TaskStatus{
		task_logger.TaskStartingStatus,
		task_logger.TaskRunningStatus,
		task_logger.TaskStoppingStatus,
	}

	var wg sync.WaitGroup
	start := make(chan struct{})

	spawn := func(fn func(i int)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < iterations; i++ {
				fn(i)
			}
		}()
	}

	for w := 0; w < workers; w++ {
		spawn(func(i int) { rj.SetStatus(statuses[i%len(statuses)]) })
		spawn(func(i int) { rj.Log(fmt.Sprintf("line %d", i)) })
		spawn(func(i int) { rj.SetCommit(fmt.Sprintf("hash%d", i), "msg") })
		spawn(func(i int) {
			rj.getStatus()
			_, logs, _ := rj.getProgress()
			if len(logs) > 0 {
				rj.ackLogRecords(1)
			}
		})
	}

	// Listener registration races with log/status emission.
	for w := 0; w < 2; w++ {
		spawn(func(i int) {
			rj.AddLogListener(func(time.Time, string) {})
			rj.AddStatusListener(func(task_logger.TaskStatus) {})
		})
	}

	close(start)
	wg.Wait()
}

func TestRunningJob_finalizeAfterRun_KeepsFinishedStatusOnRunError(t *testing.T) {
	rj := newTestRunningJob(1)
	rj.SetStatus(task_logger.TaskStoppedStatus)

	rj.finalizeAfterRun(errors.New("process killed"))

	assert.Equal(t, task_logger.TaskStoppedStatus, rj.getStatus())
}

func TestRunningJob_finalizeAfterRun_SetsFailOnRunError(t *testing.T) {
	rj := newTestRunningJob(2)
	rj.SetStatus(task_logger.TaskRunningStatus)

	rj.finalizeAfterRun(errors.New("ansible failed"))

	assert.Equal(t, task_logger.TaskFailStatus, rj.getStatus())
}

func TestRunningJob_finalizeAfterRun_SetsSuccessOnCleanReturn(t *testing.T) {
	rj := newTestRunningJob(3)
	rj.SetStatus(task_logger.TaskRunningStatus)

	rj.finalizeAfterRun(nil)

	assert.Equal(t, task_logger.TaskSuccessStatus, rj.getStatus())
}

func TestRunningJob_finalizeAfterRun_StoppingBecomesStopped(t *testing.T) {
	rj := newTestRunningJob(4)
	rj.SetStatus(task_logger.TaskStoppingStatus)

	rj.finalizeAfterRun(nil)

	assert.Equal(t, task_logger.TaskStoppedStatus, rj.getStatus())
}
