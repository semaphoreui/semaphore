package db

import (
	"slices"
	"time"
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

	PublicKey *string `db:"public_key" json:"-"`

	// Unregistered is a transient flag (never persisted) used at creation time to
	// request a runner without a token. Such a runner gets an empty token and must
	// be registered later via `semaphore runner register --runner-id <id>`.
	Unregistered bool `db:"-" json:"unregistered"`
}

// IsRegistered reports whether the runner has been registered (has a token).
func (r Runner) IsRegistered() bool {
	return r.Token != ""
}

// HasTag reports whether the runner is tagged with the given tag.
func (r Runner) HasTag(tag string) bool {
	return slices.Contains(r.Tags, tag)
}

type RunnerTag struct {
	Tag             string `db:"-" json:"tag"`
	NumberOfRunners int    `db:"-" json:"number_of_runners"`
}
