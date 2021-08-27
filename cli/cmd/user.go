package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

type userArgs struct {
	login    string
	name     string
	email    string
	password string
}

func init() {
	rootCmd.AddCommand(userCmd)
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}
