//go:build linux

package main

import (
	"errors"
	"flag"
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

func main() {
	supervisorMode := flag.Bool("supervisor", false, "run as the task supervisor")
	gracePeriod := flag.Duration("termination-grace-period", 15*time.Second, "delay before escalating to SIGKILL")
	cleanupTimeout := flag.Duration("cleanup-timeout", 15*time.Second, "maximum time to reap descendants")
	flag.Parse()

	if *gracePeriod < 0 {
		fmt.Fprintln(os.Stderr, "termination grace period cannot be negative")
		os.Exit(1)
	}
	if *cleanupTimeout <= 0 {
		fmt.Fprintln(os.Stderr, "cleanup timeout must be positive")
		os.Exit(1)
	}
	if *supervisorMode {
		command := flag.Args()
		if len(command) == 0 {
			fmt.Fprintln(os.Stderr, "task command is required")
			os.Exit(1)
		}
		os.Exit(runSupervisor(command, *gracePeriod, *cleanupTimeout))
	}
	os.Exit(runServer(*gracePeriod, *cleanupTimeout))
}

func runServer(gracePeriod, cleanupTimeout time.Duration) int {
	executable, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "find executable: %v\n", err)
		return 1
	}

	command := []string{"playground/scripts/wrapper.sh"}
	supArgs := []string{
		"--supervisor",
		"--termination-grace-period=" + gracePeriod.String(),
		"--cleanup-timeout=" + cleanupTimeout.String(),
		"--",
	}
	supArgs = append(supArgs, command...)
	supCmd := exec.Command(executable, supArgs...)
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

func runSupervisor(command []string, gracePeriod, cleanupTimeout time.Duration) int {
	supSignalCh := make(chan os.Signal, 1)
	signal.Notify(supSignalCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(supSignalCh)

	// https://man7.org/linux/man-pages/man2/PR_SET_CHILD_SUBREAPER.2const.html
	//
	// Establishing a subreaper process is useful in session management
	// frameworks where a hierarchical group of processes is managed by a
	// subreaper process that needs to be informed when one of the
	// processes—for example, a double-forked daemon—terminates (perhaps
	// so that it can restart that process).  Some init(1) frameworks
	// (e.g., systemd(1)) employ a subreaper process for similar reasons.
	subreaperEnabled := true
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		if !errors.Is(err, unix.EINVAL) {
			fmt.Fprintf(os.Stderr, "enable subreaper: %v\n", err)
			return 1
		}
		subreaperEnabled = false
		fmt.Fprintln(os.Stderr, "subreaper unavailable; cleanup is limited to the task process group")
	}

	taskCmd := exec.Command(command[0], command[1:]...)
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

	supSignal, waitErr := waitForCmd(pid, supSignalCh, gracePeriod)

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

	if subreaperEnabled {
		if err := terminateAndReapChildren(cleanupTimeout); err != nil {
			fmt.Fprintf(os.Stderr, "terminate descendants: %v\n", err)
			return 1
		}
	}
	if waitErr != nil {
		fmt.Fprintf(os.Stderr, "wait for task command exit: %v\n", waitErr)
		return 1
	}
	if killErr != nil {
		fmt.Fprintf(os.Stderr, "kill process group: %v\n", killErr)
		return 1
	}

	exitSignal := taskCmdStatus.Signal()
	if supSignal != 0 {
		exitSignal = supSignal
	}
	if exitSignal > 0 {
		fmt.Printf("Task terminated by signal %d (%s)\n", exitSignal, exitSignal)
		signal.Stop(supSignalCh)
		signal.Reset(exitSignal)
		if err := unix.Kill(os.Getpid(), exitSignal); err != nil {
			fmt.Fprintf(os.Stderr, "re-raise signal %d: %v\n", exitSignal, err)
			return 1
		}
		// Wait for the signal sent above to terminate the supervisor instead of
		// returning a normal exit status.
		for {
			_ = unix.Pause()
		}
	}

	return taskCmdStatus.ExitStatus()
}

// The signal result is zero unless the supervisor received a termination signal.
func waitForCmd(pid int, supSignalCh <-chan os.Signal, gracePeriod time.Duration) (syscall.Signal, error) {
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
		return 0, err
	case osSignal := <-supSignalCh:
		sig := osSignal.(syscall.Signal)
		if err := signalProcessGroup(pid, sig); err != nil {
			return sig, err
		}

		select {
		case err := <-exited:
			return sig, err
		case <-time.After(gracePeriod):
			if err := signalProcessGroup(pid, unix.SIGKILL); err != nil {
				return sig, err
			}
			// This can block indefinitely if the command is stuck in uninterruptible kernel sleep.
			// https://chrisdown.name/2024/02/05/reliably-creating-d-state-processes-on-demand.html
			return sig, <-exited
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

func terminateAndReapChildren(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		// Avoid waiting forever when /proc omits a live child or SIGKILL cannot
		// terminate one, for example while it is in uninterruptible kernel sleep.
		if time.Now().After(deadline) {
			return fmt.Errorf("descendant cleanup timed out after %s", timeout)
		}
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
