package cmd

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/services/runners"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runnerCmd)
}

func runRunner() {

	taskPool := runners.JobPool{}

	util.Config.PrintDbInfo()

	fmt.Printf("Tmp Path (projects home) %v\n", util.Config.TmpPath)
	fmt.Printf("Semaphore %v\n", util.Version)
	fmt.Printf("Interface %v\n", util.Config.Interface)
	fmt.Printf("Port %v\n", util.Config.Port)

	go taskPool.Run()
}

var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "Run in runner mode",
	Run: func(cmd *cobra.Command, args []string) {
		runRunner()
	},
}
