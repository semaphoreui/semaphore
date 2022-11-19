package cmd

import (
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Execute migrations",
	Run: func(cmd *cobra.Command, args []string) {
		store := createStore("migrate")
		defer store.Close("migrate")
		util.Config.PrintDbInfo()
	},
}
