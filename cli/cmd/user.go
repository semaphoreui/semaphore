package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

type UserAddArgs struct {
	Username string
	Name     string
	Email    string
	Password string
}

var UserAdd *UserAddArgs

func init() {
	rootCmd.AddCommand(userCmd)
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	// Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}