package db_lib

// repositoryTypeClient picks the client for each call from the repository it
// is given: the svn client for a Subversion URL, the git client otherwise. The
// URL scheme is the only routing key, so git and Subversion repositories live
// on one server without a setting or a schema change.
type repositoryTypeClient struct {
	git GitClient
	svn GitClient
}

func (c repositoryTypeClient) client(r GitRepository) GitClient {
	if r.Repository.IsSubversion() {
		return c.svn
	}
	return c.git
}

func (c repositoryTypeClient) Clone(r GitRepository) error {
	return c.client(r).Clone(r)
}

func (c repositoryTypeClient) Pull(r GitRepository) error {
	return c.client(r).Pull(r)
}

func (c repositoryTypeClient) Checkout(r GitRepository, target string) error {
	return c.client(r).Checkout(r, target)
}

func (c repositoryTypeClient) CanBePulled(r GitRepository) bool {
	return c.client(r).CanBePulled(r)
}

func (c repositoryTypeClient) GetLastCommitMessage(r GitRepository) (string, error) {
	return c.client(r).GetLastCommitMessage(r)
}

func (c repositoryTypeClient) GetLastCommitHash(r GitRepository) (string, error) {
	return c.client(r).GetLastCommitHash(r)
}

func (c repositoryTypeClient) GetLastRemoteCommitHash(r GitRepository) (string, error) {
	return c.client(r).GetLastRemoteCommitHash(r)
}

func (c repositoryTypeClient) GetRemoteBranches(r GitRepository) ([]string, error) {
	return c.client(r).GetRemoteBranches(r)
}
