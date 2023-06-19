package tasks

import "os"

type RemoteAnsibleJob struct {
	runner *RemoteRunner
}

func (j *RemoteAnsibleJob) RunPlaybook(args []string, environmentVars *[]string, cb func(*os.Process)) error {
	// TODO: put task to the queue
	return nil
}

func (j *RemoteAnsibleJob) RunGalaxy(args []string) error {
	// TODO: put task to the queue

	return nil
}
