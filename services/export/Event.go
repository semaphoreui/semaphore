package export

import (
	"fmt"
	"math"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type EventExporter struct {
	ValueMap[db.Event]
}

func (e *EventExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	envs, err := store.GetAllEvents(db.RetrieveQueryParams{Count: math.MaxInt})
	if err != nil {
		return err
	}

	eventsByProject := make(map[string][]db.Event)

	for _, event := range envs {

		scope := GlobalScope
		if event.ProjectID != nil {
			scope = strconv.Itoa(*event.ProjectID)
		}

		if eventsByProject[scope] == nil {
			eventsByProject[scope] = make([]db.Event, 0)
		}
		eventsByProject[scope] = append(eventsByProject[scope], event)
	}

	for scope, events := range eventsByProject {
		err = e.appendValues(events, scope)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *EventExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *EventExporter) restoreValue(val EntityObject[db.Event], store db.Store, exporter DataExporter) (err error) {
	old := val.value

	scope := GlobalScope
	if old.ProjectID != nil {
		scope = strconv.Itoa(*old.ProjectID)
	}

	old.ProjectID, err = exporter.getNewKeyIntRef(Project, GlobalScope, old.ProjectID, e)
	if err != nil {
		return err
	}

	old.UserID, err = exporter.getNewKeyIntRef(User, GlobalScope, old.UserID, e)
	if err != nil {
		return err
	}

	old.IntegrationID, err = exporter.getNewKeyIntRef(Integration, scope, old.IntegrationID, e)
	if err != nil {
		return err
	}

	err = e.restoreEventObject(&old, exporter, scope)
	if err != nil {
		return err
	}

	_, err = store.CreateEvent(old)
	return err
}

func eventObjectTypeToEntityName(t db.EventObjectType) (string, bool) {
	switch t {
	case db.EventTask:
		return Task, true
	case db.EventRepository:
		return Repository, true
	case db.EventEnvironment:
		return Environment, true
	case db.EventInventory:
		return Inventory, true
	case db.EventKey:
		return AccessKey, true
	case db.EventProject:
		return Project, true
	case db.EventSchedule:
		return Schedule, true
	case db.EventTemplate:
		return Template, true
	case db.EventUser:
		return User, true
	case db.EventView:
		return View, true
	case db.EventIntegration:
		return Integration, true
	case db.EventIntegrationExtractValue:
		return IntegrationExtractValue, true
	case db.EventIntegrationMatcher:
		return IntegrationMatcher, true
	default:
		return "", false
	}
}

func getScope(objectType, scope string) string {
	switch objectType {
	case Project:
		return GlobalScope
	case User:
		return GlobalScope
	}

	return scope
}

func (e *EventExporter) restoreEventObject(event *db.Event, exporter DataExporter, scope string) (err error) {
	if event.ObjectType != nil {
		entityName, ok := eventObjectTypeToEntityName(*event.ObjectType)
		if !ok {
			event.ObjectID = nil
			e.onError(fmt.Sprintf("Unknown event object type: %s", *event.ObjectType))
		} else {
			event.ObjectID, err = exporter.getNewKeyIntRef(entityName, getScope(entityName, scope), event.ObjectID, e)
			if err != nil {
				event.ObjectID = nil
				e.onError(fmt.Sprintf("Unable to restore event object %s, %s", entityName, err.Error()))
			}
		}
	}
	return nil
}

func (e *EventExporter) exportDependsOn() []string {
	return []string{Project, User}
}

func (e *EventExporter) importDependsOn() []string {
	return []string{Project, User, Integration, AccessKey, Schedule, Environment, Template, Task, Inventory, Repository, View}
}

func (e *EventExporter) getName() string {
	return Event
}
