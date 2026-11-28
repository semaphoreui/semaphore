package db

import (
	"encoding/json"
	"errors"
	"path"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

const inventorySourceNameMaxLength = 250

const inventorySourcePathMaxLength = 1000

const inventorySourceParamsMaxSize = 16 << 10

type InventorySourceType string

const InventorySourceGit InventorySourceType = "git"

const (
	InventorySourceSyncing = "syncing"
	InventorySourceSuccess = "success"
	InventorySourceFailed  = "failed"
)

var ErrInventorySourceTypeUnsupported = errors.New("source type is not supported")

// ErrInventoryRowOwnerChanged refuses a write when a source took or left the row after the caller read it
var ErrInventoryRowOwnerChanged error = conflictError{"owner changed, reload and retry"}

// ErrInventorySourceSyncing refuses an edit of a source while its sync runs
var ErrInventorySourceSyncing error = conflictError{"source is syncing, try again when the sync ends"}

// InventorySourceInterrupted is the error of a sync that did not end in InventorySourceStaleAfter
const InventorySourceInterrupted = "sync was interrupted"

// InventorySourceStaleAfter is how long a started sync is trusted before another may start
const InventorySourceStaleAfter = 30 * time.Minute

// InventorySource is a place a managed inventory imports its hosts and groups from
type InventorySource struct {
	ID          int                 `db:"id" json:"id" backup:"-"`
	InventoryID int                 `db:"inventory_id" json:"inventory_id" backup:"-"`
	ProjectID   int                 `db:"project_id" json:"project_id" backup:"-"`
	Name        string              `db:"name" json:"name"`
	Type        InventorySourceType `db:"type" json:"type"`

	RepositoryID *int `db:"repository_id" json:"repository_id"`
	// KeyID is the access key a type signs in with, git uses the key of its repository
	KeyID *int `db:"key_id" json:"key_id"`
	// Params are the settings of the type. They are shown in the API and kept in a backup, so secrets go to KeyID
	Params MapStringAnyField `db:"params" json:"params"`

	// Variables are written by the sync only
	Variables     string `db:"variables" json:"variables"`
	VariableCount int    `db:"variable_count" json:"variable_count"`

	Status      string               `db:"status" json:"status"`
	SyncStarted *time.Time           `db:"sync_started" json:"sync_started"`
	SyncRun     int                  `db:"sync_run" json:"-" backup:"-"`
	LastSync    *time.Time           `db:"last_sync" json:"last_sync"`
	LastSuccess *time.Time           `db:"last_success" json:"last_success"`
	LastError   string               `db:"last_error" json:"last_error"`
	LastResult  *InventorySyncResult `db:"-" json:"last_result"`
	HostCount   int                  `db:"-" json:"host_count"`
	GroupCount  int                  `db:"-" json:"group_count"`
	Created     time.Time            `db:"created" json:"created" backup:"-"`
}

// InventorySyncResult is what the last sync changed, the column keeps 20 skips
type InventorySyncResult struct {
	HostsAdded    int                 `json:"hosts_added"`
	HostsChanged  int                 `json:"hosts_changed"`
	HostsRemoved  int                 `json:"hosts_removed"`
	GroupsAdded   int                 `json:"groups_added"`
	GroupsChanged int                 `json:"groups_changed"`
	GroupsRemoved int                 `json:"groups_removed"`
	Adopted       int                 `json:"adopted"`
	Skipped       []InventorySyncSkip `json:"skipped"`
	SkippedCount  int                 `json:"skipped_count"`
	// VariablesChanged is true when the variables of group all changed
	VariablesChanged bool `json:"variables_changed"`
}

// InventorySyncSkip is a host or group the sync left out, Kind is "host" or "group"
type InventorySyncSkip struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

func (s InventorySource) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return common_errors.NewValidationError("source name must not be empty")
	}

	if utf8.RuneCountInString(s.Name) > inventorySourceNameMaxLength {
		return common_errors.NewValidationError("source name is too long")
	}

	if s.Type == "" {
		return common_errors.NewValidationError("source type must not be empty")
	}

	params, err := json.Marshal(s.Params)
	if err != nil || len(params) > inventorySourceParamsMaxSize {
		return common_errors.NewValidationError("source params are too large")
	}

	if s.Type != InventorySourceGit {
		return nil
	}

	if s.KeyID != nil {
		return common_errors.NewValidationError("a git source uses the key of its repository")
	}

	var git GitInventorySourceParams
	if err = s.FillParams(&git); err != nil {
		return common_errors.NewValidationError("source params are invalid")
	}
	return git.Validate()
}

// FillParams reads the settings of the type into target
func (s InventorySource) FillParams(target any) error {
	content, err := json.Marshal(s.Params)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, target)
}

// GitInventorySourceParams are the settings of a git source
type GitInventorySourceParams struct {
	// Path is a file or a directory of the repository
	Path string `json:"path"`
}

func (p GitInventorySourceParams) Validate() error {
	if utf8.RuneCountInString(p.Path) > inventorySourcePathMaxLength || strings.ContainsRune(p.Path, 0) {
		return common_errors.NewValidationError("source path is too long or has a wrong character")
	}

	// C:\\dir and C:file name a drive on Windows
	drive := len(p.Path) > 1 && p.Path[1] == ':' && unicode.IsLetter(rune(p.Path[0]))
	if drive || strings.HasPrefix(p.Path, "/") || strings.HasPrefix(p.Path, "\\") ||
		slices.Contains(strings.FieldsFunc(p.Path, func(r rune) bool { return r == '/' || r == '\\' }), "..") {
		return common_errors.NewValidationError("source path must be relative and must not contain ..")
	}

	// The root of a repository holds playbooks and other files, which fail the sync
	if p.Path == "" || path.Clean(strings.ReplaceAll(p.Path, "\\", "/")) == "." {
		return common_errors.NewValidationError("source path is required")
	}

	return nil
}

func (r InventorySyncResult) Changed() bool {
	return r.HostsAdded+r.HostsChanged+r.HostsRemoved+r.GroupsAdded+r.GroupsChanged+r.GroupsRemoved != 0 || r.VariablesChanged
}
