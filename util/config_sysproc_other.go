//go:build !linux && !windows

package util

import "syscall"

// GetAppSysProcAttr creates a dedicated process group on Unix systems that do
// not support Linux CLONE_NEW* namespaces.
func (conf *ConfigType) GetAppSysProcAttr() *syscall.SysProcAttr {
	res := conf.GetSysProcAttr()
	if res == nil {
		res = &syscall.SysProcAttr{}
	}
	res.Setpgid = true
	return res
}
