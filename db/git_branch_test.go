package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepository_Validate_InvalidBranch_ReturnsError(t *testing.T) {
	repo := Repository{
		Name:      "Test repository",
		GitURL:    "https://example.com/repo.git",
		GitBranch: "feature branch",
	}

	err := repo.Validate()

	assert.EqualError(t, err, "repository branch name is invalid")
}

func TestRepository_Validate_ValidBranch_ReturnsNoError(t *testing.T) {
	repo := Repository{
		Name:      "Test repository",
		GitURL:    "https://example.com/repo.git",
		GitBranch: "feature/test-branch",
	}

	err := repo.Validate()

	assert.NoError(t, err)
}

func TestTemplate_Validate_InvalidBranch_ReturnsError(t *testing.T) {
	branch := "bad..branch"
	tpl := &Template{
		Name:      "Deploy",
		Playbook:  "site.yml",
		GitBranch: &branch,
	}

	err := tpl.Validate()

	assert.EqualError(t, err, "template branch name is invalid")
}

func TestTemplate_Validate_EmptyBranchOverride_ReturnsNoError(t *testing.T) {
	branch := ""
	tpl := &Template{
		Name:      "Deploy",
		Playbook:  "site.yml",
		GitBranch: &branch,
	}

	err := tpl.Validate()

	assert.NoError(t, err)
}

func TestTask_ValidateNewTask_InvalidBranch_ReturnsError(t *testing.T) {
	branch := "-bad-branch"
	task := &Task{
		GitBranch: &branch,
	}

	err := task.ValidateNewTask(Template{})

	assert.EqualError(t, err, "task branch name is invalid")
}

func TestTask_ValidateNewTask_ValidBranch_ReturnsNoError(t *testing.T) {
	branch := "release/2026.03"
	task := &Task{
		GitBranch: &branch,
	}

	err := task.ValidateNewTask(Template{})

	assert.NoError(t, err)
}
