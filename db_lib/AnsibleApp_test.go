package db_lib

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// collectingLogger is a minimal task_logger.Logger implementation that records
// the messages passed to Log, so tests can assert on what was logged.
type collectingLogger struct {
	messages []string
}

func (l *collectingLogger) Log(msg string) { l.messages = append(l.messages, msg) }

// The remaining methods satisfy the task_logger.Logger interface with no-ops.
func (l *collectingLogger) Logf(format string, a ...any)                        {}
func (l *collectingLogger) LogWithTime(now time.Time, msg string)               {}
func (l *collectingLogger) LogfWithTime(now time.Time, format string, a ...any) {}
func (l *collectingLogger) LogCmd(cmd *exec.Cmd)                                {}
func (l *collectingLogger) SetStatus(status task_logger.TaskStatus)             {}
func (l *collectingLogger) AddStatusListener(s task_logger.StatusListener)      {}
func (l *collectingLogger) AddLogListener(s task_logger.LogListener)            {}
func (l *collectingLogger) SetCommit(hash, message string)                      {}
func (l *collectingLogger) WaitLog()                                            {}

func (l *collectingLogger) joined() string {
	return strings.Join(l.messages, "")
}

func (l *collectingLogger) countContaining(substr string) int {
	n := 0
	for _, m := range l.messages {
		if strings.Contains(m, substr) {
			n++
		}
	}
	return n
}

// newTestApp builds an AnsibleApp whose repository points at dir (a local
// repository, since the path is absolute) with the playbook at the repo root,
// so GetPlaybookDir() == getRepoPath() == dir.
func newTestApp(dir string, logger task_logger.Logger) *AnsibleApp {
	return &AnsibleApp{
		Logger: logger,
		Template: db.Template{
			ID:       1,
			Playbook: "", // playbook at repo root
		},
		Repository: db.Repository{
			ID:        1,
			ProjectID: 1,
			GitURL:    dir, // absolute path => RepositoryLocal
		},
	}
}

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(p, 0o755))
}

func mustWrite(t *testing.T, p string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(path.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, []byte("---\n"), 0o644))
}

func TestResolveGalaxyRequirements_NothingExists(t *testing.T) {
	dir := t.TempDir()
	logger := &collectingLogger{}
	app := newTestApp(dir, logger)

	rolePaths, collectionPaths := app.resolveGalaxyRequirements()

	assert.Empty(t, rolePaths)
	assert.Empty(t, collectionPaths)

	// No roles/ or collections/ dirs => no warnings about them.
	assert.Equal(t, 0, logger.countContaining("contains no requirements.yml"))

	// No shared requirements.yml anywhere => exactly one "No requirements.yml found" message.
	assert.Equal(t, 1, logger.countContaining("No requirements.yml found"))
	// It should list the searched path.
	assert.Contains(t, logger.joined(), path.Join(dir, "requirements.yml"))
}

func TestResolveGalaxyRequirements_SubdirExistsButNoFile(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, path.Join(dir, "roles"))
	mustMkdir(t, path.Join(dir, "collections"))

	logger := &collectingLogger{}
	app := newTestApp(dir, logger)

	rolePaths, collectionPaths := app.resolveGalaxyRequirements()

	assert.Empty(t, rolePaths)
	assert.Empty(t, collectionPaths)

	// Each existing-but-empty subdir produces exactly one warning.
	assert.Equal(t, 1, logger.countContaining(path.Join(dir, "roles")+" exists but contains no requirements.yml"))
	assert.Equal(t, 1, logger.countContaining(path.Join(dir, "collections")+" exists but contains no requirements.yml"))

	// Still no shared root requirements.yml => one "not found" message.
	assert.Equal(t, 1, logger.countContaining("No requirements.yml found"))
}

func TestResolveGalaxyRequirements_SubdirFilesExist(t *testing.T) {
	dir := t.TempDir()
	rolesReq := path.Join(dir, "roles", "requirements.yml")
	collectionsReq := path.Join(dir, "collections", "requirements.yml")
	mustWrite(t, rolesReq)
	mustWrite(t, collectionsReq)

	logger := &collectingLogger{}
	app := newTestApp(dir, logger)

	rolePaths, collectionPaths := app.resolveGalaxyRequirements()

	assert.Equal(t, []string{rolesReq}, rolePaths)
	assert.Equal(t, []string{collectionsReq}, collectionPaths)

	// Subdirs have their files => no warnings.
	assert.Equal(t, 0, logger.countContaining("contains no requirements.yml"))

	// No shared root requirements.yml => one "not found" message (subdir files are separate).
	assert.Equal(t, 1, logger.countContaining("No requirements.yml found"))
}

func TestResolveGalaxyRequirements_SharedRootFile(t *testing.T) {
	dir := t.TempDir()
	sharedReq := path.Join(dir, "requirements.yml")
	mustWrite(t, sharedReq)

	logger := &collectingLogger{}
	app := newTestApp(dir, logger)

	rolePaths, collectionPaths := app.resolveGalaxyRequirements()

	// Shared file is offered to BOTH types.
	assert.Equal(t, []string{sharedReq}, rolePaths)
	assert.Equal(t, []string{sharedReq}, collectionPaths)

	// Shared file exists => NO "not found" message at all.
	assert.Equal(t, 0, logger.countContaining("No requirements.yml found"))
	// And no subdir warnings (no roles/ or collections/ dirs).
	assert.Equal(t, 0, logger.countContaining("contains no requirements.yml"))
}

func TestResolveGalaxyRequirements_AllPresent(t *testing.T) {
	dir := t.TempDir()
	rolesReq := path.Join(dir, "roles", "requirements.yml")
	collectionsReq := path.Join(dir, "collections", "requirements.yml")
	sharedReq := path.Join(dir, "requirements.yml")
	mustWrite(t, rolesReq)
	mustWrite(t, collectionsReq)
	mustWrite(t, sharedReq)

	logger := &collectingLogger{}
	app := newTestApp(dir, logger)

	rolePaths, collectionPaths := app.resolveGalaxyRequirements()

	// roles: subdir file + shared file.
	assert.Contains(t, rolePaths, rolesReq)
	assert.Contains(t, rolePaths, sharedReq)
	assert.Len(t, rolePaths, 2)

	// collections: subdir file + shared file.
	assert.Contains(t, collectionPaths, collectionsReq)
	assert.Contains(t, collectionPaths, sharedReq)
	assert.Len(t, collectionPaths, 2)

	// Nothing missing => no warnings, no "not found".
	assert.Equal(t, 0, logger.countContaining("contains no requirements.yml"))
	assert.Equal(t, 0, logger.countContaining("No requirements.yml found"))
}
