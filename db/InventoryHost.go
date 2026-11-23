package db

import (
	"errors"
	"io"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"gopkg.in/yaml.v3"
)

const inventoryHostNameMaxLength = 250

// InventoryHost is a host of a managed inventory
type InventoryHost struct {
	ID          int       `db:"id" json:"id" backup:"-"`
	InventoryID int       `db:"inventory_id" json:"inventory_id" backup:"-"`
	ProjectID   int       `db:"project_id" json:"project_id" backup:"-"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Enabled     bool      `db:"enabled" json:"enabled"`
	Variables   string    `db:"variables" json:"variables"`
	Created     time.Time `db:"created" json:"created" backup:"-"`
	Groups      []int     `db:"-" json:"groups" backup:"-"`

	// VariableCount is the number of top-level variables, set when the host is saved
	VariableCount int `db:"variable_count" json:"variable_count" backup:"-"`

	// The last run of the host, set by the server and never by a request
	LastStatus *string    `db:"last_status" json:"last_status" backup:"-"`
	LastRun    *time.Time `db:"last_run" json:"last_run" backup:"-"`
	LastTaskID *int       `db:"last_task_id" json:"last_task_id" backup:"-"`
	HostID     *int       `db:"host_id" json:"host_id" backup:"-"`
}

// ParseInventoryVariables parses YAML or JSON text into a mapping
func ParseInventoryVariables(text string) (map[string]any, error) {
	invalid := common_errors.NewValidationError("variables must be a YAML or JSON mapping")
	vars := map[string]any{}

	var doc yaml.Node
	decoder := yaml.NewDecoder(strings.NewReader(text))
	err := decoder.Decode(&doc)
	if errors.Is(err, io.EOF) {
		return vars, nil
	}
	if err != nil {
		return nil, invalid
	}

	// A second document would be dropped silently
	if !errors.Is(decoder.Decode(&yaml.Node{}), io.EOF) {
		return nil, invalid
	}

	// Hosts are written into one text, where anchors of different hosts would meet
	if usesAnchors(&doc) {
		return nil, common_errors.NewValidationError("variables must not use YAML anchors, aliases or merge keys")
	}

	err = doc.Decode(&vars)
	if err != nil {
		return nil, invalid
	}

	if vars == nil {
		vars = map[string]any{}
	}

	return vars, nil
}

// CountInventoryVariables is the number of top-level variables in the text
func CountInventoryVariables(text string) int {
	vars, _ := ParseInventoryVariables(text)
	return len(vars)
}

func usesAnchors(node *yaml.Node) bool {
	if node.Kind == yaml.AliasNode || node.Anchor != "" || node.ShortTag() == "!!merge" {
		return true
	}
	return slices.ContainsFunc(node.Content, usesAnchors)
}

func (h InventoryHost) Validate() error {
	if h.Name == "" {
		return common_errors.NewValidationError("host name must not be empty")
	}

	if strings.ContainsFunc(h.Name, unicode.IsSpace) || strings.ContainsAny(h.Name, ",[]:") {
		return common_errors.NewValidationError("host name must not contain spaces, commas, brackets or colons")
	}

	if utf8.RuneCountInString(h.Name) > inventoryHostNameMaxLength {
		return common_errors.NewValidationError("host name is too long")
	}

	_, err := ParseInventoryVariables(h.Variables)
	return err
}
