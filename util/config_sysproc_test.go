//go:build !windows

package util

import (
	"strconv"
	"testing"
)

func TestParseLinuxCredentialUint(t *testing.T) {
	tests := []struct {
		in       string
		wantVal  uint32
		wantOk   bool
	}{
		{"1", 1, true},
		{"65534", 65534, true},
		{strconv.FormatUint(1<<32-1, 10), 1<<32 - 1, true},
		{"0", 0, false},
		{"4294967296", 0, false},
		{"-1", 0, false},
		{"notint", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, ok := parseLinuxCredentialUint(tt.in)
			if ok != tt.wantOk || got != tt.wantVal {
				t.Fatalf("parseLinuxCredentialUint(%q) = (%v, %v), want (%v, %v)", tt.in, got, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}
