package tasks

import "os"

type RemoteAnsibleJob struct {
	runner *RemoteRunner
}

func (j *RemoteAnsibleJob) RunPlaybook(args []string, environmentVars *[]string, cb func(*os.Process)) error {
	// TODO: upload required data to the runner
	// TODO: initiate execution
	// TODO: receiving data

	return nil
}

func (j *RemoteAnsibleJob) RunGalaxy(args []string) error {
	//return j.playbook.RunGalaxy(args)

	return nil
}
