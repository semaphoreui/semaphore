package cmd

import (
	"fmt"

	"github.com/Digital-Data-Co/forge/util"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Semaphore",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(util.Version())
	},
}
