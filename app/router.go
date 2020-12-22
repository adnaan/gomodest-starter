package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/render"

	"github.com/stripe/stripe-go/v72"

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

type Context struct {
	users      *users.API
	viewEngine *goview.ViewEngine
	pageData   goview.M
	cfg        Config
}

type APIRoute struct {
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

func Router(ctx context.Context, cfg Config) chi.Router {
	//driver := "postgres"
	//dataSource := "host=0.0.0.0 port=5432 user=gomodest dbname=gomodest sslmode=disable"
	stripe.Key = cfg.StripeSecretKey

	tasksCtx := NewTasksContext(ctx, cfg)
	defaultUsersConfig := users.Config{
		Driver:          cfg.Driver,
		Datasource:      cfg.DataSource,
		SessionSecret:   cfg.SessionSecret,
		APIMasterSecret: cfg.APIMasterSecret,
		SendMail:        sendEmailFunc(cfg),
		GothProviders: []goth.Provider{
			google.New(cfg.GoogleClientID, cfg.GoogleSecret, fmt.Sprintf("%s/auth/callback?provider=google", cfg.Domain), "email", "profile"),
		},
		Roles: map[string][]users.Permission{
			"owner": ownerRole(tasksCtx),
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

	appCtx := Context{
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
	r.Use(httplog.RequestLogger(logger))
	//r.Use(setDefaultPageData(appCtx))

	// app

	r.NotFound(rr("404"))
	// public
	r.Route("/", func(r chi.Router) {
		r.Use(setDefaultPageData(appCtx))
		r.Get("/", rr("home"))
		r.Post("/webhook/{source}", handleWebhook(appCtx))

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
	})

	// authenticated
	r.Route("/account", func(r chi.Router) {
		r.Use(usersAPI.IsAuthenticated)
		r.Use(setDefaultPageData(appCtx))
		r.Get("/", rr("account", accountPage))
		r.Post("/", rr("account", accountPageSubmit))
		r.Post("/delete", rr("account", deleteAccount))

		r.Post("/checkout", handleCreateCheckoutSession(appCtx))
		r.Get("/checkout/success", handleCheckoutSuccess(appCtx))
		r.Get("/checkout/cancel", handleCheckoutCancel(appCtx))
		r.Get("/subscription/manage", handleManageSubscription(appCtx))
	})

	r.Route("/app", func(r chi.Router) {
		r.Use(usersAPI.IsAuthenticated)
		r.Use(setDefaultPageData(appCtx))
		r.Get("/", rr("app", appPage))
	})

	authz := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// GET /api/tasks/ => get:api:tasks
			action := buildRestAction(r.Method, r.URL.Path)
			target := chi.URLParam(r, "id")
			if target == "" {
				target = "*"
			}
			allow, err := usersAPI.Can(r, action, target)
			if !allow || err != nil {
				log.Printf("Can err: %v\n", err)
				render.Render(w, r, ErrUnauthorized(err))
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	r.Route("/api", func(r chi.Router) {
		r.Use(usersAPI.IsAuthenticated)
		r.Use(middleware.AllowContentType("application/json"))
		r.With(authz).Route("/tasks", func(r chi.Router) {
			r.Get("/", List(tasksCtx))
			r.Post("/", Create(tasksCtx))
		})
		r.With(authz).Route("/tasks/{id}", func(r chi.Router) {
			r.Put("/status", UpdateStatus(tasksCtx))
			r.Put("/text", UpdateText(tasksCtx))
			r.Delete("/", Delete(tasksCtx))
		})

	})

	return r
}

func buildRestAction(method, path string) string {
	return fmt.Sprintf("%s%s", strings.ToLower(method),
		strings.ToLower(
			strings.TrimRight(
				strings.Replace(path, "/", ":", -1), ":")))
}
