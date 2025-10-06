package bolt

import (
	"encoding/json"
	"fmt"

	"github.com/Digital-Data-Co/forge/db"
	"go.etcd.io/bbolt"
)

func (d *BoltDb) migration_2_20_0() error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		// Create the task__file bucket for each project
		projects, err := d.getProjects(tx)
		if err != nil {
			return fmt.Errorf("failed to get projects: %w", err)
		}

		for _, project := range projects {
			bucketName := makeBucketId(db.TaskFileProps, project.ID)
			_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
			if err != nil {
				return fmt.Errorf("failed to create task__file bucket for project %d: %w", project.ID, err)
			}
		}

		// Create a global task__file bucket for global task files (if any)
		globalBucketName := makeBucketId(db.TaskFileProps, 0)
		_, err = tx.CreateBucketIfNotExists([]byte(globalBucketName))
		if err != nil {
			return fmt.Errorf("failed to create global task__file bucket: %w", err)
		}

		return nil
	})
}

func (d *BoltDb) getProjects(tx *bbolt.Tx) ([]db.Project, error) {
	var projects []db.Project
	
	bucketName := makeBucketId(db.ProjectProps, 0)
	b := tx.Bucket([]byte(bucketName))
	if b == nil {
		return projects, nil
	}

	c := b.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var project db.Project
		if err := json.Unmarshal(v, &project); err != nil {
			continue // Skip invalid entries
		}
		projects = append(projects, project)
	}

	return projects, nil
}

