package migration

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/export"
)

type Migrator struct {
	OldStore db.Store
	NewStore db.Store

	ErrLogSize     int
	SkipTaskOutput bool
}

func (m *Migrator) Migrate() error {
	if err := m.migrateProject(); err != nil {
		return err
	}

	return nil
}

func (m *Migrator) migrateProject() error {

	mapper := export.NewKeyMapper()
	p := export.InitProjectExporters(mapper, m.SkipTaskOutput)

	err := p.Load(m.OldStore)
	if err != nil {
		return err
	}

	err = p.Restore(m.NewStore, m.ErrLogSize)
	if err != nil {
		return err
	}

	return err
}
