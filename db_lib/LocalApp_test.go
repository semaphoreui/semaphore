package db_lib

import (
	"os"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/util"
)

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.HasPrefix(s, item) {
			return true
		}
	}
	return false
}

func TestGetEnvironmentVars(t *testing.T) {

	os.Setenv("SEMAPHORE_TEST", "test123")
	os.Setenv("SEMAPHORE_TEST2", "test222")
	os.Setenv("PASSWORD", "test222")

	util.Config = &util.ConfigType{
		ForwardedEnvVars: []string{"SEMAPHORE_TEST"},
		EnvVars: map[string]string{
			"ANSIBLE_FORCE_COLOR": "False",
		},
	}

	res := getEnvironmentVars()

	expected := []string{
		"SEMAPHORE_TEST=test123",
		"ANSIBLE_FORCE_COLOR=False",
		"PATH=",
	}

	if len(res) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, res)
	}

	for _, e := range expected {

		if !contains(res, e) {
			t.Errorf("Expected %v, got %v", expected, res)
		}
	}
}
