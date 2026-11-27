package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	log "github.com/sirupsen/logrus"
)

var ErrAnsibleInventoryNotFound = errors.New("ansible-inventory is not installed on the server")

const ansibleInventoryTimeout = 5 * time.Minute

// gitInventoryTimeout and ansibleInventoryTimeout together stay below db.InventorySourceStaleAfter
const gitInventoryTimeout = 10 * time.Minute

// maxInventoryOutput limits the ansible-inventory JSON, parsed it takes about 25 times its size in memory, 30 000 hosts with a few variables are about 2 MB
const maxInventoryOutput = 32 << 20

var errInventoryTooLarge = errors.New("inventory is too large")

// urlCredentials takes everything up to the last @ of an address, a token may hold a slash or an @
var urlCredentials = regexp.MustCompile(`://[^\s'"]*@`)

// GitInventoryFetcher reads an inventory file of a project repository with ansible-inventory
type GitInventoryFetcher struct {
	store             db.Store
	keyInstaller      db_lib.AccessKeyInstaller
	encryptionService db_lib.SecretDeserializer

	// cloneLocks has a channel of one place for every source
	cloneLocks sync.Map
}

func NewGitInventoryFetcher(
	store db.Store,
	keyInstaller db_lib.AccessKeyInstaller,
	encryptionService db_lib.SecretDeserializer,
) *GitInventoryFetcher {
	return &GitInventoryFetcher{
		store:             store,
		keyInstaller:      keyInstaller,
		encryptionService: encryptionService,
	}
}

func (f *GitInventoryFetcher) Fetch(ctx context.Context, source db.InventorySource) (db.InventoryDocument, error) {
	if source.RepositoryID == nil {
		return db.InventoryDocument{}, errors.New("source has no repository")
	}

	repo, err := f.store.GetRepository(source.ProjectID, *source.RepositoryID)
	if err != nil {
		return db.InventoryDocument{}, err
	}

	var root string

	if repo.GetType() == db.RepositoryLocal {
		root = repo.GetFullPath(0)
	} else {
		if err = f.encryptionService.DeserializeSecret(&repo.SSHKey); err != nil {
			return db.InventoryDocument{}, err
		}

		hostConfigs, err := db_lib.InstallProjectHostConfigs(f.store, f.encryptionService, repo.ProjectID, task_logger.NopLogger{})
		if err != nil {
			return db.InventoryDocument{}, err
		}
		defer hostConfigs.Destroy()

		// Syncs of one source share a clone directory, the wait ends with the sync
		tmpDirName := fmt.Sprintf("inventory_source_%d", source.ID)
		lock, _ := f.cloneLocks.LoadOrStore(tmpDirName, make(chan struct{}, 1))
		select {
		case lock.(chan struct{}) <- struct{}{}:
			defer func() { <-lock.(chan struct{}) }()
		case <-ctx.Done():
			return db.InventoryDocument{}, errors.New("another sync of the source is still running")
		}

		// A stalled git would hold the lock of the source and block every later sync
		deadline := time.Now().Add(gitInventoryTimeout)
		if end, ok := ctx.Deadline(); ok && end.Before(deadline) {
			deadline = end
		}

		gitLog := &cloneLog{output: cappedBuffer{limit: 64 << 10}}
		git := db_lib.GitRepository{
			Deadline:     deadline,
			Repository:   repo,
			TmpDirName:   tmpDirName,
			Client:       db_lib.CreateDefaultGitClient(f.keyInstaller),
			Logger:       gitLog,
			HostConfigs:  hostConfigs,
			ShallowClone: true,
		}

		// A directory left by a killed server may hold a clone of another repository
		root = git.GetFullPath()
		if err = os.RemoveAll(root); err != nil {
			return db.InventoryDocument{}, err
		}
		defer os.RemoveAll(root) //nolint:errcheck

		if err = git.CloneOrPull(); err != nil {
			if !time.Now().Before(deadline) {
				err = errors.New("repository clone timed out")
			}
			lines := strings.Split(strings.TrimSpace(gitLog.output.String()), "\n")
			text := fmt.Sprintf("%s: %s", err, strings.Join(lines[max(len(lines)-20, 0):], "\n"))
			err = errors.New(sanitizeSourceError(stripServerPaths(text, root+string(filepath.Separator), root)))
			log.WithError(err).WithField("source_id", source.ID).Warn("Failed to clone the repository of an inventory source")
			return db.InventoryDocument{}, err
		}
	}

	// Server paths stay out of the error, users see paths inside the repository
	paths := []string{root + string(filepath.Separator), root}
	if realRoot, err := filepath.EvalSymlinks(root); err == nil {
		paths = append(paths, realRoot+string(filepath.Separator), realRoot)
	}

	if err = rejectOutsideLinks(root); err != nil {
		return db.InventoryDocument{}, errors.New(stripServerPaths(err.Error(), paths...))
	}

	var params db.GitInventorySourceParams
	if err = source.FillParams(&params); err != nil {
		return db.InventoryDocument{}, err
	}

	target, err := resolveInventoryPath(root, params.Path)
	if err != nil {
		return db.InventoryDocument{}, errors.New(stripServerPaths(err.Error(), paths...))
	}

	output, err := runAnsibleInventory(ctx, root, target)
	if err != nil {
		return db.InventoryDocument{}, err
	}

	return ParseAnsibleInventory(output)
}

// cloneLog keeps the end of what git wrote, a failed clone shows it
type cloneLog struct {
	task_logger.NopLogger
	output cappedBuffer
}

func (l *cloneLog) Log(msg string) {
	_, _ = l.output.Write([]byte(msg + "\n"))
}

func (l *cloneLog) Logf(format string, a ...any) {
	l.Log(fmt.Sprintf(format, a...))
}

// LogCmd gives both streams one writer, so exec writes to it from one goroutine
func (l *cloneLog) LogCmd(cmd *exec.Cmd) func() {
	cmd.Stdout = &l.output
	cmd.Stderr = &l.output
	return func() {}
}

// resolveInventoryPath joins root and relative, the result after links are resolved must stay in root
// The path is returned with its own links, Ansible looks for group_vars next to the path it is given
func resolveInventoryPath(root, relative string) (string, error) {
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}

	target, err := filepath.EvalSymlinks(filepath.Join(realRoot, relative))
	if err != nil {
		return "", errors.New("inventory path not found in the repository")
	}

	if !isInside(realRoot, target) {
		return "", errors.New("inventory path is outside of the repository")
	}

	return filepath.Join(realRoot, relative), nil
}

func isInside(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// rejectOutsideLinks fails when a link in the repository leads out of it, ansible-inventory follows such links
func rejectOutsideLinks(root string) error {
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}

	return filepath.WalkDir(realRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.Type()&fs.ModeSymlink == 0 {
			return nil
		}

		target, err := filepath.EvalSymlinks(path)
		if err == nil && isInside(realRoot, target) {
			return nil
		}

		rel, _ := filepath.Rel(realRoot, path)
		return fmt.Errorf("repository has a link that points outside of it: %s", rel)
	})
}

// cappedBuffer keeps at most limit bytes, without cancel the last ones, with cancel a write past the limit fails and stops the command
// The buffer is not embedded, its ReadFrom would let io.Copy in os/exec pass Write by
type cappedBuffer struct {
	buf    bytes.Buffer
	limit  int
	cancel context.CancelCauseFunc
	// cut is set when the beginning was dropped, the first line may then be a piece of an address with credentials
	cut bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if b.buf.Len()+len(p) <= b.limit {
		return b.buf.Write(p)
	}
	if b.cancel != nil {
		b.cancel(errInventoryTooLarge)
		return 0, errInventoryTooLarge
	}

	b.buf.Write(p)
	b.buf.Next(b.buf.Len() - b.limit)
	b.cut = true
	return len(p), nil
}

func (b *cappedBuffer) Bytes() []byte {
	return b.buf.Bytes()
}

func (b *cappedBuffer) String() string {
	text := b.buf.String()
	if _, rest, found := strings.Cut(text, "\n"); b.cut && found {
		return rest
	}
	return text
}

func runAnsibleInventory(ctx context.Context, root, path string) ([]byte, error) {
	return runInventoryCommand(ctx, root, maxInventoryOutput, "ansible-inventory", "-i", path, "--list", "--export")
}

// runInventoryCommand runs the command with its own HOME and stops it when its output passes limit
func runInventoryCommand(ctx context.Context, root string, limit int, name string, args ...string) ([]byte, error) {
	ctx, stop := context.WithCancelCause(ctx)
	defer stop(nil)
	ctx, cancel := context.WithTimeout(ctx, ansibleInventoryTimeout)
	defer cancel()

	home, err := os.MkdirTemp("", "inventory_source_home_")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(home) //nolint:errcheck

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = 10 * time.Second
	cmd.Env = append(db_lib.BaseEnvironmentVars(),
		"ANSIBLE_INVENTORY_ENABLED=yaml,ini",
		"ANSIBLE_INVENTORY_UNPARSED_FAILED=true",
		"ANSIBLE_INVENTORY_ANY_UNPARSED_IS_FAILED=true",
		// Older Ansible also skips .ini in a directory, the list is fixed so every version reads the same files
		"ANSIBLE_INVENTORY_IGNORE_EXTS=.pyc,.pyo,.swp,.bak,~,.rpm,.md,.txt,.rst,.orig,.cfg,.retry",
		"HOME="+home,
	)

	stdout := &cappedBuffer{limit: limit, cancel: stop}
	stderr := &cappedBuffer{limit: 64 << 10}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err = cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, ErrAnsibleInventoryNotFound
		}
		stderrText := stderr.String()
		if ctx.Err() != nil {
			return nil, fmt.Errorf("ansible-inventory: %w", context.Cause(ctx))
		}

		// Server paths stay out of the error, users see paths inside the repository
		paths := []string{root + string(filepath.Separator), home}
		if realRoot, err := filepath.EvalSymlinks(root); err == nil {
			paths = append(paths, realRoot+string(filepath.Separator))
		}
		// The cut below may split an address, so its credentials go first
		stderrText = sanitizeSourceError(stripServerPaths(stderrText, paths...))

		text := strings.TrimSpace(stderrText)
		if runes := []rune(text); len(runes) > 500 {
			text = string(runes[len(runes)-500:])
		}
		return nil, fmt.Errorf("ansible-inventory: %w: %s", err, text)
	}

	return stdout.Bytes(), nil
}

// stripServerPaths removes the paths from the text, the longest first so a path inside another leaves no rest
func stripServerPaths(text string, paths ...string) string {
	paths = slices.Clone(paths)
	slices.SortFunc(paths, func(a, b string) int { return len(b) - len(a) })

	for _, p := range paths {
		if p != "" {
			text = strings.ReplaceAll(text, p, "")
		}
	}

	return text
}

// sanitizeSourceError removes the credentials from the addresses in an error text
func sanitizeSourceError(text string) string {
	return urlCredentials.ReplaceAllString(text, "://")
}
