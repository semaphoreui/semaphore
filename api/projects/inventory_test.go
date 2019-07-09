package projects

import (
	"runtime"
	"testing"
)

func TestIsValidInventoryPath(t *testing.T) {
	if !IsValidInventoryPath("inventories/test") {
		t.Fatal(" a path below the cwd should be valid")
	}

	if !IsValidInventoryPath("inventories/test/../prod") {
		t.Fatal(" a path below the cwd should be valid")
	}

	if IsValidInventoryPath("/test/../../../inventory") {
		t.Fatal(" a path out of the cwd should be invalid")
	}

	if IsValidInventoryPath("/test/inventory") {
		t.Fatal(" a path out of the cwd should be invalid")
	}

	if runtime.GOOS == "windows" && IsValidInventoryPath("c:\\test\\inventory") {
		t.Fatal(" a path out of the cwd should be invalid")
	}
}
