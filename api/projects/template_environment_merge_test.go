package projects

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
)

func TestMergeTemplateEnvironmentIDs_omittedPreservesPrevious(t *testing.T) {
	prev := db.Template{EnvironmentIDs: []int{10, 20}}
	updated := db.Template{}
	mergeTemplateEnvironmentIDs(&updated, prev)
	if len(updated.EnvironmentIDs) != 2 || updated.EnvironmentIDs[0] != 10 || updated.EnvironmentIDs[1] != 20 {
		t.Fatalf("expected preserved ids, got %#v", updated.EnvironmentIDs)
	}
}

func TestMergeTemplateEnvironmentIDs_explicitEmptyClears(t *testing.T) {
	prev := db.Template{EnvironmentIDs: []int{10}}
	updated := db.Template{EnvironmentIDs: []int{}}
	mergeTemplateEnvironmentIDs(&updated, prev)
	if updated.EnvironmentIDs == nil || len(updated.EnvironmentIDs) != 0 {
		t.Fatalf("expected empty slice, got %#v", updated.EnvironmentIDs)
	}
}

func TestMergeTemplateEnvironmentIDs_deprecatedEnvironmentID(t *testing.T) {
	prev := db.Template{EnvironmentIDs: []int{99}}
	updated := db.Template{EnvironmentID: 5}
	mergeTemplateEnvironmentIDs(&updated, prev)
	if len(updated.EnvironmentIDs) != 1 || updated.EnvironmentIDs[0] != 5 {
		t.Fatalf("expected [5], got %#v", updated.EnvironmentIDs)
	}
}

func TestMergeTemplateEnvironmentIDs_explicitReplace(t *testing.T) {
	prev := db.Template{EnvironmentIDs: []int{1, 2, 3}}
	updated := db.Template{EnvironmentIDs: []int{7}}
	mergeTemplateEnvironmentIDs(&updated, prev)
	if len(updated.EnvironmentIDs) != 1 || updated.EnvironmentIDs[0] != 7 {
		t.Fatalf("expected [7], got %#v", updated.EnvironmentIDs)
	}
}
