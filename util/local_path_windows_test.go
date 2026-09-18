//go:build windows

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeLocalFilesystemPath(t *testing.T) {
	assert.Equal(t, `D:/ps-demo-script`, NormalizeLocalFilesystemPath(`/D:/ps-demo-script`))
	assert.Equal(t, `D:\ps-demo-script\extra`, NormalizeLocalFilesystemPath(`/d/ps-demo-script/extra`))
}
