package db

import "github.com/go-git/go-git/v5/plumbing"

func ValidateGitBranch(branch string, objectName string) error {
	if branch == "" {
		return nil
	}

	if err := plumbing.NewBranchReferenceName(branch).Validate(); err != nil {
		return NewValidationError(objectName + " branch name is invalid")
	}

	return nil
}
