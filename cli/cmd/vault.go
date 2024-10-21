package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

type vaultArgs struct {
	oldKey string
}

var targetVaultArgs vaultArgs

func init() {
	rootCmd.AddCommand(vaultCmd)
}

var vaultCmd = &cobra.Command{
	Use:     "vaults",
	Aliases: []string{"vault"},
	Short:   "Manage access keys and other secrets",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}
