package server

import (
	"context"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
)

func ValidateInventorySource(source db.InventorySource) error {
	return db.ErrInventorySourceTypeUnsupported
}

func FetchInventorySource(ctx context.Context, store db.Store, secrets db_lib.SecretDeserializer, source db.InventorySource) (db.InventoryDocument, error) {
	return db.InventoryDocument{}, db.ErrInventorySourceTypeUnsupported
}
