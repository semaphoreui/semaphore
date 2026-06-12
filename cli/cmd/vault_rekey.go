package cmd

import (
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/spf13/cobra"
)

func init() {
	vaultRekeyCmd.PersistentFlags().StringVar(&targetVaultArgs.oldKey, "old-key", "", "Old encryption key")

	vaultCmd.AddCommand(vaultRekeyCmd)
}

var vaultRekeyCmd = &cobra.Command{
	Use:   "rekey",
	Short: "Re-encrypt Key Store in database with using current encryption key",
	Long: "To update the encryption key, modify it within the configuration file and " +
		"then employ the 'vault rekey --old-key <old-key>' command to ensure the re-encryption of the " +
		"pre-existing keys stored in the database.",
	Run: func(cmd *cobra.Command, args []string) {
		store := createStore("")
		encryptionService := server.NewAccessKeyEncryptionService(store, store, store, store)
		defer store.Close()

		err := encryptionService.RekeyAccessKeys(targetVaultArgs.oldKey)

		if err != nil {
			panic(err)
		}

	},
}
