//go:build pro

package stage_parsers

import (
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

func writeLogTaskResultLog(state *AnsibleResultStageParserState) {
	if !util.Config.Log.Tasks.Enabled {
		return
	}

	err := util.Config.Log.Tasks.WriteResult(*state)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "parsing",
			"app":     "ansible",
			"stage":   "result",
		}).Error("write result log")
	}
}
