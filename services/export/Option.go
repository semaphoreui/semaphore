package export

import (
	"github.com/semaphoreui/semaphore/db"
)

type OptionExporter struct {
	ValueMap[db.Option]
}

func (e *OptionExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	options, err := store.GetOptions(db.RetrieveQueryParams{})
	if err != nil {
		return err
	}

	err = e.appendValues(getOption(options), GlobalScope)
	if err != nil {
		return err
	}
	return nil
}

func getOption(opts map[string]string) []db.Option {
	values := make([]db.Option, 0)

	for key, val := range opts {
		values = append(values, db.Option{
			Key:   key,
			Value: val,
		})
	}

	return values
}

func (e *OptionExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *OptionExporter) restoreValue(val EntityObject[db.Option], store db.Store, exporter DataExporter) (err error) {
	return store.SetOption(val.value.Key, val.value.Value)
}

func (e *OptionExporter) exportDependsOn() []string {
	return []string{}
}

func (e *OptionExporter) importDependsOn() []string {
	return []string{}
}

func (e *OptionExporter) getName() string {
	return Option
}
