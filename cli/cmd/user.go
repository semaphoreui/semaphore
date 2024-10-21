package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

type userArgs struct {
	login    string
	name     string
	email    string
	password string
	admin    bool
}

var targetUserArgs userArgs

func init() {
	rootCmd.AddCommand(userCmd)
}

var userCmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user"},
	Short:   "Manage users",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}
