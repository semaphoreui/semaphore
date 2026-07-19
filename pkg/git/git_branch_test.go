package git

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func TestRepository_Validate_InvalidBranch_ReturnsError(t *testing.T) {
	repo := db.Repository{
		Name:      "Test repository",
		GitURL:    "https://example.com/repo.git",
		GitBranch: "feature branch",
	}

	err := repo.Validate()

	assert.EqualError(t, err, "repository branch name is invalid")
}

func TestRepository_Validate_ValidBranch_ReturnsNoError(t *testing.T) {
	repo := db.Repository{
		Name:      "Test repository",
		GitURL:    "https://example.com/repo.git",
		GitBranch: "feature/test-branch",
	}

	err := repo.Validate()

	assert.NoError(t, err)
}

func TestTemplate_Validate_InvalidBranch_ReturnsError(t *testing.T) {
	branch := "bad..branch"
	tpl := &db.Template{
		Name:      "Deploy",
		Playbook:  "site.yml",
		GitBranch: &branch,
	}

	err := tpl.Validate()

	assert.EqualError(t, err, "template branch name is invalid")
}

func TestTemplate_Validate_EmptyBranchOverride_ReturnsNoError(t *testing.T) {
	branch := ""
	tpl := &db.Template{
		Name:      "Deploy",
		Playbook:  "site.yml",
		GitBranch: &branch,
	}

	err := tpl.Validate()

	assert.NoError(t, err)
}

func TestTask_ValidateNewTask_InvalidBranch_ReturnsError(t *testing.T) {
	branch := "-bad-branch"
	task := &db.Task{
		GitBranch: &branch,
	}

	err := task.ValidateNewTask(db.Template{})

	assert.EqualError(t, err, "task branch name is invalid")
}

func TestTask_ValidateNewTask_ValidBranch_ReturnsNoError(t *testing.T) {
	branch := "release/2026.03"
	task := &db.Task{
		GitBranch: &branch,
	}

	err := task.ValidateNewTask(db.Template{})

	assert.NoError(t, err)
}

func TestValidateCommitHash(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{"empty is allowed", "", false},
		{"short abbreviated hash", "a1b2c3d", false},
		{"full sha1", "0123456789abcdef0123456789abcdef01234567", false},
		{"full sha256", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", false},
		{"uppercase hex", "ABCDEF1", false},
		{"too short", "a1b2c3", true},
		{"branch name", "main", true},
		{"option injection", "--upload-pack=touch /tmp/pwn", true},
		{"leading dashes", "--HEAD", true},
		{"non-hex chars", "z1b2c3d", true},
		{"ref with slash", "refs/heads/main", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommitHash(tt.hash, "task")
			if tt.wantErr {
				assert.EqualError(t, err, "task commit hash is invalid")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTask_ValidateNewTask_InvalidCommitHash_ReturnsError(t *testing.T) {
	hash := "--upload-pack=evil"
	task := &db.Task{
		CommitHash: &hash,
	}

	err := task.ValidateNewTask(db.Template{})

	assert.EqualError(t, err, "task commit hash is invalid")
}
