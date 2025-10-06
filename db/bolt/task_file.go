package bolt

import (
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/pkg/tz"
	"go.etcd.io/bbolt"
)

func (d *BoltDb) CreateTaskFile(taskFile db.TaskFile) (db.TaskFile, error) {
	err := taskFile.PreInsert(nil)
	if err != nil {
		return db.TaskFile{}, err
	}

	taskFile.ID = 0
	taskFile.Created = tz.Now()
	
	res, err := d.createObject(taskFile.ProjectID, db.TaskFileProps, taskFile)
	if err != nil {
		return db.TaskFile{}, err
	}
	
	return res.(db.TaskFile), nil
}

func (d *BoltDb) GetTaskFiles(projectID int, taskID int) ([]db.TaskFile, error) {
	var result []db.TaskFile
	
	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(db.TaskFileProps, projectID))
		if b == nil {
			return db.ErrNotFound
		}

		c := b.Cursor()
		
		return apply(c, db.TaskFileProps, db.RetrieveQueryParams{}, func(item any) bool {
			taskFile := item.(db.TaskFile)
			return taskFile.TaskID == taskID
		}, func(i any) error {
			taskFile := i.(db.TaskFile)
			result = append(result, taskFile)
			return nil
		})
	})
	
	return result, err
}

func (d *BoltDb) GetTaskFile(projectID int, taskID int, fileID int) (db.TaskFile, error) {
	var result db.TaskFile
	found := false
	
	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(db.TaskFileProps, projectID))
		if b == nil {
			return db.ErrNotFound
		}

		c := b.Cursor()
		
		return apply(c, db.TaskFileProps, db.RetrieveQueryParams{}, func(item any) bool {
			taskFile := item.(db.TaskFile)
			return taskFile.TaskID == taskID && taskFile.ID == fileID
		}, func(i any) error {
			result = i.(db.TaskFile)
			found = true
			return nil
		})
	})
	
	if err != nil {
		return db.TaskFile{}, err
	}
	
	if !found {
		return db.TaskFile{}, db.ErrNotFound
	}
	
	return result, nil
}

func (d *BoltDb) DeleteTaskFile(projectID int, taskID int, fileID int) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(db.TaskFileProps, projectID))
		if b == nil {
			return db.ErrNotFound
		}

		c := b.Cursor()
		
		return apply(c, db.TaskFileProps, db.RetrieveQueryParams{}, func(item any) bool {
			taskFile := item.(db.TaskFile)
			return taskFile.TaskID == taskID && taskFile.ID == fileID
		}, func(i any) error {
			taskFile := i.(db.TaskFile)
			key := makeObjectKey(db.TaskFileProps, taskFile.ID)
			return b.Delete(key)
		})
	})
}

func (d *BoltDb) DeleteTaskFiles(projectID int, taskID int) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(db.TaskFileProps, projectID))
		if b == nil {
			return db.ErrNotFound
		}

		c := b.Cursor()
		
		return apply(c, db.TaskFileProps, db.RetrieveQueryParams{}, func(item any) bool {
			taskFile := item.(db.TaskFile)
			return taskFile.TaskID == taskID
		}, func(i any) error {
			taskFile := i.(db.TaskFile)
			key := makeObjectKey(db.TaskFileProps, taskFile.ID)
			return b.Delete(key)
		})
	})
}
