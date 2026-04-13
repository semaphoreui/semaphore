//go:build !windows

package util

import (
	"os"
	"os/user"
	"strconv"
	"syscall"
)

func (conf *ConfigType) getProcessCredential() (uid *int, gid *int) {

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

	return
}

func (conf *ConfigType) GetSysProcAttr() (res *syscall.SysProcAttr) {

	if conf.Process.Chroot != "" {
		res = &syscall.SysProcAttr{}
		res.Chroot = conf.Process.Chroot
	}

	uid, gid := conf.getProcessCredential()

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

// ChownDir changes ownership of the directory to the process config user/group.
// This is needed because directories are created by the main Semaphore process,
// but child processes (git, ansible, etc.) run as the configured process user.
func ChownDir(path string) error {
	if Config == nil {
		return nil
	}

	uid, gid := Config.getProcessCredential()

	if uid == nil || gid == nil {
		return nil
	}

	return os.Chown(path, *uid, *gid)
}

