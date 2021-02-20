package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/adnaan/gomodest/app/internal/models"

	"github.com/go-playground/form"

	"github.com/go-chi/render"

	"github.com/stripe/stripe-go/v72"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"

	"github.com/adnaan/users"

	"github.com/go-chi/httplog"

	rl "github.com/adnaan/renderlayout"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Context struct {
	users       *users.API
	cfg         Config
	formDecoder *form.Decoder
	db          *models.Client
	ctx         context.Context
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

	db, err := models.Open(cfg.Driver, cfg.DataSource)
	if err != nil {
		panic(err)
	}
	if err := db.Schema.Create(ctx); err != nil {
		panic(err)
	}

	appCtx := Context{
		db:          db,
		ctx:         ctx,
		cfg:         cfg,
		formDecoder: form.NewDecoder(),
	}

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
			"owner": ownerRole(appCtx),
		},
	}
	usersAPI, err := users.NewDefaultAPI(ctx, defaultUsersConfig)
	if err != nil {
		log.Fatal(err)
	}

	appCtx.users = usersAPI

	// logger
	logger := httplog.NewLogger(cfg.Name,
		httplog.Options{
			JSON:     cfg.LogFormatJSON,
			LogLevel: cfg.LogLevel,
		})

	indexLayout, err := rl.New(
		rl.Layout("index"),
		rl.DisableCache(true),
		rl.DefaultHandler(defaultPageHandler(appCtx)),
		rl.ErrorKey("userError"),
	)

	if err != nil {
		log.Fatal(err)
	}

	// middlewares
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(middleware.Heartbeat(cfg.HealthPath))
	r.Use(middleware.Recoverer)
	r.Use(httplog.RequestLogger(logger))

	r.NotFound(indexLayout.HandleStatic("404"))
	// public
	r.Route("/", func(r chi.Router) {
		//r.Use(setDefaultPageData(appCtx))

		r.Post("/webhook/{source}", handleWebhook(appCtx))
		r.Get("/", indexLayout.HandleStatic("home"))
		r.Get("/signup", indexLayout.HandleStatic("signup"))
		r.Post("/signup", indexLayout.Handle("signup", signupPageSubmit(appCtx)))

		r.Get("/confirm/{token}", indexLayout.Handle("confirmed", confirmEmailPage(appCtx)))

		r.Get("/login", indexLayout.Handle("login", loginPage(appCtx)))
		r.Post("/login", indexLayout.Handle("login", loginPageSubmit(appCtx)))
		r.Get("/auth/callback", indexLayout.Handle("login", gothAuthCallbackPage(appCtx)))
		r.Get("/auth", indexLayout.Handle("login", gothAuthPage(appCtx)))

		r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
			provider := r.URL.Query().Get("provider")
			if provider != "" {
				usersAPI.HandleGothLogout(w, r)
				return
			}
			usersAPI.Logout(w, r)
		})
		r.Get("/magic-link-sent", indexLayout.HandleStatic("magic"))
		r.Get("/magic-login/{otp}", indexLayout.Handle("login", magicLinkLoginConfirm(appCtx)))

		r.Get("/forgot", indexLayout.HandleStatic("forgot"))
		r.Post("/forgot", indexLayout.Handle("forgot", forgotPageSubmit(appCtx)))
		r.Get("/reset/{token}", indexLayout.HandleStatic("reset"))
		r.Post("/reset/{token}", indexLayout.Handle("reset", resetPageSubmit(appCtx)))
		r.Get("/change/{token}", indexLayout.Handle("changed", confirmEmailChangePage(appCtx)))
	})

	// authenticated
	r.Route("/account", func(r chi.Router) {
		r.Use(usersAPI.IsAuthenticated)
		r.Get("/", indexLayout.Handle("account", accountPage(appCtx)))
		r.Post("/", indexLayout.Handle("account", accountPageSubmit(appCtx)))
		r.Post("/delete", indexLayout.Handle("account", deleteAccount(appCtx)))

		r.Post("/checkout", handleCreateCheckoutSession(appCtx))
		r.Get("/checkout/success", handleCheckoutSuccess(appCtx))
		r.Get("/checkout/cancel", handleCheckoutCancel(appCtx))
		r.Get("/subscription/manage", handleManageSubscription(appCtx))
	})

	r.Route("/app", func(r chi.Router) {
		r.Use(usersAPI.IsAuthenticated)
		r.Get("/", indexLayout.Handle("app", appPage(appCtx)))
		r.Post("/tasks/new", indexLayout.Handle("new_task", createNewTaskSubmit(appCtx)))
		r.Get("/tasks/delete/{id}", indexLayout.Handle("app", deleteTaskSubmit(appCtx)))
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
			r.Get("/", List(appCtx))
			r.Post("/", Create(appCtx))
		})
		r.With(authz).Route("/tasks/{id}", func(r chi.Router) {
			r.Put("/status", UpdateStatus(appCtx))
			r.Put("/text", UpdateText(appCtx))
			r.Delete("/", Delete(appCtx))
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
