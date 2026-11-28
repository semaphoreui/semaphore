package db_lib

import "github.com/semaphoreui/semaphore/util"

// CreateDefaultGitClient returns the configured git client, which hands
// Subversion repositories over to the svn client.
func CreateDefaultGitClient(keyInstaller AccessKeyInstaller) GitClient {
	return repositoryTypeClient{
		git: createConfiguredGitClient(keyInstaller),
		svn: CreateSvnClient(keyInstaller),
	}
}

func createConfiguredGitClient(keyInstaller AccessKeyInstaller) GitClient {
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
