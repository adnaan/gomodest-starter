package app

import (
	"fmt"
	"net/http"

	"github.com/adnaan/authn"
	"github.com/adnaan/gomodest-starter/app/gen/models/task"
	rl "github.com/adnaan/renderlayout"
	"github.com/go-chi/chi"
	"github.com/lithammer/shortuuid/v3"
)

func listTasks(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		userID := authn.AccountIDFromContext(r)
		tasks, err := appCtx.db.Task.Query().Where(task.Owner(userID)).All(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return rl.D{
			"tasks": tasks,
		}, nil
	}
}

func createNewTask(appCtx Context) rl.Data {
	type req struct {
		Text string `json:"text"`
	}

	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		req := new(req)
		err := r.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		err = appCtx.formDecoder.Decode(req, r.Form)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		if req.Text == "" {
			return nil, fmt.Errorf("%w", fmt.Errorf("empty task"))
		}

		userID := authn.AccountIDFromContext(r)
		_, err = appCtx.db.Task.Create().
			SetID(shortuuid.New()).
			SetStatus(task.StatusInprogress).
			SetOwner(userID).
			SetText(req.Text).
			Save(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return nil, nil
	}
}

func deleteTask(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		id := chi.URLParam(r, "id")
		userID := authn.AccountIDFromContext(r)

		_, err := appCtx.db.Task.Delete().Where(task.And(
			task.Owner(userID), task.ID(id),
		)).Exec(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return nil, nil
	}
}

func editTask(appCtx Context) rl.Data {
	type req struct {
		Text string `json:"text"`
	}
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		req := new(req)
		err := r.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		err = appCtx.formDecoder.Decode(req, r.Form)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		if req.Text == "" {
			return nil, fmt.Errorf("%w", fmt.Errorf("empty task"))
		}

		id := chi.URLParam(r, "id")
		userID := authn.AccountIDFromContext(r)
		err = appCtx.db.Task.Update().Where(task.And(
			task.Owner(userID), task.ID(id),
		)).SetText(req.Text).Exec(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return nil, nil
	}
}
