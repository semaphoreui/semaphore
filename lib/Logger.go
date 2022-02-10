package lib

import "os/exec"

type Logger interface {
	Log(msg string)
	LogCmd(cmd *exec.Cmd)
}
