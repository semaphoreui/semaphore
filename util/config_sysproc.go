//go:build !windows

package util

import (
	"math"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

func (conf *ConfigType) getProcessCredential() (uid uint32, gid uint32) {

	if conf.Process.User != "" {
		usr, err := user.Lookup(conf.Process.User)
		if err != nil {
			return
		}
		u, err := strconv.Atoi(usr.Uid)
		if err != nil {
			return
		}

		if u > 0 && u <= math.MaxUint32 {
			uid = uint32(u)
		}

		g, err := strconv.Atoi(usr.Gid)
		if err != nil {
			return
		}

		if g > 0 && g <= math.MaxUint32 {
			gid = uint32(g)
		}
	}

	if conf.Process.UID != nil {
		uid = *conf.Process.UID
	}

	if conf.Process.GID != nil {
		gid = *conf.Process.GID
	}

	return
}

func (conf *ConfigType) GetSysProcAttr() (res *syscall.SysProcAttr) {

	if conf.Process.Chroot != "" {
		res = &syscall.SysProcAttr{}
		res.Chroot = conf.Process.Chroot
	}

	uid, gid := conf.getProcessCredential()

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


// ChownDir changes ownership of the directory to the process config user/group.
// This is needed because directories are created by the main Semaphore process,
// but child processes (git, ansible, etc.) run as the configured process user.
func ChownDir(path string) error {
	uid, gid := Config.getProcessCredential()

	if uid <= 0 || gid <= 0 || uid > math.MaxInt32 || gid > math.MaxInt32 {
		return nil
	}

	return os.Chown(path, int(uid), int(gid))
}
