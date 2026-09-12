package db

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var buildVersionRE = regexp.MustCompile(`^(.*[^\d])?(\d+)([^\d].*)?$`)

// GetNextBuildVersion derives the next version from a startVersion template and
// the most recent version. The numeric segment of startVersion is bumped by one
// (or startVersion's own number is used when it is ahead of currentVersion).
// When startVersion has no numeric segment, or currentVersion does not match its
// prefix/suffix shape, startVersion is returned unchanged.
//
// It backs both build-template task versions and workflow run versions.
func GetNextBuildVersion(startVersion string, currentVersion string) string {
	m := buildVersionRE.FindStringSubmatch(startVersion)

	if m == nil {
		return startVersion
	}

	var prefix, suffix, body string

	switch len(m) - 1 {
	case 3:
		prefix = m[1]
		body = m[2]
		suffix = m[3]
	case 2:
		if _, err := strconv.Atoi(m[1]); err == nil {
			body = m[1]
			suffix = m[2]
		} else {
			prefix = m[1]
			body = m[2]
		}
	case 1:
		body = m[1]
	default:
		return startVersion
	}

	if !strings.HasPrefix(currentVersion, prefix) ||
		!strings.HasSuffix(currentVersion, suffix) {
		return startVersion
	}

	curr, err := strconv.Atoi(currentVersion[len(prefix) : len(currentVersion)-len(suffix)])
	if err != nil {
		return startVersion
	}

	start, err := strconv.Atoi(body)
	if err != nil {
		panic(err)
	}

	var newVer int
	if start > curr {
		newVer = start
	} else {
		newVer = curr + 1
	}

	return prefix + fmt.Sprintf("%0*d", len(body), newVer) + suffix
}
