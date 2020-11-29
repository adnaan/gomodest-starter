package pkg

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/foolin/goview"

	"github.com/adnaan/users"

	"github.com/go-chi/httplog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const (
	appCtxDataKey = "app_ctx_data"
)

type AppContext struct {
	users      *users.API
	viewEngine *goview.ViewEngine
}

// NewRouter ...
func NewRouter() http.Handler {
	ctx := context.Background()
	driver := "postgres"
	dataSource := "host=0.0.0.0 port=5432 user=gomodest dbname=gomodest sslmode=disable"
	usersAPI, err := users.NewDefaultAPI(ctx, driver, dataSource, "mycookiesecret")
	if err != nil {
		log.Fatal(err)
	}

	// logger
	logger := httplog.NewLogger("gomodest",
		httplog.Options{
			JSON:     true,
			LogLevel: "ERROR",
		})

	indexLayout, err := viewEngine("index")
	if err != nil {
		panic(err)
	}

	appCtx := AppContext{
		users:      usersAPI,
		viewEngine: indexLayout,
	}

	// middlewares
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.Recoverer)
	r.Use(setDefaultPageData(appCtx))
	r.Use(httplog.RequestLogger(logger))

	// routes
	// public
	r.NotFound(renderPage(appCtx, "404", nil))
	r.Get("/", renderPage(appCtx, "home", nil))
	r.Get("/login", renderPage(appCtx, "login", loginPage))
	r.Get("/signup", renderPage(appCtx, "signup", nil))

	r.Post("/signup", usersAPI.Signup)
	r.Get("/confirm/{token}", usersAPI.ConfirmEmail)
	r.Get("/change/{token}", usersAPI.ConfirmEmailChange)
	r.Post("/login", usersAPI.Login)
	r.Get("/logout", usersAPI.Logout)

	// authenticated
	r.Route("/account", func(r chi.Router) {
		r.Use(usersAPI.IsAuthenticated)
		r.Get("/", renderPage(appCtx, "account", accountPage))
		r.Post("/", renderPage(appCtx, "account", accountPageSubmit))
	})

	r.Route("/app", func(r chi.Router) {
		r.Use(usersAPI.IsAuthenticated)
		r.Get("/", renderPage(appCtx, "app", appPage))
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(usersAPI.IsAuthenticated)
		r.Use(middleware.AllowContentType("application/json"))
		r.Get("/todos", listTodos)
		r.Post("/todos", addTodo)
		r.Delete("/todos", deleteTodo)
	})

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web", "dist"))
	fileServer(r, "/static", filesDir)

	return r
}
