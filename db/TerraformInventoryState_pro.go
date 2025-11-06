package db

import (
	"reflect"
	"time"
)

type TerraformInventoryState struct {
	ID          int       `db:"id" json:"id"`
	Created     time.Time `db:"created" json:"created"`
	TaskID      *int      `db:"task_id" json:"task_id,omitempty"`
	ProjectID   int       `db:"project_id" json:"project_id"`
	InventoryID int       `db:"inventory_id" json:"inventory_id"`
	State       string    `db:"state" json:"state,omitempty"`
}

var TerraformInventoryStateProps = ObjectProps{
	TableName:            "project__terraform_inventory_state",
	Type:                 reflect.TypeOf(TerraformInventoryState{}),
	PrimaryColumnName:    "id",
	SortableColumns:      []string{"created"},
	DefaultSortingColumn: "created",
	SortInverted:         true,
}
