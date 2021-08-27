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
		store := createStore()
		defer store.Close()
		fmt.Println("\n DB migrations run on startup automatically")
	},
}