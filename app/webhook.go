package app

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/render"

	"github.com/stripe/stripe-go/v72/webhook"

	"github.com/go-chi/chi"
)

func handleWebhook(appCtx Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		source := chi.URLParam(r, "source")

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("ioutil.ReadAll: %v", err)
			render.Status(r, http.StatusBadRequest)
			return
		}

		switch source {
		case "stripe":
			event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), appCtx.cfg.StripeWebhookSecret)
			if err != nil {
				log.Printf("webhook.ConstructEvent: %v", err)
				render.Status(r, http.StatusBadRequest)
				return
			}

			if event.Type != "checkout.session.completed" {
				return
			}

			log.Printf("webhook %+v\n", event)
		}

	}
}
