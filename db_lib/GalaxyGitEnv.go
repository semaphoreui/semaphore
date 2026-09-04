package db_lib

import (
	"net/url"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// sqQuote quotes s for GIT_CONFIG_PARAMETERS: single-quoted, with embedded
// single quotes written as `'\''`.
func sqQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// galaxyGitEnv gives ansible-galaxy the credential already configured on the
// repository, so roles and collections hosted on that same server install
// without a .netrc or ssh config workaround (SEM-198, GitHub #3677, #3708).
//
// For `scm: git` requirements ansible-galaxy shells out to `git clone`, and
// those clones inherit nothing from Semaphore. Credentials travel in
// GIT_CONFIG_PARAMETERS rather than on the command line so they stay out of
// `ps` output; git reports the pre-rewrite URL, so they stay out of the task
// log too.
func galaxyGitEnv(repo db.Repository) (env []string) {
	// Without this git prompts on /dev/tty and the task hangs instead of failing.
	env = append(env, "GIT_TERMINAL_PROMPT=0")

	if repo.GetType() != db.RepositoryHTTP || repo.SSHKey.Type != db.AccessKeyLoginPassword {
		return
	}

	plain, err := url.Parse(repo.GitURL)
	if err != nil || plain.Host == "" {
		return
	}

	// Scoped to this exact scheme://host[:port] so no other server named in
	// requirements.yml is ever offered the credential.
	plain.Path, plain.RawQuery, plain.Fragment, plain.User = "/", "", "", nil

	withAuth := *plain
	// An empty login means the password is the whole credential, matching
	// Repository.GetGitURL.
	if login := repo.SSHKey.LoginPassword.Login; login == "" {
		withAuth.User = url.User(repo.SSHKey.LoginPassword.Password)
	} else {
		withAuth.User = url.UserPassword(login, repo.SSHKey.LoginPassword.Password)
	}

	return append(env, "GIT_CONFIG_PARAMETERS="+sqQuote(
		"url."+withAuth.String()+".insteadOf="+plain.String()))
}
