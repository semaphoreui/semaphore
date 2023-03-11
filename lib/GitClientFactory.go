package lib

func CreateDefaultGitClient() GitClient {
	return CreateGoGitClient()
}

func CreateGoGitClient() GitClient {
	return GoGitClient{}
}

func CreateCmdGitClient() GitClient {
	return CmdGitClient{}
}
