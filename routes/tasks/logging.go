package tasks

import (
	"bufio"
	"encoding/json"
	"os/exec"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/routes/sockets"
)

func (t *task) log(msg string) {
	for _, user := range t.users {
		b, err := json.Marshal(&map[string]interface{}{
			"type":       "log",
			"m":          msg,
			"task_id":    t.task.ID,
			"project_id": t.projectID,
		})

		if err != nil {
			panic(err)
		}

		sockets.Message(user, b)
	}

	go func() {
		_, err := database.Mysql.Exec("insert into task__output set task_id=?, output=?, time=NOW(6)", t.task.ID, msg)
		if err != nil {
			panic(err)
		}
	}()
}

func (t *task) logPipe(scanner *bufio.Scanner) {
	for scanner.Scan() {
		t.log(scanner.Text())
	}
}

func (t *task) logCmd(cmd *exec.Cmd) {
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	go t.logPipe(bufio.NewScanner(stderr))
	go t.logPipe(bufio.NewScanner(stdout))
}
