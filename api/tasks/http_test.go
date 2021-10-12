package tasks

import (
	"testing"
)

func TestGetNextBuildVersion(t *testing.T) {
	s := getNextBuildVersion("new-1.4-patch", "new-1.5-patch")
	if s != "new-1.6-patch" {
		t.Fatal()
	}

	s = getNextBuildVersion("new-1.4", "new-1.5")
	if s != "new-1.6" {
		t.Fatal()
	}

	s = getNextBuildVersion("1.4-patch", "1.5-patch")
	if s != "1.6-patch" {
		t.Fatal()
	}

	s = getNextBuildVersion("1.4.8", "1.4.9")
	if s != "1.4.10" {
		t.Fatal()
	}

	s = getNextBuildVersion("0", "7")
	if s != "8" {
		t.Fatal()
	}
}
