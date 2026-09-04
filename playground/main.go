//go:build linux

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
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

	supCmd := exec.Command(executable, supervisorMode)
	supCmd.Stdout = os.Stdout
	supCmd.Stderr = os.Stderr

	if err := supCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start supervisor: %v\n", err)
		return 1
	}
	fmt.Printf("Supervisor PID: %d\n", supCmd.Process.Pid)

	err = supCmd.Wait()
	if supCmd.ProcessState == nil {
		fmt.Fprintf(os.Stderr, "wait for supervisor: %v\n", err)
		return 1
	}

	status := supCmd.ProcessState.Sys().(syscall.WaitStatus)
	if status.Signaled() {
		fmt.Printf("Supervisor terminated by signal %d (%s)\n", status.Signal(), status.Signal())
		return 128 + int(status.Signal())
	}

	exitCode := status.ExitStatus()
	fmt.Printf("Supervisor exited with status %d\n", exitCode)
	return exitCode
}

func runSupervisor() int {
	supSigTermCh := make(chan os.Signal, 1)
	signal.Notify(supSigTermCh, syscall.SIGTERM)
	defer signal.Stop(supSigTermCh)

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

	taskCmd := exec.Command("bash", "playground/scripts/spawn-background.sh")
	taskCmd.Stdout = os.Stdout
	taskCmd.Stderr = os.Stderr
	// Give the task command a dedicated process group so kill(-pid, ...) targets its task tree.
	taskCmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := taskCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start task command: %v\n", err)
		return 1
	}

	pid := taskCmd.Process.Pid
	fmt.Printf("Task command PID: %d\n", pid)

	supSigTermReceived, waitErr := waitForCmd(pid, supSigTermCh)

	// Kill same-group descendants in one operation before using /proc below to
	// find adopted descendants that escaped into another process group.
	killErr := signalProcessGroup(pid, unix.SIGKILL)

	// Reap only after the group signal; WNOWAIT kept the PID/PGID reserved against reuse.
	// Wait returns an error for a valid non-zero exit, so read the status from ProcessState.
	if err := taskCmd.Wait(); taskCmd.ProcessState == nil {
		fmt.Fprintf(os.Stderr, "reap task command: %v\n", err)
		return 1
	}
	taskCmdStatus := taskCmd.ProcessState.Sys().(syscall.WaitStatus)

	if err := terminateAndReapChildren(); err != nil {
		fmt.Fprintf(os.Stderr, "terminate descendants: %v\n", err)
		return 1
	}
	if waitErr != nil {
		fmt.Fprintf(os.Stderr, "wait for task command exit: %v\n", waitErr)
		return 1
	}
	if killErr != nil {
		fmt.Fprintf(os.Stderr, "kill process group: %v\n", killErr)
		return 1
	}

	if supSigTermReceived || taskCmdStatus.Signal() == syscall.SIGTERM {
		fmt.Println("Task terminated by SIGTERM")
		signal.Stop(supSigTermCh)
		signal.Reset(syscall.SIGTERM)
		if err := unix.Kill(os.Getpid(), unix.SIGTERM); err != nil {
			fmt.Fprintf(os.Stderr, "re-raise SIGTERM: %v\n", err)
			return 1
		}
		// Wait for the SIGTERM sent above to terminate the supervisor instead of
		// returning a normal exit status.
		for {
			_ = unix.Pause()
		}
	}

	return taskCmdStatus.ExitStatus()
}

// The bool reports whether the supervisor received SIGTERM before the command exited.
func waitForCmd(pid int, supSigTermCh <-chan os.Signal) (bool, error) {
	exited := make(chan error, 1)
	go func() {
		for {
			var info unix.Siginfo
			// https://man7.org/linux/man-pages/man2/waitpid.2.html
			// WNOWAIT keeps the command's PID and PGID reserved for process-group cleanup.
			err := unix.Waitid(unix.P_PID, pid, &info, unix.WEXITED|unix.WNOWAIT, nil)
			if errors.Is(err, unix.EINTR) {
				continue
			}
			exited <- err
			return
		}
	}()

	select {
	case err := <-exited:
		return false, err
	case <-supSigTermCh:
		if err := signalProcessGroup(pid, unix.SIGTERM); err != nil {
			return true, err
		}

		select {
		case err := <-exited:
			return true, err
		case <-time.After(time.Second):
			if err := signalProcessGroup(pid, unix.SIGKILL); err != nil {
				return true, err
			}
			// This can block indefinitely if the command is stuck in uninterruptible kernel sleep.
			// https://chrisdown.name/2024/02/05/reliably-creating-d-state-processes-on-demand.html
			return true, <-exited
		}
	}
}

func signalProcessGroup(pid int, sig unix.Signal) error {
	// https://man7.org/linux/man-pages/man2/kill.2.html
	// ESRCH means that the target process or process group no longer exists.
	if err := unix.Kill(-pid, sig); err != nil && !errors.Is(err, unix.ESRCH) {
		return err
	}
	return nil
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
