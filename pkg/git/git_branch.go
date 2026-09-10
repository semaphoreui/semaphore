package git

import (
	"regexp"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

func ValidateGitBranch(branch string, objectName string) error {
	if branch == "" {
		return nil
	}

	if err := plumbing.NewBranchReferenceName(branch).Validate(); err != nil {
		return common_errors.NewValidationError(objectName + " branch name is invalid")
	}

	return nil
}

// commitHashPattern matches a full or abbreviated git object hash (SHA-1 or
// SHA-256), 7 to 64 hex characters. Restricting the value to hex prevents it
// from carrying git command-line options (e.g. leading "--") or arbitrary
// refs when it is passed to `git checkout`.
var commitHashPattern = regexp.MustCompile(`^[0-9a-fA-F]{7,64}$`)

// ValidateCommitHash rejects a commit hash that is not a plain hex object name.
func ValidateCommitHash(hash string, objectName string) error {
	if hash == "" {
		return nil
	}

	if !commitHashPattern.MatchString(hash) {
		return common_errors.NewValidationError(objectName + " commit hash is invalid")
	}

	return nil
}
