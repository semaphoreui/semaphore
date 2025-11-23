package schedules

import (
	"testing"
	"time"
)

func TestValidateCronFormat(t *testing.T) {
	err := ValidateCronFormat("* * * *")
	if err == nil {
		t.Fatal("")
	}

	err = ValidateCronFormat("* * 1 * *")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestOneTimeSchedule(t *testing.T) {
	future := time.Now().Add(time.Hour)
	schedule := oneTimeSchedule{runAt: future}

	if schedule.Next(time.Now()) != future {
		t.Fatalf("expected next run at %v", future)
	}

	if !schedule.Next(future).IsZero() {
		t.Fatalf("expected schedule to stop after run time")
	}
}
