package db_lib

import "github.com/Digital-Data-Co/semaphore/util"

func CreateDefaultGitClient(keyInstaller AccessKeyInstaller) GitClient {
	switch util.Config.GitClientId {
	case util.GoGitClientId:
		return CreateGoGitClient(keyInstaller)
	case util.CmdGitClientId:
		return CreateCmdGitClient(keyInstaller)
	default:
		return CreateCmdGitClient(keyInstaller)
	}
}

func CreateGoGitClient(keyInstaller AccessKeyInstaller) GitClient {
	return GoGitClient{
		keyInstaller: keyInstaller,
	}
}

func CreateCmdGitClient(keyInstaller AccessKeyInstaller) GitClient {
	return CmdGitClient{
		keyInstaller: keyInstaller,
	}
}
