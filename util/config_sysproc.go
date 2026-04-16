//go:build !windows

package util

import (
	"math"
	"os/user"
	"strconv"
	"syscall"
)

func (conf *ConfigType) GetSysProcAttr() (res *syscall.SysProcAttr) {

	if conf.Process.Chroot != "" {
		res = &syscall.SysProcAttr{}
		res.Chroot = conf.Process.Chroot
	}

	var uid uint32
	var gid uint32

	if conf.Process.UID != nil {
		uid = *conf.Process.UID
	}

	if conf.Process.GID != nil {
		gid = *conf.Process.GID
	}

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

		if u > 0 && u <= math.MaxUint32 {
			uid = uint32(u)
		}

		if g > 0 && g <= math.MaxUint32 {
			gid = uint32(g)
		}

	}

	if uid > 0 && gid > 0 {
		if res == nil {
			res = &syscall.SysProcAttr{}
		}

		res.Credential = &syscall.Credential{
			Uid: uid,
			Gid: gid,
		}
	}

	return
}
