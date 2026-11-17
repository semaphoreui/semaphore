package db

import (
	"errors"
	"io"
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

	// Decoding skips a null key with its value, so that value would stay unchecked
	if hasNullKey(&doc) {
		return nil, common_errors.NewValidationError("variable names must not be null")
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

func hasNullKey(node *yaml.Node) bool {
	for i, child := range node.Content {
		if node.Kind == yaml.MappingNode && i%2 == 0 {
			key := child
			if key.Kind == yaml.AliasNode {
				key = key.Alias
			}
			if key.ShortTag() == "!!null" {
				return true
			}
		}
		if hasNullKey(child) {
			return true
		}
	}
	return false
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
