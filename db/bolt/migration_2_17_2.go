package bolt

type migration_2_17_2 struct {
	migration
}

func (d migration_2_17_2) Apply() error {
	// No-op migration for BoltDB.
	// The run_once field is added to the Schedule struct and will be handled automatically.
	return nil
}
