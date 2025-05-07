package stage_parsers

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
)

type AnsibleResultStageParser struct{}

const ansibleResultMaker = "PLAY RECAP *****************************************"

func (p AnsibleResultStageParser) NeedParse() bool {
	return true
}

func (p AnsibleResultStageParser) IsStart(currentStage *db.TaskStage, output db.TaskOutput) bool {
	if currentStage == nil {
		return false
	}

	if currentStage.Type != db.TaskStageRunning {
		return false
	}

	return strings.HasPrefix(output.Output, ansibleResultMaker)
}

func (p AnsibleResultStageParser) IsEnd(currentStage *db.TaskStage, output db.TaskOutput) bool {
	if currentStage == nil {
		return false
	}

	if currentStage.Type != db.TaskStagePrintResult {
		return false
	}

	return strings.TrimSpace(output.Output) == ""
}

type AnsibleResultHost struct {
	Host        string `json:"host"`
	Ok          int    `json:"ok"`
	Changed     int    `json:"changed"`
	Unreachable int    `json:"unreachable"`
	Failed      int    `json:"failed"`
	Skipped     int    `json:"skipped"`
	Rescued     int    `json:"rescued"`
	Ignored     int    `json:"ignored"`
}

var ansibleResultHostRE = regexp.MustCompile(
	`^([^\s]+)\s*:\s*` +
		`ok=(\d+)\s+` +
		`changed=(\d+)\s+` +
		`unreachable=(\d+)\s+` +
		`failed=(\d+)\s+` +
		`skipped=(\d+)\s+` +
		`rescued=(\d+)\s+` +
		`ignored=(\d+)$`)

func toInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
	}
	return v
}

func (p AnsibleResultStageParser) Parse(outputs []db.TaskOutput) (res map[string]any, err error) {

	hosts := make([]AnsibleResultHost, 0)

	for _, output := range outputs {

		line := util.ClearFromAnsiCodes(strings.TrimSpace(output.Output))

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, ansibleResultMaker) {
			continue
		}

		m := ansibleResultHostRE.FindStringSubmatch(line)
		if m == nil {
			log.WithFields(log.Fields{
				"task_id": output.TaskID,
			}).Warnf("invalid ansible result host: %s", line)
			continue
		}

		hosts = append(hosts, AnsibleResultHost{
			Host:        m[1],
			Ok:          toInt(m[2]),
			Changed:     toInt(m[3]),
			Unreachable: toInt(m[4]),
			Failed:      toInt(m[5]),
			Skipped:     toInt(m[6]),
			Rescued:     toInt(m[7]),
			Ignored:     toInt(m[8]),
		})
	}

	res = make(map[string]any)
	res["hosts"] = hosts

	return
}

type AnsibleRunningStageParser struct{}

func (p AnsibleRunningStageParser) NeedParse() bool {
	return false
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

func (p AnsibleRunningStageParser) Parse(outputs []db.TaskOutput) (map[string]any, error) {
	// Implement the parsing logic for Ansible results
	return nil, nil
}
