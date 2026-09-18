//go:build !windows

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeLocalFilesystemPath(t *testing.T) {
	for _, p := range []string{`/D:/ps-demo-script`, `/d/ps-demo-script/extra`, `/usr/src`} {
		assert.Equal(t, p, NormalizeLocalFilesystemPath(p), "non-Windows must leave path unchanged")
	}
}
