//go:build !linux && !windows

package util

import "syscall"

// GetAppSysProcAttr falls back to the base SysProcAttr on OSes that don't
// support Linux CLONE_NEW* namespaces. The flags are configurable but
// silently inert here.
func (conf *ConfigType) GetAppSysProcAttr() *syscall.SysProcAttr {
	return conf.GetSysProcAttr()
}
