package schedules

import "testing"

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