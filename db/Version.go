package db

import (
	"fmt"
	"time"
)

// Version represents sql schema version
type Version struct {
	Major int
	Minor int
	Patch int
	Build string

	UpgradedDate *time.Time
	Notes        *string
}

// VersionString returns a well formatted string of the current Version
func (version Version) VersionString() string {
	s := fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)

	if len(version.Build) == 0 {
		return s
	}

	return fmt.Sprintf("%s-%s", s, version.Build)
}

// HumanoidVersion adds a v to the VersionString
func (version Version) HumanoidVersion() string {
	return "v" + version.VersionString()
}

func GetVersions() []Version {
	return []Version{
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
		{Major: 2, Minor: 8, Patch: 0},
		{Major: 2, Minor: 8, Patch: 1},
		{Major: 2, Minor: 8, Patch: 7},
		{Major: 2, Minor: 8, Patch: 8},
		{Major: 2, Minor: 8, Patch: 20},
		{Major: 2, Minor: 8, Patch: 25},
		{Major: 2, Minor: 8, Patch: 26},
	}
}
