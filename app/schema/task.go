package schema

import (
	"time"

	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

func (Task) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "tasks"},
	}
}

// Fields of the Task.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.String("owner"),
		field.Text("text"),
		field.Enum("status").Values("todo", "inprogress", "done").Default("todo").Optional(),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Task.
func (Task) Edges() []ent.Edge {
	return nil
}
