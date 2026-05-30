//go:build windows

package util

import (
	"syscall"
)

func (conf *ConfigType) GetSysProcAttr() (res *syscall.SysProcAttr) {

	return
}

// GetAppSysProcAttr is a no-op on Windows. Linux namespace isolation is
// not available, so child apps run with the same attributes as the main
// process.
func (conf *ConfigType) GetAppSysProcAttr() *syscall.SysProcAttr {
	return conf.GetSysProcAttr()
}

func ChownDir(path string) error {
	return nil
}

