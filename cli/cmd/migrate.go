package cmd

import (
	"fmt"
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
		fmt.Println("\n db migrations run on startup automatically")
	},
}
