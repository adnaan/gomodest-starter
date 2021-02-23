// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/adnaan/gomodest/app/gen/models/task"
)

// TaskCreate is the builder for creating a Task entity.
type TaskCreate struct {
	config
	mutation *TaskMutation
	hooks    []Hook
}

// SetOwner sets the "owner" field.
func (tc *TaskCreate) SetOwner(s string) *TaskCreate {
	tc.mutation.SetOwner(s)
	return tc
}

// SetText sets the "text" field.
func (tc *TaskCreate) SetText(s string) *TaskCreate {
	tc.mutation.SetText(s)
	return tc
}

// SetStatus sets the "status" field.
func (tc *TaskCreate) SetStatus(t task.Status) *TaskCreate {
	tc.mutation.SetStatus(t)
	return tc
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (tc *TaskCreate) SetNillableStatus(t *task.Status) *TaskCreate {
	if t != nil {
		tc.SetStatus(*t)
	}
	return tc
}

// SetCreatedAt sets the "created_at" field.
func (tc *TaskCreate) SetCreatedAt(t time.Time) *TaskCreate {
	tc.mutation.SetCreatedAt(t)
	return tc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (tc *TaskCreate) SetNillableCreatedAt(t *time.Time) *TaskCreate {
	if t != nil {
		tc.SetCreatedAt(*t)
	}
	return tc
}

// SetUpdatedAt sets the "updated_at" field.
func (tc *TaskCreate) SetUpdatedAt(t time.Time) *TaskCreate {
	tc.mutation.SetUpdatedAt(t)
	return tc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (tc *TaskCreate) SetNillableUpdatedAt(t *time.Time) *TaskCreate {
	if t != nil {
		tc.SetUpdatedAt(*t)
	}
	return tc
}

// SetID sets the "id" field.
func (tc *TaskCreate) SetID(s string) *TaskCreate {
	tc.mutation.SetID(s)
	return tc
}

// Mutation returns the TaskMutation object of the builder.
func (tc *TaskCreate) Mutation() *TaskMutation {
	return tc.mutation
}

// Save creates the Task in the database.
func (tc *TaskCreate) Save(ctx context.Context) (*Task, error) {
	var (
		err  error
		node *Task
	)
	tc.defaults()
	if len(tc.hooks) == 0 {
		if err = tc.check(); err != nil {
			return nil, err
		}
		node, err = tc.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*TaskMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = tc.check(); err != nil {
				return nil, err
			}
			tc.mutation = mutation
			node, err = tc.sqlSave(ctx)
			mutation.done = true
			return node, err
		})
		for i := len(tc.hooks) - 1; i >= 0; i-- {
			mut = tc.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, tc.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX calls Save and panics if Save returns an error.
func (tc *TaskCreate) SaveX(ctx context.Context) *Task {
	v, err := tc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// defaults sets the default values of the builder before save.
func (tc *TaskCreate) defaults() {
	if _, ok := tc.mutation.Status(); !ok {
		v := task.DefaultStatus
		tc.mutation.SetStatus(v)
	}
	if _, ok := tc.mutation.CreatedAt(); !ok {
		v := task.DefaultCreatedAt()
		tc.mutation.SetCreatedAt(v)
	}
	if _, ok := tc.mutation.UpdatedAt(); !ok {
		v := task.DefaultUpdatedAt()
		tc.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (tc *TaskCreate) check() error {
	if _, ok := tc.mutation.Owner(); !ok {
		return &ValidationError{Name: "owner", err: errors.New("models: missing required field \"owner\"")}
	}
	if _, ok := tc.mutation.Text(); !ok {
		return &ValidationError{Name: "text", err: errors.New("models: missing required field \"text\"")}
	}
	if v, ok := tc.mutation.Status(); ok {
		if err := task.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf("models: validator failed for field \"status\": %w", err)}
		}
	}
	if _, ok := tc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New("models: missing required field \"created_at\"")}
	}
	if _, ok := tc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New("models: missing required field \"updated_at\"")}
	}
	return nil
}

func (tc *TaskCreate) sqlSave(ctx context.Context) (*Task, error) {
	_node, _spec := tc.createSpec()
	if err := sqlgraph.CreateNode(ctx, tc.driver, _spec); err != nil {
		if cerr, ok := isSQLConstraintError(err); ok {
			err = cerr
		}
		return nil, err
	}
	return _node, nil
}

func (tc *TaskCreate) createSpec() (*Task, *sqlgraph.CreateSpec) {
	var (
		_node = &Task{config: tc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: task.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeString,
				Column: task.FieldID,
			},
		}
	)
	if id, ok := tc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := tc.mutation.Owner(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: task.FieldOwner,
		})
		_node.Owner = value
	}
	if value, ok := tc.mutation.Text(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: task.FieldText,
		})
		_node.Text = value
	}
	if value, ok := tc.mutation.Status(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeEnum,
			Value:  value,
			Column: task.FieldStatus,
		})
		_node.Status = value
	}
	if value, ok := tc.mutation.CreatedAt(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: task.FieldCreatedAt,
		})
		_node.CreatedAt = value
	}
	if value, ok := tc.mutation.UpdatedAt(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: task.FieldUpdatedAt,
		})
		_node.UpdatedAt = value
	}
	return _node, _spec
}

// TaskCreateBulk is the builder for creating many Task entities in bulk.
type TaskCreateBulk struct {
	config
	builders []*TaskCreate
}

// Save creates the Task entities in the database.
func (tcb *TaskCreateBulk) Save(ctx context.Context) ([]*Task, error) {
	specs := make([]*sqlgraph.CreateSpec, len(tcb.builders))
	nodes := make([]*Task, len(tcb.builders))
	mutators := make([]Mutator, len(tcb.builders))
	for i := range tcb.builders {
		func(i int, root context.Context) {
			builder := tcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*TaskMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				nodes[i], specs[i] = builder.createSpec()
				var err error
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, tcb.builders[i+1].mutation)
				} else {
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, tcb.driver, &sqlgraph.BatchCreateSpec{Nodes: specs}); err != nil {
						if cerr, ok := isSQLConstraintError(err); ok {
							err = cerr
						}
					}
				}
				mutation.done = true
				if err != nil {
					return nil, err
				}
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, tcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (tcb *TaskCreateBulk) SaveX(ctx context.Context) []*Task {
	v, err := tcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}
