package pkg

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"

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
	cfg        Config
}

func router(ctx context.Context, cfg Config) chi.Router {
	//driver := "postgres"
	//dataSource := "host=0.0.0.0 port=5432 user=gomodest dbname=gomodest sslmode=disable"

	if cfg.Host == "0.0.0.0" || cfg.Host == "localhost" {
		cfg.Domain = fmt.Sprintf("%s://%s:%d", cfg.Scheme, cfg.Host, cfg.Port)
	}

	defaultUsersConfig := users.Config{
		Driver:        cfg.Driver,
		Datasource:    cfg.DataSource,
		SessionSecret: cfg.SessionSecret,
		SendMail:      sendEmailFunc(cfg),
		GothProviders: []goth.Provider{
			google.New(cfg.GoogleClientID, cfg.GoogleSecret, fmt.Sprintf("%s/auth/callback?provider=google", cfg.Domain), "email", "profile"),
		},
	}
	usersAPI, err := users.NewDefaultAPI(ctx, defaultUsersConfig)
	if err != nil {
		log.Fatal(err)
	}

	// logger
	logger := httplog.NewLogger(cfg.Name,
		httplog.Options{
			JSON:     cfg.LogFormatJSON,
			LogLevel: cfg.LogLevel,
		})

	indexLayout, err := viewEngine(cfg, "index")
	if err != nil {
		panic(err)
	}

	appCtx := AppContext{
		users:      usersAPI,
		viewEngine: indexLayout,
		cfg:        cfg,
	}

	rr := newRenderer(appCtx)

	// middlewares
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(middleware.Heartbeat(cfg.HealthPath))
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
	r.Get("/auth/callback", rr("login", gothAuthCallbackPage))
	r.Get("/auth", rr("login", gothAuthPage))

	r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		provider := r.URL.Query().Get("provider")
		if provider != "" {
			usersAPI.HandleGothLogout(w, r)
			return
		}
		usersAPI.Logout(w, r)
	})

	r.Get("/magic-link-sent", rr("magic"))
	r.Get("/magic-login/{otp}", rr("login", magicLinkLoginConfirm))

	r.Get("/forgot", rr("forgot"))
	r.Post("/forgot", rr("forgot", forgotPageSubmit))
	r.Get("/reset/{token}", rr("reset"))
	r.Post("/reset/{token}", rr("reset", resetPageSubmit))
	r.Get("/change/{token}", rr("changed", confirmEmailChangePage))

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

	return r
}
