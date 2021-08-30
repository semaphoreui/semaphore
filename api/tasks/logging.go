package tasks

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/util"
)

func (t *task) log(msg string) {
	now := time.Now()

	for _, user := range t.users {
		b, err := json.Marshal(&map[string]interface{}{
			"type":       "log",
			"output":     msg,
			"time":       now,
			"task_id":    t.task.ID,
			"project_id": t.projectID,
		})

		util.LogPanic(err)

		sockets.Message(user, b)
	}

	pool.logger <- logRecord{
		task: t,
		output: msg,
		time: now,
	}
}

// Readln reads from the pipe
func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix = true
		err      error
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func (t *task) logPipe(reader *bufio.Reader) {

	line, err := Readln(reader)
	for err == nil {
		t.log(line)
		line, err = Readln(reader)
	}

	if err != nil && err.Error() != "EOF" {
		//don't panic on this errors, sometimes it throw not dangerous "read |0: file already closed" error
		util.LogWarningWithFields(err, log.Fields{"error": "Failed to read task output"})
	}

}

func (t *task) logCmd(cmd *exec.Cmd) {
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	go t.logPipe(bufio.NewReader(stderr))
	go t.logPipe(bufio.NewReader(stdout))
}

func (t *task) panicOnError(err error, msg string) {
	if err != nil {
		t.log(msg)
		util.LogPanicWithFields(err, log.Fields{"error": msg})
	}
}
