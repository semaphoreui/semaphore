package tasks

import (
	"bufio"
	"os/exec"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/routes/sockets"
)

func (t *task) log(msg string) {
	for _, user := range t.users {
		sockets.Message(user, []byte(msg))
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
