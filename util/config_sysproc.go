//go:build !windows

package util

import (
	"os/user"
	"strconv"
	"syscall"
)

func (conf *ConfigType) GetSysProcAttr() (res *syscall.SysProcAttr) {

	if conf.Process.Chroot != "" {
		res = &syscall.SysProcAttr{}
		res.Chroot = conf.Process.Chroot
	}

	var uid *int
	var gid *int

	uid = nil
	gid = conf.Process.GID

	if conf.Process.User != "" {
		usr, err := user.Lookup(conf.Process.User)
		if err != nil {
			return
		}

		u, err := strconv.Atoi(usr.Uid)
		if err != nil {
			return
		}

		g, err := strconv.Atoi(usr.Gid)
		if err != nil {
			return
		}

		uid = &u
		gid = &g
	}

	if uid != nil && gid != nil {
		if res == nil {
			res = &syscall.SysProcAttr{}
		}

		res.Credential = &syscall.Credential{
			Uid: uint32(*uid),
			Gid: uint32(*gid),
		}
	}

	return
}
