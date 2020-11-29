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
	pageData   goview.M
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

	rr := newRenderer(appCtx)

	// middlewares
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.Recoverer)
	r.Use(setDefaultPageData(appCtx))
	r.Use(httplog.RequestLogger(logger))

	// routes
	// public
	r.NotFound(rr("404"))
	r.Get("/", rr("home"))

	r.Get("/signup", rr("signup"))
	r.Post("/signup", rr("signup", signupPageSubmit))

	r.Get("/confirm/{token}", rr("confirmed", confirmEmailPage))

	r.Get("/login", rr("login", loginPage))
	r.Post("/login", rr("login", loginPageSubmit))

	r.Get("/forgot", rr("forgot"))
	r.Post("/forgot", rr("forgot", forgotPageSubmit))
	r.Get("/reset/{token}", rr("reset"))
	r.Post("/reset/{token}", rr("reset", resetPageSubmit))
	r.Get("/change/{token}", rr("changed", confirmEmailChangePage))

	r.Get("/logout", usersAPI.Logout)

	// authenticated
	r.Route("/account", func(r chi.Router) {
		r.Use(usersAPI.IsAuthenticated)
		r.Get("/", rr("account", accountPage))
		r.Post("/", rr("account", accountPageSubmit))
	})

	r.Route("/app", func(r chi.Router) {
		r.Use(usersAPI.IsAuthenticated)
		r.Get("/", rr("app", appPage))
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
