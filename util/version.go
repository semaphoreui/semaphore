package util

import (
	"os"
	"strings"
)

var (
	Ver    = "undefined"
	Commit = "00000000"
	Date   = ""
)

func Version() string {
	// Check for SEMAPHORE_BUILD_INFO environment variable override
	if buildInfo := os.Getenv("SEMAPHORE_BUILD_INFO"); buildInfo != "" {
		return buildInfo
	}
	
	// Fall back to default version construction
	return strings.Join([]string{
		Ver,
		Commit,
		Date,
	}, "-")
}
