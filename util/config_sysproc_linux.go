//go:build linux

package util

import "syscall"

// GetAppSysProcAttr returns SysProcAttr for child apps (ansible, terraform,
// shell templates). It creates a dedicated process group and adds configured
// Linux namespace isolation flags. It is NOT used for git: git needs host
// access for SSH agents and credential helpers.
func (conf *ConfigType) GetAppSysProcAttr() *syscall.SysProcAttr {
	res := conf.GetSysProcAttr()
	if res == nil {
		res = &syscall.SysProcAttr{}
	}
	res.Setpgid = true
	res.Cloneflags |= conf.Process.AppNamespaces.cloneFlags()
	return res
}

func (ns ConfigAppNamespaces) cloneFlags() (flags uintptr) {
	if ns.User {
		flags |= syscall.CLONE_NEWUSER
	}
	if ns.Mount {
		flags |= syscall.CLONE_NEWNS
	}
	if ns.PID {
		flags |= syscall.CLONE_NEWPID
	}
	if ns.IPC {
		flags |= syscall.CLONE_NEWIPC
	}
	if ns.UTS {
		flags |= syscall.CLONE_NEWUTS
	}
	return
}
