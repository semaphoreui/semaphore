package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("bash", "playground/scripts/spawn-background.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if cmd.ProcessState == nil {
		fmt.Fprintf(os.Stderr, "run Bash: %v\n", err)
		os.Exit(1)
	}

	exitCode := cmd.ProcessState.ExitCode()
	fmt.Printf("Bash exited with status %d\n", exitCode)
	os.Exit(exitCode)
}
