package app

import (
	"context"
	"net/http"
	"time"

	"github.com/adnaan/users"

	"github.com/go-chi/chi"

	"github.com/lithammer/shortuuid/v3"

	"github.com/go-chi/render"

	"github.com/adnaan/gomodest/app/internal/models"
	"github.com/adnaan/gomodest/app/internal/models/task"
)

type TasksContext struct {
	client *models.Client
	cfg    Config
	ctx    context.Context
}

func NewTasksContext(ctx context.Context, cfg Config) TasksContext {
	client, err := models.Open(cfg.Driver, cfg.DataSource)
	if err != nil {
		panic(err)
	}
	if err := client.Schema.Create(ctx); err != nil {
		panic(err)
	}

	return TasksContext{
		client: client,
		cfg:    cfg,
		ctx:    ctx,
	}
}

func List(t TasksContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(users.CtxUserIdKey).(string)
		tasks, err := t.client.Task.Query().Where(task.Owner(userID)).All(t.ctx)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		render.JSON(w, r, tasks)
	}
}

func Create(t TasksContext) http.HandlerFunc {
	var req struct {
		Text string `json:"text"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(users.CtxUserIdKey).(string)
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		newTask, err := t.client.Task.Create().
			SetID(shortuuid.New()).
			SetStatus(task.StatusInprogress).
			SetOwner(userID).
			SetText(req.Text).
			Save(t.ctx)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}
		render.JSON(w, r, newTask)
	}
}

func UpdateStatus(t TasksContext) http.HandlerFunc {
	var req struct {
		Status string `json:"status"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		updatedTask, err := t.client.Task.
			UpdateOneID(id).
			SetUpdatedAt(time.Now()).
			SetStatus(task.Status(req.Status)).
			Save(t.ctx)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}
		render.JSON(w, r, updatedTask)
	}
}

func UpdateText(t TasksContext) http.HandlerFunc {
	var req struct {
		Text string `json:"text"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		updatedTask, err := t.client.Task.
			UpdateOneID(id).
			SetUpdatedAt(time.Now()).
			SetText(req.Text).
			Save(t.ctx)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}
		render.JSON(w, r, updatedTask)
	}
}

func Delete(t TasksContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		err := t.client.Task.DeleteOneID(id).Exec(t.ctx)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}
		render.Status(r, http.StatusOK)
		render.JSON(w, r, struct {
			Success bool `json:"success"`
		}{
			Success: true,
		})
	}
}
