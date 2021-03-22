package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"

	"github.com/lithammer/shortuuid/v3"

	"github.com/go-chi/render"

	"github.com/adnaan/authn"
	"github.com/adnaan/gomodest/app/gen/models/task"
)

func List(t Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := authn.AccountIDFromContext(r)
		tasks, err := t.db.Task.Query().Where(task.Owner(userID)).All(t.ctx)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		render.JSON(w, r, tasks)
	}
}

func Create(t Context) http.HandlerFunc {
	type req struct {
		Text string `json:"text"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(req)
		userID := authn.AccountIDFromContext(r)
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		newTask, err := t.db.Task.Create().
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

func UpdateStatus(t Context) http.HandlerFunc {
	type req struct {
		Status string `json:"status"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(req)
		id := chi.URLParam(r, "id")

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		updatedTask, err := t.db.Task.
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

func UpdateText(t Context) http.HandlerFunc {
	type req struct {
		Text string `json:"text"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(req)
		id := chi.URLParam(r, "id")

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		updatedTask, err := t.db.Task.
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

func Delete(t Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		err := t.db.Task.DeleteOneID(id).Exec(t.ctx)
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
