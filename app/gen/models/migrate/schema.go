// Code generated (@generated) by entc, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// TasksColumns holds the columns for the "tasks" table.
	TasksColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString},
		{Name: "owner", Type: field.TypeString},
		{Name: "text", Type: field.TypeString, Size: 2147483647},
		{Name: "status", Type: field.TypeEnum, Nullable: true, Enums: []string{"todo", "inprogress", "done"}, Default: "todo"},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
	}
	// TasksTable holds the schema information for the "tasks" table.
	TasksTable = &schema.Table{
		Name:        "tasks",
		Columns:     TasksColumns,
		PrimaryKey:  []*schema.Column{TasksColumns[0]},
		ForeignKeys: []*schema.ForeignKey{},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		TasksTable,
	}
)

func init() {
	TasksTable.Annotation = &entsql.Annotation{
		Table: "tasks",
	}
}
