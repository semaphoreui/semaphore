//go:build windows

package util

import (
	"syscall"
)

func (conf *ConfigType) GetSysProcAttr() (res *syscall.SysProcAttr) {

	return
}

func ChownDir(path string) error {
	return nil
}

