package bolt

type migration_2_17_2 struct {
	migration
}

func (d migration_2_17_2) Apply() error {
	// No-op migration for BoltDB.
	// The project_id field is added to the Role struct and will be handled automatically.
	return nil
}
