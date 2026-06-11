package db

import (
	"encoding/base64"
	"slices"
	"time"

	"github.com/gorilla/securecookie"
)

type RunnerState string

type RunnerTagFilterMode string

const (
	RunnerFilterTagCompleteMatch RunnerTagFilterMode = "complete_match"
	RunnerFilterHasAnyTag        RunnerTagFilterMode = "has_any_tag"
	RunnerFilterIgnoreTags       RunnerTagFilterMode = "ignore_tags"
	RunnerFilterIsDefault        RunnerTagFilterMode = "is_default"
)

type Runner struct {
	ID                int        `db:"id" json:"id"`
	Token             string     `db:"token" json:"-"`
	ProjectID         *int       `db:"project_id" json:"project_id"`
	Webhook           string     `db:"webhook" json:"webhook"`
	MaxParallelTasks  int        `db:"max_parallel_tasks" json:"max_parallel_tasks"`
	Active            bool       `db:"active" json:"active"`
	IsDefault         bool       `db:"is_default" json:"is_default"`
	Name              string     `db:"name" json:"name"`
	Tags              []string   `db:"-" json:"tags" backup:"tags"`
	Touched           *time.Time `db:"touched" json:"touched"`
	CleaningRequested *time.Time `db:"cleaning_requested" json:"cleaning_requested"`

	// StartedAt is the runner process's start time, reported by the runner on
	// every poll (X-Runner-Started-At header) and persisted next to Touched.
	// It changes on every restart, which is how the server detects that a
	// runner lost its in-memory job pool while still polling.
	StartedAt *time.Time `db:"started_at" json:"started_at"`

	PublicKey *string `db:"public_key" json:"-"`

	// Registered is a transient flag (never persisted) used at creation time to
	// request a runner without an auth token. Such a runner gets a one-time,
	// short-lived registration token instead and must be registered later by
	// presenting that token to `semaphore runner register`.
	Registered bool `db:"-" json:"registered"`

	// RegistrationTokenHash is the stored SHA-256 hash of the one-time registration
	// token (the plaintext is never persisted). It is issued on demand via
	// RegenerateRegistrationToken, not at creation time.
	RegistrationTokenHash      *string    `db:"registration_token" json:"-"`
	RegistrationTokenExpiresAt *time.Time `db:"registration_token_expires_at" json:"-"`
}

// IsRegistered reports whether the runner has been registered (has a token).
func (r Runner) IsRegistered() bool {
	return r.Token != ""
}

// GenerateRunnerToken creates a new runner authentication token.
func GenerateRunnerToken() string {
	return base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
}

// HasTag reports whether the runner is tagged with the given tag.
func (r Runner) HasTag(tag string) bool {
	return slices.Contains(r.Tags, tag)
}

// IsOnline reports whether the runner is considered reachable for dispatch.
// A poll-based runner is online while its last poll (Touched) is within
// offlineTimeout. Webhook-driven runners do not poll, so heartbeat staleness
// does not apply to them — they are always dispatch candidates.
func (r Runner) IsOnline(now time.Time, offlineTimeout time.Duration) bool {
	if r.Webhook != "" {
		return true
	}
	return r.Touched != nil && now.Sub(*r.Touched) <= offlineTimeout
}

type RunnerTag struct {
	Tag             string `db:"-" json:"tag"`
	NumberOfRunners int    `db:"-" json:"number_of_runners"`
}
