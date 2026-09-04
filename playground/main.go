//go:build linux

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const supervisorMode = "--supervisor"

func main() {
	if len(os.Args) == 2 && os.Args[1] == supervisorMode {
		os.Exit(runSupervisor())
	}
	os.Exit(runServer())
}

func runServer() int {
	executable, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "find executable: %v\n", err)
		return 1
	}

	cmd := exec.Command(executable, supervisorMode)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if cmd.ProcessState == nil {
		fmt.Fprintf(os.Stderr, "run supervisor: %v\n", err)
		return 1
	}

	exitCode := cmd.ProcessState.ExitCode()
	fmt.Printf("Supervisor exited with status %d\n", exitCode)
	return exitCode
}

func runSupervisor() int {
	// https://man7.org/linux/man-pages/man2/PR_SET_CHILD_SUBREAPER.2const.html
	//
	// Establishing a subreaper process is useful in session management
	// frameworks where a hierarchical group of processes is managed by a
	// subreaper process that needs to be informed when one of the
	// processes—for example, a double-forked daemon—terminates (perhaps
	// so that it can restart that process).  Some init(1) frameworks
	// (e.g., systemd(1)) employ a subreaper process for similar reasons.
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		fmt.Fprintf(os.Stderr, "enable subreaper: %v\n", err)
		return 1
	}

	cmd := exec.Command("bash", "playground/scripts/spawn-background.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Give Bash a dedicated process group so kill(-pid, ...) targets its task tree.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start Bash: %v\n", err)
		return 1
	}

	pid := cmd.Process.Pid
	var info unix.Siginfo
	// https://man7.org/linux/man-pages/man2/waitpid.2.html
	// WEXITED
	//         Wait for children that have terminated.
	// WNOWAIT
	//         Leave the child in a waitable state; a later wait call can
	//         be used to again retrieve the child status information.
	if err := unix.Waitid(unix.P_PID, pid, &info, unix.WEXITED|unix.WNOWAIT, nil); err != nil {
		fmt.Fprintf(os.Stderr, "wait for Bash exit: %v\n", err)
		return 1
	}

	// https://man7.org/linux/man-pages/man2/kill.2.html
	// ESRCH
	//         The target process or process group does not exist. Note
	//         that an existing process might be a zombie, a process that
	//         has terminated execution, but has not yet been waited for.
	if err := unix.Kill(-pid, unix.SIGKILL); err != nil && !errors.Is(err, unix.ESRCH) {
		fmt.Fprintf(os.Stderr, "kill process group: %v\n", err)
		return 1
	}

	// Wait returns an error for a valid non-zero exit, so read the status from ProcessState.
	if err := cmd.Wait(); cmd.ProcessState == nil {
		fmt.Fprintf(os.Stderr, "reap Bash: %v\n", err)
		return 1
	}
	exitCode := cmd.ProcessState.ExitCode()

	if err := terminateAndReapChildren(); err != nil {
		fmt.Fprintf(os.Stderr, "terminate descendants: %v\n", err)
		return 1
	}

	fmt.Printf("Bash exited with status %d\n", exitCode)
	return exitCode
}

func terminateAndReapChildren() error {
	for {
		pids, err := childPIDs()
		if err != nil {
			return err
		}
		for _, pid := range pids {
			err := unix.Kill(pid, unix.SIGKILL)
			if err != nil && !errors.Is(err, unix.ESRCH) {
				return err
			}
		}

		var info unix.Siginfo
		// https://man7.org/linux/man-pages/man2/waitpid.2.html
		// P_ALL waits for any child. WNOHANG avoids blocking if the /proc child-list
		// read races with a child exiting, creating another child, or being adopted;
		// the loop can rescan and signal the newly visible child instead.
		err = unix.Waitid(unix.P_ALL, 0, &info, unix.WEXITED|unix.WNOHANG, nil)

		// ECHILD (for waitpid() or waitid()) The process specified by pid
		//        (waitpid()) or idtype and id (waitid()) does not exist or
		//        is not a child of the calling process.  (This can happen
		//        for one's own child if the action for SIGCHLD is set to
		//        SIG_IGN.  See also the Linux Notes section about threads.)
		if errors.Is(err, unix.ECHILD) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Signo == 0 {
			// Children still exist but none have exited; rescan without blocking.
			// Bazel uses the same delay while waiting for SIGKILL to take effect:
			// src/main/tools/process-tools-linux.cc#L47-L51
			time.Sleep(100 * time.Microsecond)
		}
	}
}

func childPIDs() ([]int, error) {
	// https://man7.org/linux/man-pages/man5/proc_tid_children.5.html
	// A Go process can have multiple OS threads, so inspect every thread's child list.
	paths, err := filepath.Glob("/proc/self/task/*/children")
	if err != nil {
		return nil, err
	}

	pids := make(map[int]struct{})
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for field := range strings.FieldsSeq(string(data)) {
			pid, err := strconv.Atoi(field)
			if err != nil {
				return nil, err
			}
			pids[pid] = struct{}{}
		}
	}

	result := make([]int, 0, len(pids))
	for pid := range pids {
		result = append(result, pid)
	}
	return result, nil
}
