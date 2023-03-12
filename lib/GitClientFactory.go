package lib

import "github.com/ansible-semaphore/semaphore/util"

func CreateDefaultGitClient() GitClient {
	if util.Config.UseExternalGit {
		return CreateCmdGitClient()
	}

	return CreateGoGitClient()
}

func CreateGoGitClient() GitClient {
	return GoGitClient{}
}

func CreateCmdGitClient() GitClient {
	return CmdGitClient{}
}
