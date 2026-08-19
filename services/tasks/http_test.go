package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
)

func TestGetNextBuildVersion(t *testing.T) {
	s := db.GetNextBuildVersion("new-1.4-patch", "new-1.5-patch")
	if s != "new-1.6-patch" {
		t.Fatal()
	}

	s = db.GetNextBuildVersion("new-1.4", "new-1.5")
	if s != "new-1.6" {
		t.Fatal()
	}

	s = db.GetNextBuildVersion("1.4-patch", "1.5-patch")
	if s != "1.6-patch" {
		t.Fatal()
	}

	s = db.GetNextBuildVersion("1.4.8", "1.4.9")
	if s != "1.4.10" {
		t.Fatal()
	}

	s = db.GetNextBuildVersion("0", "7")
	if s != "8" {
		t.Fatal()
	}

	s = db.GetNextBuildVersion("1.2.0001", "1.2.0005")
	if s != "1.2.0006" {
		t.Fatal("expected 1.2.0006, got " + s)
	}

	s = db.GetNextBuildVersion("0001", "0099")
	if s != "0100" {
		t.Fatal("expected 0100, got " + s)
	}

	s = db.GetNextBuildVersion("build-0010-rc", "build-0042-rc")
	if s != "build-0043-rc" {
		t.Fatal("expected build-0043-rc, got " + s)
	}
}
