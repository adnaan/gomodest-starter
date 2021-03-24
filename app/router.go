package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/hako/branca"

	"github.com/adnaan/authn"

	"github.com/adnaan/gomodest-starter/app/gen/models"

	"github.com/go-playground/form"

	"github.com/stripe/stripe-go/v72"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"

	"github.com/go-chi/httplog"

	rl "github.com/adnaan/renderlayout"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/mattn/go-sqlite3"
)

type Context struct {
	authn       *authn.API
	cfg         Config
	formDecoder *form.Decoder
	db          *models.Client
	branca      *branca.Branca
}

type APIRoute struct {
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

func Router(ctx context.Context, cfg Config) chi.Router {
	//driver := "postgres"
	//dataSource := "host=0.0.0.0 port=5432 user=gomodest-starter dbname=gomodest-starter sslmode=disable"
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
		cfg:         cfg,
		formDecoder: form.NewDecoder(),
		branca:      branca.NewBranca(cfg.APIMasterSecret),
	}

	authnConfig := authn.Config{
		Driver:        cfg.Driver,
		Datasource:    cfg.DataSource,
		SessionSecret: cfg.SessionSecret,
		SendMail:      sendEmailFunc(cfg),
		GothProviders: []goth.Provider{
			google.New(
				cfg.GoogleClientID,
				cfg.GoogleSecret,
				fmt.Sprintf("%s/auth/callback?provider=google", cfg.Domain),
				"email", "profile",
			),
		},
	}

	appCtx.authn = authn.New(ctx, authnConfig)

	// logger
	logger := httplog.NewLogger(cfg.Name,
		httplog.Options{
			JSON:     cfg.LogFormatJSON,
			LogLevel: cfg.LogLevel,
		})

	index, err := rl.New(
		rl.Layout("index"),
		rl.DisableCache(true),
		rl.Debug(false),
		rl.DefaultData(defaultPageHandler(appCtx)),
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
		r.Post("/webhook/{source}", handleWebhook(appCtx))
		r.Get("/", index("home"))
		r.Get("/signup", index("account/signup"))
		r.Post("/signup", index("account/signup", signupPageSubmit(appCtx)))
		r.Get("/confirm/{token}", index("account/confirmed", confirmEmailPage(appCtx)))

		r.Get("/login", index("account/login", loginPage(appCtx)))
		r.Post("/login", index("account/login", loginPageSubmit(appCtx)))
		r.Get("/auth/callback", index("account/login", gothAuthCallbackPage(appCtx)))
		r.Get("/auth", index("account/login", gothAuthPage(appCtx)))

		r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
			acc, err := appCtx.authn.CurrentAccount(r)
			if err != nil {
				log.Println("err logging out ", err)
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
			acc.Logout(w, r)
		})
		r.Get("/magic-link-sent", index("account/magic"))
		r.Get("/magic-login/{otp}", index("account/login", magicLinkLoginConfirm(appCtx)))

		r.Get("/forgot", index("account/forgot"))
		r.Post("/forgot", index("account/forgot", forgotPageSubmit(appCtx)))
		r.Get("/reset/{token}", index("account/reset"))
		r.Post("/reset/{token}", index("account/reset", resetPageSubmit(appCtx)))
		r.Get("/change/{token}", index("account/changed", confirmEmailChangePage(appCtx)))
	})

	// authenticated
	r.Route("/account", func(r chi.Router) {
		r.Use(appCtx.authn.IsAuthenticated)
		r.Get("/", index("account/main", accountPage(appCtx)))
		r.Post("/", index("account/main", accountPageSubmit(appCtx)))
		r.Post("/delete", index("account/main", deleteAccount(appCtx)))

		r.Post("/checkout", handleCreateCheckoutSession(appCtx))
		r.Get("/checkout/success", handleCheckoutSuccess(appCtx))
		r.Get("/checkout/cancel", handleCheckoutCancel(appCtx))
		r.Get("/subscription/manage", handleManageSubscription(appCtx))
	})

	r.Route("/app", func(r chi.Router) {
		r.Use(appCtx.authn.IsAuthenticated)
		r.Get("/", index("app", appPage(appCtx), listTasks(appCtx)))
		r.Post("/tasks/new", index("app", createNewTask(appCtx), listTasks(appCtx)))
		r.Post("/tasks/{id}/edit", index("app", editTask(appCtx), listTasks(appCtx)))
		r.Post("/tasks/{id}/delete", index("app", deleteTask(appCtx), listTasks(appCtx)))
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(appCtx.authn.IsAuthenticated)
		r.Use(middleware.AllowContentType("application/json"))
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", list(appCtx))
			r.Post("/", create(appCtx))
		})
		r.Route("/tasks/{id}", func(r chi.Router) {
			r.Put("/status", updateStatus(appCtx))
			r.Put("/text", updateText(appCtx))
			r.Delete("/", delete(appCtx))
		})
	})

	return r
}
