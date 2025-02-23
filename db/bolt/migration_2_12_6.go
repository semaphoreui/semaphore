package bolt

type migration_2_12_6 struct {
	migration
}

func (d migration_2_12_6) Apply() (err error) {
	// Add the "path" field to the repository table
	return
}
