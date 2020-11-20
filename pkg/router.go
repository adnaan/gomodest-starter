package pkg

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/httplog"

	"github.com/gorilla/sessions"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	store *sessions.FilesystemStore
)

const (
	appCtxDataKey = "app_ctx_data"
)

// NewRouter ...
func NewRouter() http.Handler {

	store = sessions.NewFilesystemStore("", []byte("something-very-secret"))

	// logger
	logger := httplog.NewLogger("gomodest",
		httplog.Options{
			JSON:     true,
			LogLevel: "INFO",
		})

	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.Recoverer)
	r.Use(setDefaultPageData)
	r.Use(httplog.RequestLogger(logger))

	indexLayout, err := viewEngine("index")
	if err != nil {
		panic(err)
	}

	r.NotFound(renderPage(indexLayout, "404", nil))
	r.Get("/", renderPage(indexLayout, "home", nil))
	r.Get("/account", renderPage(indexLayout, "account", accountPage))
	r.Post("/account", renderPage(indexLayout, "account", accountPageSubmit))
	r.Post("/login", renderPage(indexLayout, "login", loginPageSubmit))
	r.Get("/login", renderPage(indexLayout, "login", nil))
	r.Get("/logout", logoutHandler)

	r.Route("/app", func(r chi.Router) {
		r.Use(isAuthenticated)
		r.Get("/", renderPage(indexLayout, "app", appPage))
	})

	r.Route("/todos", func(r chi.Router) {
		r.Use(isAuthenticatedAPI)
		r.Use(middleware.AllowContentType("application/json"))
		r.Get("/", listTodos)
		r.Post("/", addTodo)
		r.Delete("/", deleteTodo)
	})

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web", "dist"))
	fileServer(r, "/static", filesDir)

	return r

}
