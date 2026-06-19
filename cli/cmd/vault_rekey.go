package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

func init() {
	vaultRekeyCmd.PersistentFlags().StringVar(&targetVaultArgs.oldKey, "old-key", "",
		"Explicit old encryption key (legacy single-key migration; not needed when the old key is configured as a secondary)")
	vaultRekeyCmd.PersistentFlags().StringVar(&targetVaultArgs.backupFile, "backup", "",
		"Write a backup of current access key ciphertexts to this file before re-encrypting")
	vaultRekeyCmd.PersistentFlags().StringVar(&targetVaultArgs.rollbackFile, "rollback", "",
		"Restore access key ciphertexts from a backup file instead of re-encrypting")

	vaultCmd.AddCommand(vaultRekeyCmd)
}

// accessKeyBackupEntry is one line of a rekey backup file.
type accessKeyBackupEntry struct {
	ProjectID int    `json:"project_id"`
	KeyID     int    `json:"key_id"`
	Secret    string `json:"secret"`
}

var vaultRekeyCmd = &cobra.Command{
	Use:   "rekey",
	Short: "Re-encrypt the Key Store with the current encryption key",
	Long: "Re-encrypt all locally stored secrets with the current access keyring primary key.\n\n" +
		"Zero-downtime rotation:\n" +
		"  1. Set the new key as encryption_keys.access_key.primary and move the old\n" +
		"     key to encryption_keys.access_key.secondary, then restart.\n" +
		"  2. Run `vault rekey` to re-encrypt everything (access keys and the JWT\n" +
		"     signing key) under the new primary.\n" +
		"  3. Run `vault check`; once everything reports the primary, remove the\n" +
		"     secondary.\n\n" +
		"Legacy: `vault rekey --old-key <old-key>` decrypts with an explicit old key.",
	Run: func(cmd *cobra.Command, args []string) {
		store := createStore("")
		defer store.Close()

		encryptionService := server.NewAccessKeyEncryptionService(store, store, store, store)

		if targetVaultArgs.rollbackFile != "" {
			if err := rollbackAccessKeys(store, encryptionService, targetVaultArgs.rollbackFile); err != nil {
				panic(err)
			}
			fmt.Println("Rollback complete.")
			return
		}

		if targetVaultArgs.backupFile != "" {
			if err := backupAccessKeys(store, targetVaultArgs.backupFile); err != nil {
				panic(err)
			}
			fmt.Printf("Backup written to %s\n", targetVaultArgs.backupFile)
		}

		if err := encryptionService.RekeyAccessKeys(targetVaultArgs.oldKey); err != nil {
			panic(err)
		}

		if err := util.RekeyJWTSigningKey(store, targetVaultArgs.oldKey); err != nil {
			panic(err)
		}

		fmt.Println("Rekey complete.")
	},
}

// eachLocalAccessKey iterates every locally-encrypted access key across all
// projects, invoking fn for each. Keys backed by external secret storages are
// skipped (their secrets are not encrypted with the access keyring).
func eachLocalAccessKey(store db.Store, fn func(key db.AccessKey) error) error {
	projects, err := store.GetAllProjects()
	if err != nil {
		return err
	}

	for _, project := range projects {
		for offset := 0; ; offset++ {
			keys, err := store.GetAccessKeys(
				project.ID,
				db.GetAccessKeyOptions{IgnoreOwner: true},
				db.RetrieveQueryParams{Count: server.RekeyBatchSize, Offset: offset * server.RekeyBatchSize},
			)
			if err != nil {
				return err
			}
			if len(keys) == 0 {
				break
			}
			for _, key := range keys {
				if key.SourceStorageType != nil {
					continue
				}
				if err := fn(key); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func backupAccessKeys(store db.Store, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	enc := json.NewEncoder(w)

	err = eachLocalAccessKey(store, func(key db.AccessKey) error {
		if key.Secret == nil {
			return nil
		}
		return enc.Encode(accessKeyBackupEntry{
			ProjectID: derefInt(key.ProjectID),
			KeyID:     key.ID,
			Secret:    *key.Secret,
		})
	})
	if err != nil {
		return err
	}

	return w.Flush()
}

func rollbackAccessKeys(store db.Store, encryptionService server.AccessKeyEncryptionService, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Index current keys by id for lookup.
	current := map[int]db.AccessKey{}
	if err := eachLocalAccessKey(store, func(key db.AccessKey) error {
		current[key.ID] = key
		return nil
	}); err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 8*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry accessKeyBackupEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			return err
		}

		key, ok := current[entry.KeyID]
		if !ok {
			continue
		}

		// Populate the plaintext fields from the current ciphertext so that
		// UpdateAccessKey's validation passes, then write back the backed-up
		// ciphertext verbatim.
		if err := encryptionService.DeserializeSecret(&key); err != nil {
			return err
		}

		secret := entry.Secret
		key.Secret = &secret
		key.OverrideSecret = true

		if err := store.UpdateAccessKey(key); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
