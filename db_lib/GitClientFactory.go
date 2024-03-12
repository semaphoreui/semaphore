package db_lib

import "github.com/ansible-semaphore/semaphore/util"

func CreateDefaultGitClient() GitClient {
	switch util.Config.GitClientId {
	case util.GoGitClientId:
		return CreateGoGitClient()
	case util.CmdGitClientId:
		return CreateCmdGitClient()
	default:
		return CreateCmdGitClient()
	}
}

func CreateGoGitClient() GitClient {
	return GoGitClient{}
}

func CreateCmdGitClient() GitClient {
	return CmdGitClient{}
}
