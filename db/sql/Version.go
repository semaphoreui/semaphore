package sql

import (
	"fmt"
	"strings"
	"time"
)

// Version represents an sql schema version
type Version struct {
	Major int
	Minor int
	Patch int
	Build string

	UpgradedDate *time.Time
	Notes        *string
}

// Versions holds all sql schema version references
var Versions []*Version

// VersionString returns a well formatted string of the current Version
func (version *Version) VersionString() string {
	s := fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)

	if len(version.Build) == 0 {
		return s
	}

	return fmt.Sprintf("%s-%s", s, version.Build)
}

// HumanoidVersion adds a v to the VersionString
func (version *Version) HumanoidVersion() string {
	return "v" + version.VersionString()
}

// GetPath is the humanoid version with the file format appended
func (version *Version) GetPath() string {
	return version.HumanoidVersion() + ".sql"
}

//GetErrPath is the humanoid version with '.err' and file format appended
func (version *Version) GetErrPath() string {
	return version.HumanoidVersion() + ".err.sql"
}

// GetSQL takes a path to an SQL file and returns it from packr as a slice of strings separated by newlines
func (version *Version) GetSQL(path string) []string {
	sql, err := dbAssets.MustString(path)
	if err != nil {
		panic(err)
	}
	return strings.Split(sql, ";\n")
}

func init() {
	Versions = []*Version{
		{},
		{Major: 1},
		{Major: 1, Minor: 2},
		{Major: 1, Minor: 3},
		{Major: 1, Minor: 4},
		{Major: 1, Minor: 5},
		{Major: 1, Minor: 6},
		{Major: 1, Minor: 7},
		{Major: 1, Minor: 8},
		{Major: 1, Minor: 9},
		{Major: 2, Minor: 2, Patch: 1},
		{Major: 2, Minor: 3},
		{Major: 2, Minor: 3, Patch: 1},
		{Major: 2, Minor: 3, Patch: 2},
		{Major: 2, Minor: 4},
		{Major: 2, Minor: 5},
		{Major: 2, Minor: 5, Patch: 2},
		{Major: 2, Minor: 7, Patch: 1},
		{Major: 2, Minor: 7, Patch: 4},
		{Major: 2, Minor: 7, Patch: 6},
		{Major: 2, Minor: 7, Patch: 8},
		{Major: 2, Minor: 7, Patch: 9},
		{Major: 2, Minor: 7, Patch: 10},
		{Major: 2, Minor: 7, Patch: 12},
		{Major: 2, Minor: 7, Patch: 13},
	}
}
