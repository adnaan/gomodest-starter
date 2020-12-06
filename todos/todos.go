package todos

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/go-chi/render"

	"github.com/asdine/storm/v3"
)

type Todo struct {
	ID   string `json:"id" storm:"id"`
	Text string `json:"text"`
}

func (t *Todo) Bind(r *http.Request) error {
	return nil
}

func (t *Todo) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func Add(w http.ResponseWriter, r *http.Request) {
	todo := &Todo{}
	if err := render.Bind(r, todo); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := storm.Open("todos.db")
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}
	defer db.Close()

	todo.ID = uuid.New().String()

	err = db.Save(todo)
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, todo)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	todo := &Todo{}
	if err := render.Bind(r, todo); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	fmt.Printf("%+v\n", todo)

	db, err := storm.Open("todos.db")
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}
	defer db.Close()

	err = db.DeleteStruct(todo)
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	render.Status(r, http.StatusOK)
}

func todoListResponse(todos []*Todo) []render.Renderer {
	list := make([]render.Renderer, 0)
	for _, todo := range todos {
		list = append(list, todo)
	}
	return list
}

func List(w http.ResponseWriter, r *http.Request) {
	db, err := storm.Open("todos.db")
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}
	defer db.Close()

	var todos []*Todo
	err = db.All(&todos)
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	if err := render.RenderList(w, r, todoListResponse(todos)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusOK)
}
