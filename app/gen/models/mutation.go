// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/adnaan/gomodest/app/gen/models/predicate"
	"github.com/adnaan/gomodest/app/gen/models/task"

	"entgo.io/ent"
)

const (
	// Operation types.
	OpCreate    = ent.OpCreate
	OpDelete    = ent.OpDelete
	OpDeleteOne = ent.OpDeleteOne
	OpUpdate    = ent.OpUpdate
	OpUpdateOne = ent.OpUpdateOne

	// Node types.
	TypeTask = "Task"
)

// TaskMutation represents an operation that mutates the Task nodes in the graph.
type TaskMutation struct {
	config
	op            Op
	typ           string
	id            *string
	owner         *string
	text          *string
	status        *task.Status
	created_at    *time.Time
	updated_at    *time.Time
	clearedFields map[string]struct{}
	done          bool
	oldValue      func(context.Context) (*Task, error)
	predicates    []predicate.Task
}

var _ ent.Mutation = (*TaskMutation)(nil)

// taskOption allows management of the mutation configuration using functional options.
type taskOption func(*TaskMutation)

// newTaskMutation creates new mutation for the Task entity.
func newTaskMutation(c config, op Op, opts ...taskOption) *TaskMutation {
	m := &TaskMutation{
		config:        c,
		op:            op,
		typ:           TypeTask,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withTaskID sets the ID field of the mutation.
func withTaskID(id string) taskOption {
	return func(m *TaskMutation) {
		var (
			err   error
			once  sync.Once
			value *Task
		)
		m.oldValue = func(ctx context.Context) (*Task, error) {
			once.Do(func() {
				if m.done {
					err = fmt.Errorf("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Task.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withTask sets the old Task of the mutation.
func withTask(node *Task) taskOption {
	return func(m *TaskMutation) {
		m.oldValue = func(context.Context) (*Task, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m TaskMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m TaskMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, fmt.Errorf("models: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// SetID sets the value of the id field. Note that this
// operation is only accepted on creation of Task entities.
func (m *TaskMutation) SetID(id string) {
	m.id = &id
}

// ID returns the ID value in the mutation. Note that the ID
// is only available if it was provided to the builder.
func (m *TaskMutation) ID() (id string, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// SetOwner sets the "owner" field.
func (m *TaskMutation) SetOwner(s string) {
	m.owner = &s
}

// Owner returns the value of the "owner" field in the mutation.
func (m *TaskMutation) Owner() (r string, exists bool) {
	v := m.owner
	if v == nil {
		return
	}
	return *v, true
}

// OldOwner returns the old "owner" field's value of the Task entity.
// If the Task object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *TaskMutation) OldOwner(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, fmt.Errorf("OldOwner is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, fmt.Errorf("OldOwner requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldOwner: %w", err)
	}
	return oldValue.Owner, nil
}

// ResetOwner resets all changes to the "owner" field.
func (m *TaskMutation) ResetOwner() {
	m.owner = nil
}

// SetText sets the "text" field.
func (m *TaskMutation) SetText(s string) {
	m.text = &s
}

// Text returns the value of the "text" field in the mutation.
func (m *TaskMutation) Text() (r string, exists bool) {
	v := m.text
	if v == nil {
		return
	}
	return *v, true
}

// OldText returns the old "text" field's value of the Task entity.
// If the Task object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *TaskMutation) OldText(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, fmt.Errorf("OldText is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, fmt.Errorf("OldText requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldText: %w", err)
	}
	return oldValue.Text, nil
}

// ResetText resets all changes to the "text" field.
func (m *TaskMutation) ResetText() {
	m.text = nil
}

// SetStatus sets the "status" field.
func (m *TaskMutation) SetStatus(t task.Status) {
	m.status = &t
}

// Status returns the value of the "status" field in the mutation.
func (m *TaskMutation) Status() (r task.Status, exists bool) {
	v := m.status
	if v == nil {
		return
	}
	return *v, true
}

// OldStatus returns the old "status" field's value of the Task entity.
// If the Task object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *TaskMutation) OldStatus(ctx context.Context) (v task.Status, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, fmt.Errorf("OldStatus is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, fmt.Errorf("OldStatus requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldStatus: %w", err)
	}
	return oldValue.Status, nil
}

// ClearStatus clears the value of the "status" field.
func (m *TaskMutation) ClearStatus() {
	m.status = nil
	m.clearedFields[task.FieldStatus] = struct{}{}
}

// StatusCleared returns if the "status" field was cleared in this mutation.
func (m *TaskMutation) StatusCleared() bool {
	_, ok := m.clearedFields[task.FieldStatus]
	return ok
}

// ResetStatus resets all changes to the "status" field.
func (m *TaskMutation) ResetStatus() {
	m.status = nil
	delete(m.clearedFields, task.FieldStatus)
}

// SetCreatedAt sets the "created_at" field.
func (m *TaskMutation) SetCreatedAt(t time.Time) {
	m.created_at = &t
}

// CreatedAt returns the value of the "created_at" field in the mutation.
func (m *TaskMutation) CreatedAt() (r time.Time, exists bool) {
	v := m.created_at
	if v == nil {
		return
	}
	return *v, true
}

// OldCreatedAt returns the old "created_at" field's value of the Task entity.
// If the Task object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *TaskMutation) OldCreatedAt(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, fmt.Errorf("OldCreatedAt is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, fmt.Errorf("OldCreatedAt requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldCreatedAt: %w", err)
	}
	return oldValue.CreatedAt, nil
}

// ResetCreatedAt resets all changes to the "created_at" field.
func (m *TaskMutation) ResetCreatedAt() {
	m.created_at = nil
}

// SetUpdatedAt sets the "updated_at" field.
func (m *TaskMutation) SetUpdatedAt(t time.Time) {
	m.updated_at = &t
}

// UpdatedAt returns the value of the "updated_at" field in the mutation.
func (m *TaskMutation) UpdatedAt() (r time.Time, exists bool) {
	v := m.updated_at
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdatedAt returns the old "updated_at" field's value of the Task entity.
// If the Task object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *TaskMutation) OldUpdatedAt(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, fmt.Errorf("OldUpdatedAt is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, fmt.Errorf("OldUpdatedAt requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldUpdatedAt: %w", err)
	}
	return oldValue.UpdatedAt, nil
}

// ResetUpdatedAt resets all changes to the "updated_at" field.
func (m *TaskMutation) ResetUpdatedAt() {
	m.updated_at = nil
}

// Op returns the operation name.
func (m *TaskMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (Task).
func (m *TaskMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *TaskMutation) Fields() []string {
	fields := make([]string, 0, 5)
	if m.owner != nil {
		fields = append(fields, task.FieldOwner)
	}
	if m.text != nil {
		fields = append(fields, task.FieldText)
	}
	if m.status != nil {
		fields = append(fields, task.FieldStatus)
	}
	if m.created_at != nil {
		fields = append(fields, task.FieldCreatedAt)
	}
	if m.updated_at != nil {
		fields = append(fields, task.FieldUpdatedAt)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *TaskMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case task.FieldOwner:
		return m.Owner()
	case task.FieldText:
		return m.Text()
	case task.FieldStatus:
		return m.Status()
	case task.FieldCreatedAt:
		return m.CreatedAt()
	case task.FieldUpdatedAt:
		return m.UpdatedAt()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *TaskMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case task.FieldOwner:
		return m.OldOwner(ctx)
	case task.FieldText:
		return m.OldText(ctx)
	case task.FieldStatus:
		return m.OldStatus(ctx)
	case task.FieldCreatedAt:
		return m.OldCreatedAt(ctx)
	case task.FieldUpdatedAt:
		return m.OldUpdatedAt(ctx)
	}
	return nil, fmt.Errorf("unknown Task field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *TaskMutation) SetField(name string, value ent.Value) error {
	switch name {
	case task.FieldOwner:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetOwner(v)
		return nil
	case task.FieldText:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetText(v)
		return nil
	case task.FieldStatus:
		v, ok := value.(task.Status)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetStatus(v)
		return nil
	case task.FieldCreatedAt:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreatedAt(v)
		return nil
	case task.FieldUpdatedAt:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdatedAt(v)
		return nil
	}
	return fmt.Errorf("unknown Task field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *TaskMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *TaskMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *TaskMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown Task numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *TaskMutation) ClearedFields() []string {
	var fields []string
	if m.FieldCleared(task.FieldStatus) {
		fields = append(fields, task.FieldStatus)
	}
	return fields
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *TaskMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *TaskMutation) ClearField(name string) error {
	switch name {
	case task.FieldStatus:
		m.ClearStatus()
		return nil
	}
	return fmt.Errorf("unknown Task nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *TaskMutation) ResetField(name string) error {
	switch name {
	case task.FieldOwner:
		m.ResetOwner()
		return nil
	case task.FieldText:
		m.ResetText()
		return nil
	case task.FieldStatus:
		m.ResetStatus()
		return nil
	case task.FieldCreatedAt:
		m.ResetCreatedAt()
		return nil
	case task.FieldUpdatedAt:
		m.ResetUpdatedAt()
		return nil
	}
	return fmt.Errorf("unknown Task field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *TaskMutation) AddedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *TaskMutation) AddedIDs(name string) []ent.Value {
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *TaskMutation) RemovedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *TaskMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *TaskMutation) ClearedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *TaskMutation) EdgeCleared(name string) bool {
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *TaskMutation) ClearEdge(name string) error {
	return fmt.Errorf("unknown Task unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *TaskMutation) ResetEdge(name string) error {
	return fmt.Errorf("unknown Task edge %s", name)
}
