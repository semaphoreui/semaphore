package lib

func CreateDefaultGitClient() GitClient {
	return CreateCmdGitClient()
}

func CreateGoGitClient() GitClient {
	return GoGitClient{}
}

func CreateCmdGitClient() GitClient {
	return CmdGitClient{}
}
