package bolt

import (
	"go.etcd.io/bbolt"
)

type migration_2_17_0 struct {
	migration
}

func (m migration_2_17_0) Apply() error {
	return m.db.Update(func(tx *bbolt.Tx) error {
		// Create buckets for compliance entities
		buckets := []string{
			"scap_content",
			"scap_profile",
			"compliance_policy",
			"policy_assignment",
			"compliance_scan",
			"compliance_report",
			"compliance_rule_result",
		}

		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return err
			}
		}

		return nil
	})
}
