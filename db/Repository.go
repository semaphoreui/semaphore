package db

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/git"
	"github.com/semaphoreui/semaphore/util"

	log "github.com/sirupsen/logrus"
)

type RepositoryType string

const (
	RepositoryGit   RepositoryType = "git"
	RepositorySSH   RepositoryType = "ssh"
	RepositoryHTTP  RepositoryType = "https"
	RepositoryFile  RepositoryType = "file"
	RepositoryLocal RepositoryType = "local"
)

// Repository is the model for code stored in a git repository
type Repository struct {
	ID        int    `db:"id" json:"id" backup:"-"`
	Name      string `db:"name" json:"name" binding:"required"`
	ProjectID int    `db:"project_id" json:"project_id" backup:"-"`
	GitURL    string `db:"git_url" json:"git_url" binding:"required"`
	GitBranch string `db:"git_branch" json:"git_branch" binding:"required"`
	SSHKeyID  int    `db:"ssh_key_id" json:"ssh_key_id" binding:"required" backup:"-"`

	SSHKey AccessKey `db:"-" json:"-" backup:"-"`
}

func (r Repository) ClearCache() error {
	return util.ClearDir(util.Config.GetProjectTmpDir(r.ProjectID), true, r.getDirNamePrefix())
}

func (r Repository) getDirNamePrefix() string {
	return "repository_" + strconv.Itoa(r.ID) + "_"
}

func (r Repository) GetDirName(templateID int) string {
	return r.getDirNamePrefix() + "template_" + strconv.Itoa(templateID)
}

const branchDirNameMaxLen = 48
const branchDirNameHashBytes = 16

// branchDirName turns a branch name into a path-safe directory suffix. The
// hash distinguishes branches whose readable names would otherwise collide.
func branchDirName(branch string) string {
	sum := sha1.Sum([]byte(branch))

	readable := strings.Map(func(c rune) rune {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' ||
			c == '.' || c == '_' || c == '-' {
			return c
		}
		return '-'
	}, branch)

	if len(readable) > branchDirNameMaxLen {
		readable = readable[:branchDirNameMaxLen]
	}

	if readable == "" {
		return fmt.Sprintf("%x", sum[:branchDirNameHashBytes])
	}

	return readable + "_" + fmt.Sprintf("%x", sum[:branchDirNameHashBytes])
}

// GetCheckoutDirName returns the checkout directory name for this template and
// branch. Different branches must not share a working tree while tasks run.
func (r Repository) GetCheckoutDirName(templateID int) string {
	return r.GetDirName(templateID) + "_" + branchDirName(r.GitBranch)
}

// GetHomePath returns the per-template "home" directory with a "_home" suffix.
// Currently this path is used for home-like directories such as ANSIBLE_HOME so
// that parallel tasks from different templates get isolated home directories
// (preventing concurrent ansible-galaxy writes to the same collections path),
// while keeping these artifacts separate from the repository files.
func (r Repository) GetHomePath(templateID int) string {
	return path.Join(util.Config.GetProjectTmpDir(r.ProjectID), r.GetDirName(templateID)+"_home")
}

// GetInternalPath returns a per-template directory under the project tmp dir for Semaphore-owned
// metadata (e.g. galaxy requirements hashes). It is not a copy of the repository.
func (r Repository) GetInternalPath(templateID int) string {
	return path.Join(util.Config.GetProjectTmpDir(r.ProjectID), r.GetDirName(templateID)+"_internal")
}

// GetFullPath returns the path where the repository source code lives.
// The repository is cloned directly into its branch-specific checkout
// directory (e.g. repository_15_template_114_main_1a2b3c4d).
func (r Repository) GetFullPath(templateID int) string {
	if r.GetType() == RepositoryLocal {
		return r.GetGitURL(false)
	}
	return path.Join(util.Config.GetProjectTmpDir(r.ProjectID), r.GetCheckoutDirName(templateID))
}

// GetGitURL returns the URL git is invoked with. With embedCredentials set,
// the repository's login/password access key is embedded in the userinfo,
// percent encoded, so credentials containing "@", ":", "#" or "%" survive
// intact. Otherwise the URL is returned exactly as configured. It may still
// carry userinfo typed into the URL itself, which go-git uses as basic auth,
// so it is not safe to log in either mode. Use GetRedactedGitURL for that.
//
// A URL net/url cannot parse is returned with its userinfo removed in both
// modes. git cannot authenticate with it anyway, and go-git quotes the whole
// URL in the parse error it returns, so handing it over unchanged would leak
// the credentials into task logs.
func (r Repository) GetGitURL(embedCredentials bool) string {
	rawURL := r.GitURL

	if r.GetType() == RepositoryLocal {
		return util.NormalizeLocalFilesystemPath(rawURL)
	}

	if !hasURLScheme(rawURL) {
		return rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		// Do not fail the task here, but make the reason visible: a URL git
		// cannot be handed credentials for shows up later only as an opaque
		// authentication error. The error itself quotes the URL and is not logged.
		log.WithFields(log.Fields{
			"context":       "repository",
			"repository_id": r.ID,
		}).Warn("can not parse repository url, using it without credentials")
		return redactURLUserinfo(rawURL)
	}

	if !embedCredentials {
		return rawURL
	}

	if r.GetType() == RepositoryHTTP && r.SSHKey.Type == AccessKeyLoginPassword {
		if r.SSHKey.LoginPassword.Login == "" {
			if r.SSHKey.LoginPassword.Password == "" {
				return rawURL
			}
			parsed.User = url.User(r.SSHKey.LoginPassword.Password)
		} else {
			parsed.User = url.UserPassword(r.SSHKey.LoginPassword.Login, r.SSHKey.LoginPassword.Password)
		}

		// Credentials are still embedded for plain http so existing installations
		// keep working, but the transport is unencrypted and the credentials are
		// sent in the clear.
		if strings.EqualFold(parsed.Scheme, "http") {
			log.WithFields(log.Fields{
				"context":       "repository",
				"repository_id": r.ID,
			}).Warn("sending git credentials over an unencrypted http connection, use https instead")
		}

		return parsed.String()
	}

	return rawURL
}

// GetRedactedGitURL returns the repository URL with any userinfo removed, for
// writing to task logs. Unlike GetGitURL(false) it never returns credentials,
// including ones typed into the URL itself.
func (r Repository) GetRedactedGitURL() string {
	if r.GetType() == RepositoryLocal {
		return util.NormalizeLocalFilesystemPath(r.GitURL)
	}
	return redactURLUserinfo(r.GitURL)
}

// redactURLUserinfo drops everything between "://" and the last "@" of a URL.
//
// net/url is deliberately not used: credentials are typed in by hand and are
// often not valid URL syntax. A token containing "/" or "#" makes net/url read
// it as the host or the fragment and report no userinfo at all, so a parser
// based redaction would log it verbatim. Cutting at the last "@" can over-trim
// a URL that has an "@" after the host, which only affects how it is displayed.
// scp-style SSH addresses ("git@host:path") have no scheme, carry no secret and
// are returned unchanged.
func redactURLUserinfo(rawURL string) string {
	schemeEnd := strings.Index(rawURL, "://")
	if schemeEnd < 0 {
		return rawURL
	}
	rest := rawURL[schemeEnd+3:]

	at := strings.LastIndex(rest, "@")
	if at < 0 {
		return rawURL
	}
	return rawURL[:schemeEnd+3] + rest[at+1:]
}

func (r Repository) GetType() RepositoryType {
	if strings.HasPrefix(r.GitURL, "/") {
		return RepositoryLocal
	}

	if util.IsWindowsLocalRepositoryPath(r.GitURL) {
		return RepositoryLocal
	}

	re := regexp.MustCompile(`^(\w+)://`)
	m := re.FindStringSubmatch(r.GitURL)
	if m == nil {
		return RepositorySSH
	}

	// URL schemes are case-insensitive (RFC 3986), and ValidateGitURL accepts
	// any spelling. Without normalizing here, "HTTPS://host/repo" would be
	// reported as type "HTTPS", match neither branch below, and never be given
	// the repository's credentials.
	protocol := strings.ToLower(m[1])

	switch protocol {
	case "http", "https":
		return RepositoryHTTP
	default:
		return RepositoryType(protocol)
	}
}

func (r Repository) Validate() error {
	if r.Name == "" {
		return common_errors.NewValidationError("repository name can't be empty")
	}

	if r.GitURL == "" {
		return common_errors.NewValidationError("repository url can't be empty")
	}

	if err := ValidateGitURL(r.GitURL, "repository"); err != nil {
		return err
	}

	if r.GetType() != RepositoryLocal && r.GitBranch == "" {
		return common_errors.NewValidationError("repository branch can't be empty")
	}

	if err := git.ValidateGitBranch(r.GitBranch, "repository"); err != nil {
		return err
	}

	return nil
}
