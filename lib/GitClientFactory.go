package lib

import "github.com/ansible-semaphore/semaphore/util"

func CreateDefaultGitClient() GitClient {
	switch util.Config.GitClient {
	case "go_git":
		return CreateGoGitClient()
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
