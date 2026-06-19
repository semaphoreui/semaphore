package cmd

import (
	"fmt"
	"os"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

func init() {
	vaultCmd.AddCommand(vaultCheckCmd)
}

var vaultCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Verify that all encrypted secrets decrypt under the current keyring",
	Long: "Read-only check that every locally stored Access Key secret and the JWT\n" +
		"signing key decrypt under the current keyring, reporting which key slot\n" +
		"decrypts each. Use it after `vault rekey` to confirm a retired secondary\n" +
		"key can be safely removed: when everything reports the primary, nothing\n" +
		"depends on the secondary any more.",
	Run: func(cmd *cobra.Command, args []string) {
		store := createStore("")
		defer store.Close()

		deserializer := server.NewLocalAccessKeyDeserializer()

		var total, onPrimary, onSecondary, failed int

		err := eachLocalAccessKey(store, func(key db.AccessKey) error {
			total++

			if _, err := deserializer.DeserializeSecret2(&key, util.Config.AccessSecretPrimaryKey()); err == nil {
				onPrimary++
				return nil
			}

			if _, err := deserializer.DeserializeSecret(&key); err == nil {
				onSecondary++
				fmt.Printf("access key %d (project %d): decrypts with a SECONDARY key — rekey pending\n",
					key.ID, derefInt(key.ProjectID))
				return nil
			}

			failed++
			fmt.Printf("access key %d (project %d): FAILED to decrypt with any key\n",
				key.ID, derefInt(key.ProjectID))
			return nil
		})
		if err != nil {
			panic(err)
		}

		fmt.Printf("\nAccess keys: %d total, %d on primary, %d on secondary (rekey pending), %d failed\n",
			total, onPrimary, onSecondary, failed)

		slot, err := util.CheckJWTSigningKey(store)
		if err != nil {
			panic(err)
		}
		if slot == "" {
			fmt.Println("JWT signing key: not set")
		} else {
			fmt.Printf("JWT signing key: %s\n", slot)
			if slot == "FAILED" {
				failed++
			}
		}

		if failed > 0 {
			os.Exit(1)
		}
	},
}
