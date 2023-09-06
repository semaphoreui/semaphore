package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	vaultCmd.PersistentFlags().StringVar(&targetVaultArgs.oldKey, "old-key", "", "Old encryption key")

	vaultCmd.AddCommand(vaultRekeyCmd)
}

var vaultRekeyCmd = &cobra.Command{
	Use:   "rekey",
	Short: "Re-encrypt Key Store in database to current encryption key",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		store := createStore("")
		defer store.Close("")

		err := store.RekeyAccessKeys(targetVaultArgs.oldKey)

		if err != nil {
			panic(err)
		}

	},
}
