package bolt

import (
	"encoding/json"
	"github.com/ansible-semaphore/semaphore/db"
	"go.etcd.io/bbolt"
	"time"
)

//func (d *BoltDb) getEventObjectName(evt db.Event) (string, error) {
//	if evt.ObjectID == nil || evt.ObjectType == nil {
//		return "", nil
//	}
//	switch *evt.ObjectType {
//	case "task":
//		task, err := d.GetTask(*evt.ProjectID, *evt.ObjectID)
//		if err != nil {
//			return "", err
//		}
//		return task.Playbook, nil
//	default:
//		return "", nil
//	}
//}

// getEvents filter and sort enumerable object passed via parameter.
func (d *BoltDb) getEvents(c enumerable, params db.RetrieveQueryParams, filter func(db.Event) bool) (events []db.Event, err error) {

	i := 0 // offset counter
	n := 0 // number of added items

	events = []db.Event{}

	for k, v := c.First(); k != nil; k, v = c.Next() {
		if params.Offset > 0 && i < params.Offset {
			i++
			continue
		}

		var evt db.Event
		err = json.Unmarshal(v, &evt)

		if err != nil {
			break
		}

		if !filter(evt) {
			continue
		}

		if evt.ProjectID != nil {
			var proj db.Project
			proj, err = d.GetProject(*evt.ProjectID)
			if err != nil {
				break
			}
			evt.ProjectName = &proj.Name
		}

		events = append(events, evt)

		n++

		if n > params.Count {
			break
		}
	}

	err = db.FillEvents(d, events)

	return
}

func (d *BoltDb) CreateEvent(evt db.Event) (newEvent db.Event, err error) {
	newEvent = evt
	newEvent.Created = time.Now()

	err = d.db.Update(func(tx *bbolt.Tx) error {
		b, err2 := tx.CreateBucketIfNotExists([]byte("events"))
		if err2 != nil {
			return err2
		}

		str, err2 := json.Marshal(newEvent)
		if err2 != nil {
			return err2
		}

		id, err2 := b.NextSequence()
		if err2 != nil {
			return err2
		}

		id = MaxID - id

		return b.Put(intObjectID(id).ToBytes(), str)
	})

	return
}

func (d *BoltDb) GetUserEvents(userID int, params db.RetrieveQueryParams) (events []db.Event, err error) {
	err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("events"))
		if b == nil {
			return nil
		}

		c := b.Cursor()
		events, err = d.getEvents(c, params, func(evt db.Event) bool {
			if evt.ProjectID == nil {
				return false
			}
			_, err2 := d.GetProjectUser(*evt.ProjectID, userID)
			return err2 == nil
		})

		return nil
	})

	return
}

func (d *BoltDb) GetEvents(projectID int, params db.RetrieveQueryParams) (events []db.Event, err error) {
	err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("events"))
		if b == nil {
			return nil
		}

		c := b.Cursor()
		events, err = d.getEvents(c, params, func(evt db.Event) bool {
			if evt.ProjectID == nil {
				return false
			}
			return *evt.ProjectID == projectID
		})

		return nil
	})

	return
}
