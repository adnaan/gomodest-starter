package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"

	"github.com/lithammer/shortuuid/v3"

	"github.com/go-chi/render"

	"github.com/adnaan/authn"
	"github.com/adnaan/gomodest-starter/app/gen/models/task"
)

func list(t Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := authn.AccountIDFromContext(r)
		tasks, err := t.db.Task.Query().Where(task.Owner(userID)).All(r.Context())
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		render.JSON(w, r, tasks)
	}
}

func create(t Context) http.HandlerFunc {
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
			Save(r.Context())
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}
		render.JSON(w, r, newTask)
	}
}

func updateStatus(t Context) http.HandlerFunc {
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
			Save(r.Context())
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}
		render.JSON(w, r, updatedTask)
	}
}

func updateText(t Context) http.HandlerFunc {
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
			Save(r.Context())
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}
		render.JSON(w, r, updatedTask)
	}
}

func delete(t Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		err := t.db.Task.DeleteOneID(id).Exec(r.Context())
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
