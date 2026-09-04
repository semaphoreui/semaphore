//go:build !windows

package util

import (
	"io/fs"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
)

// parseLinuxCredentialUint parses a UID/GID string from the password database.
// It rejects zero and values that do not fit in uint32 so they are never
// silently truncated when passed to syscall.Credential.
func parseLinuxCredentialUint(s string) (v uint32, ok bool) {
	u64, err := strconv.ParseUint(s, 10, 32)
	if err != nil || u64 == 0 {
		return 0, false
	}
	return uint32(u64), true
}

func (conf *ConfigType) getProcessCredential() (uid uint32, gid uint32) {

	if conf.Process.User != "" {
		usr, err := user.Lookup(conf.Process.User)
		if err != nil {
			return
		}
		if u, ok := parseLinuxCredentialUint(usr.Uid); ok {
			uid = u
		}

		if g, ok := parseLinuxCredentialUint(usr.Gid); ok {
			gid = g
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

// ChownTree changes ownership of the whole tree to the process config
// user/group. Needed when the main Semaphore process creates files inside a
// directory that child processes (git, ansible, etc.) have to write to.
func ChownTree(root string) error {
	uid, gid := Config.getProcessCredential()

	if uid <= 0 || gid <= 0 || uid > math.MaxInt32 || gid > math.MaxInt32 {
		return nil
	}

	u, g := int(uid), int(gid)

	return filepath.WalkDir(root, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return os.Lchown(path, u, g)
	})
}
