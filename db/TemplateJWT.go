package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/semaphoreui/semaphore/util"
)

// maxJWTAudienceEntries caps the number of audience values to keep the token
// size reasonable and prevent abuse via huge config payloads.
const maxJWTAudienceEntries = 32

// TemplateJWTParams holds the JWT configuration for a template.
type TemplateJWTParams struct {
	Enabled  bool     `json:"enabled,omitempty"`
	Audience []string `json:"audience,omitempty"`
	TTL      string   `json:"ttl,omitempty"`
}

// Scan implements sql.Scanner so TemplateJWTParams can be read from the database.
func (p *TemplateJWTParams) Scan(value any) error {
	if value == nil {
		*p = TemplateJWTParams{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*p = TemplateJWTParams{}
			return nil
		}
		return json.Unmarshal(v, p)
	case string:
		if v == "" {
			*p = TemplateJWTParams{}
			return nil
		}
		return json.Unmarshal([]byte(v), p)
	default:
		return errors.New("unsupported type for TemplateJWTParams")
	}
}

// Value implements driver.Valuer so TemplateJWTParams can be written to the database.
func (p *TemplateJWTParams) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

// ParsedTTL returns the parsed TTL or zero when unset.
func (p *TemplateJWTParams) ParsedTTL() (time.Duration, error) {
	if p == nil || p.TTL == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(p.TTL)
	if err != nil {
		return 0, fmt.Errorf("invalid JWT TTL %q: %w", p.TTL, err)
	}
	return d, nil
}

// Validate enforces some sanity checks on the JWT parameters to prevent misconfiguration and abuse.
func (p *TemplateJWTParams) Validate() error {
	if p == nil || !p.Enabled {
		return nil
	}

	if len(p.Audience) > maxJWTAudienceEntries {
		return &ValidationError{fmt.Sprintf("JWT audience must contain at most %d entries", maxJWTAudienceEntries)}
	}
	if slices.Contains(p.Audience, "") {
		return &ValidationError{"JWT audience entries must not be empty"}
	}

	if p.TTL != "" {
		ttl, err := p.ParsedTTL()
		if err != nil {
			return &ValidationError{err.Error()}
		}
		if ttl <= 0 {
			return &ValidationError{"JWT TTL must be positive"}
		}
		max := globalJWTMaxTTL()
		if max > 0 && ttl > max {
			return &ValidationError{fmt.Sprintf("JWT TTL %s exceeds configured maximum %s", ttl, max)}
		}
	}

	return nil
}

// globalJWTMaxTTL returns the configured max TTL for JWTs.
func globalJWTMaxTTL() time.Duration {
	if util.Config == nil {
		return 0
	}
	if util.Config.JWTMaxTTL == "" {
		return 24 * time.Hour
	}
	d, err := time.ParseDuration(util.Config.JWTMaxTTL)
	if err != nil || d <= 0 {
		return 24 * time.Hour
	}
	return d
}
