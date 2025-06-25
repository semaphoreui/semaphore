package stage_parsers

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"regexp"
	"strings"
)

type AnsibleRunningStageParserFailedTask struct {
	Task   string
	Host   string
	Answer string
}

type AnsibleRunningStageParserState struct {
	CurrentTask       string
	CurrentFailedHost string
	CurrentHostAnswer string
	FailedTasks       []AnsibleRunningStageParserFailedTask
	Tasks             int
}

type AnsibleRunningStageParser struct {
	state *AnsibleRunningStageParserState
}

func (p AnsibleRunningStageParser) NeedParse() bool {
	return true
}

func (p AnsibleRunningStageParser) IsStart(currentStage *db.TaskStage, output db.TaskOutput) bool {

	if currentStage == nil {
		return false
	}

	if currentStage.Type != db.TaskStageInit {
		return false
	}

	return strings.HasPrefix(output.Output, "PLAY [")
}

func (p AnsibleRunningStageParser) IsEnd(currentStage *db.TaskStage, output db.TaskOutput) bool {
	return false
}

const ansibleTaskMaker = "TASK ["
const failedTaskMaker = "fatal: ["

var newTaskMakerRE = regexp.MustCompile(`^\w+: \[`)
var fatalTaskRE = regexp.MustCompile(`^fatal: \[([^]]+)]: FAILED! => (.*)$`)

func (p AnsibleRunningStageParser) Parse(currentStage *db.TaskStage, output db.TaskOutput, store db.Store, projectID int) (ok bool, err error) {

	if currentStage == nil {
		return
	}

	if currentStage.Type != db.TaskStageRunning {
		return
	}

	ok = true

	line := util.ClearFromAnsiCodes(strings.TrimSpace(output.Output))

	if strings.HasPrefix(line, ansibleTaskMaker) {
		p.state.Tasks++

		if p.state.CurrentFailedHost != "" {
			err = store.CreateAnsibleTaskError(db.AnsibleTaskError{
				TaskID:    currentStage.TaskID,
				ProjectID: projectID,
				Host:      p.state.CurrentFailedHost,
				Task:      p.state.CurrentTask,
				Error:     p.state.CurrentHostAnswer,
			})

			if err != nil {
				return
			}
		}

		end := strings.Index(line, "]")
		p.state.CurrentTask = line[len(ansibleTaskMaker):end]
		p.state.CurrentFailedHost = ""
		p.state.CurrentHostAnswer = ""

	} else if strings.HasPrefix(line, failedTaskMaker) {

		if p.state.CurrentFailedHost != "" {
			err = store.CreateAnsibleTaskError(db.AnsibleTaskError{
				TaskID:    currentStage.TaskID,
				ProjectID: projectID,
				Host:      p.state.CurrentFailedHost,
				Task:      p.state.CurrentTask,
				Error:     p.state.CurrentHostAnswer,
			})

			if err != nil {
				return
			}
		}

		m := fatalTaskRE.FindStringSubmatch(line)
		if len(m) > 0 {
			host := strings.TrimSpace(m[1])
			msg := strings.TrimSpace(m[2])
			p.state.CurrentFailedHost = host
			p.state.CurrentHostAnswer = msg
		} else {
			end := strings.Index(line, "]")
			start := len(failedTaskMaker)
			p.state.CurrentFailedHost = line[start:end]
			p.state.CurrentHostAnswer = ""
		}

	} else if p.state.CurrentFailedHost != "" {
		m := newTaskMakerRE.FindStringSubmatch(line)
		if line == "" || len(m) > 0 {

			if p.state.CurrentFailedHost != "" {
				err = store.CreateAnsibleTaskError(db.AnsibleTaskError{
					TaskID:    currentStage.TaskID,
					ProjectID: projectID,
					Host:      p.state.CurrentFailedHost,
					Task:      p.state.CurrentTask,
					Error:     p.state.CurrentHostAnswer,
				})

				if err != nil {
					return
				}
			}
			p.state.CurrentFailedHost = ""
			p.state.CurrentHostAnswer = ""

		} else {
			p.state.CurrentHostAnswer += "\n" + line
		}
	}

	return
}

func (p AnsibleRunningStageParser) Result() (res map[string]any) {

	res = make(map[string]any)
	failed := make(map[string]any)
	res["failed"] = failed

	for _, task := range p.state.FailedTasks {
		failed[task.Host] = map[string]any{
			"task":   task.Task,
			"host":   task.Host,
			"answer": task.Answer,
		}
	}

	res["tasks"] = p.state.Tasks

	return
}

func (p AnsibleRunningStageParser) State() any {

	return p.state
}
