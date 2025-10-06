package bolt

import (
	"encoding/json"
	"log"

	"github.com/Digital-Data-Co/forge/db"
)

func init() {
	db.RegisterMigration("2.19.0", migration_2_19_0)
}

func migration_2_19_0(store *Store) error {
	log.Println("Running migration 2.19.0: Add folder field to templates")

	// Get the templates bucket
	templatesBucket := store.db.Bucket([]byte("templates"))
	if templatesBucket == nil {
		log.Println("Templates bucket not found, skipping migration")
		return nil
	}

	// Iterate through all templates and add folder field
	return templatesBucket.ForEach(func(k, v []byte) error {
		var template db.Template
		if err := json.Unmarshal(v, &template); err != nil {
			log.Printf("Error unmarshaling template %s: %v", string(k), err)
			return nil // Continue with next template
		}

		// Add folder field if it doesn't exist
		if template.Folder == nil {
			template.Folder = nil // Initialize as nil
		}

		// Marshal and save back
		updatedData, err := json.Marshal(template)
		if err != nil {
			log.Printf("Error marshaling template %s: %v", string(k), err)
			return nil
		}

		return templatesBucket.Put(k, updatedData)
	})
}
