package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/adnaan/gomodest/app/gen/models"

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

	index, err := rl.New(
		rl.Layout("index"),
		rl.DisableCache(true),
		rl.DefaultData(defaultPageHandler(appCtx)),
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

	r.NotFound(index("404"))
	// public
	r.Route("/", func(r chi.Router) {
		//r.Use(setDefaultPageData(appCtx))

		r.Post("/webhook/{source}", handleWebhook(appCtx))
		r.Get("/", index("home"))
		r.Get("/signup", index("signup"))
		r.Post("/signup", index("signup", signupPageSubmit(appCtx)))

		r.Get("/confirm/{token}", index("confirmed", confirmEmailPage(appCtx)))

		r.Get("/login", index("login", loginPage(appCtx)))
		r.Post("/login", index("login", loginPageSubmit(appCtx)))
		r.Get("/auth/callback", index("login", gothAuthCallbackPage(appCtx)))
		r.Get("/auth", index("login", gothAuthPage(appCtx)))

		r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
			provider := r.URL.Query().Get("provider")
			if provider != "" {
				usersAPI.HandleGothLogout(w, r)
				return
			}
			usersAPI.Logout(w, r)
		})
		r.Get("/magic-link-sent", index("magic"))
		r.Get("/magic-login/{otp}", index("login", magicLinkLoginConfirm(appCtx)))

		r.Get("/forgot", index("forgot"))
		r.Post("/forgot", index("forgot", forgotPageSubmit(appCtx)))
		r.Get("/reset/{token}", index("reset"))
		r.Post("/reset/{token}", index("reset", resetPageSubmit(appCtx)))
		r.Get("/change/{token}", index("changed", confirmEmailChangePage(appCtx)))
	})

	// authenticated
	r.Route("/account", func(r chi.Router) {
		r.Use(usersAPI.IsAuthenticated)
		r.Get("/", index("account", accountPage(appCtx)))
		r.Post("/", index("account", accountPageSubmit(appCtx)))
		r.Post("/delete", index("account", deleteAccount(appCtx)))

		r.Post("/checkout", handleCreateCheckoutSession(appCtx))
		r.Get("/checkout/success", handleCheckoutSuccess(appCtx))
		r.Get("/checkout/cancel", handleCheckoutCancel(appCtx))
		r.Get("/subscription/manage", handleManageSubscription(appCtx))
	})

	r.Route("/app", func(r chi.Router) {
		r.Use(usersAPI.IsAuthenticated)
		r.Get("/", index("app", appPage(appCtx), listTasks(appCtx)))
		r.Post("/tasks/new", index("app", createNewTask(appCtx), listTasks(appCtx)))
		r.Post("/tasks/{id}/edit", index("app", editTask(appCtx), listTasks(appCtx)))
		r.Post("/tasks/{id}/delete", index("app", deleteTask(appCtx), listTasks(appCtx)))
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
