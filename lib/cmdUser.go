//go:build !windows

package lib

import (
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

func cmdUser(cmd *exec.Cmd, username string) {
	u, err := user.Lookup(username)
	if err != nil {
		panic(err)
	}
	uid, err := strconv.ParseInt(u.Uid, 10, 32)
	gid, err := strconv.ParseInt(u.Gid, 10, 32)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
}
