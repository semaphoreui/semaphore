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

// Scan implements sql.Scanner so TemplateJWTParams can be read from a TEXT
// column. NULL or empty values produce a zero-valued struct.
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

// Value implements driver.Valuer so TemplateJWTParams is JSON-encoded when
// written to a TEXT column. A nil pointer is stored as SQL NULL.
func (p *TemplateJWTParams) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

// ParsedTTL returns the parsed TTL or zero when unset. An error is returned
// when the value is not a valid Go duration.
func (p *TemplateJWTParams) ParsedTTL() (time.Duration, error) {
	if p == nil || p.TTL == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(p.TTL)
	if err != nil {
		return 0, fmt.Errorf("invalid jwt_params.ttl %q: %w", p.TTL, err)
	}
	return d, nil
}

// Validate enforces the per-template JWT invariants:
//   - audience entries are non-empty and within the size cap;
//   - TTL is a positive Go duration not exceeding the configured global
//     ceiling (util.Config.JWTMaxTTL, default 24h).
//
// Validation is skipped entirely when Enabled is false.
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

// globalJWTMaxTTL returns the configured ceiling, falling back to 24h when
// unset or unparseable. Returns 0 only if util.Config is nil (e.g. in tests
// that don't bootstrap config) — callers treat 0 as "no ceiling enforced".
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
