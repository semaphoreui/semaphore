package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

func init() {
	vaultCmd.AddCommand(vaultCheckCmd)
}

var vaultCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Report which key id encrypts each stored secret",
	Long: "Read-only. Reports, per key id, how many locally stored Access Key secrets\n" +
		"and the JWT signing key it encrypts, plus the JWT option's status. Use it\n" +
		"after `vault rekey` to confirm a retired key is safe to remove: a key with\n" +
		"zero references can be deleted from the keyset. Rows whose key id is missing\n" +
		"from the keyset are flagged and cause a non-zero exit.",
	Run: func(cmd *cobra.Command, args []string) {
		store := createStore("")
		defer store.Close()

		activeAccess := util.Config.ActiveAccessKeyID()
		activeOption := util.Config.ActiveOptionKeyID()
		isActive := func(id string) bool {
			return id != "" && (id == activeAccess || id == activeOption)
		}

		counts := map[string]int{} // key id ("" = legacy/no prefix) -> row count
		var total, missing int

		err := eachLocalAccessKey(store, func(key db.AccessKey) error {
			if key.Secret == nil || *key.Secret == "" {
				return nil
			}
			total++
			id := util.SecretKeyID(*key.Secret) // "" for legacy un-prefixed values
			counts[id]++
			if id != "" && !util.Config.HasKeyID(id) {
				missing++
			}
			return nil
		})
		if err != nil {
			panic(err)
		}

		fmt.Printf("Access keys: %d total\n", total)

		// Report every key id seen in the database, plus keyset keys with no rows.
		ids := map[string]struct{}{}
		for id := range counts {
			ids[id] = struct{}{}
		}
		for _, id := range util.Config.KeyIDs() {
			ids[id] = struct{}{}
		}
		sorted := make([]string, 0, len(ids))
		for id := range ids {
			sorted = append(sorted, id)
		}
		sort.Strings(sorted)

		for _, id := range sorted {
			n := counts[id]
			switch {
			case id == "":
				fmt.Printf("  legacy (no id): %d rows — rekey to stamp an id\n", n)
			case !util.Config.HasKeyID(id):
				fmt.Printf("  %s: %d rows — MISSING KEY (cannot decrypt)\n", id, n)
			case isActive(id):
				fmt.Printf("  %s: %d rows — active\n", id, n)
			case n == 0:
				fmt.Printf("  %s: 0 rows — retired, SAFE TO REMOVE\n", id)
			default:
				fmt.Printf("  %s: %d rows — retired, rekey pending\n", id, n)
			}
		}

		slot, err := util.CheckJWTSigningKey(store)
		if err != nil {
			panic(err)
		}
		if slot == "" {
			fmt.Println("JWT signing key: not set")
		} else {
			fmt.Printf("JWT signing key: %s\n", slot)
		}

		if missing > 0 {
			os.Exit(1)
		}
	},
}
