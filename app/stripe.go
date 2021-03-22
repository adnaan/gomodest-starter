package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v72"
	portalsession "github.com/stripe/stripe-go/v72/billingportal/session"
	"github.com/stripe/stripe-go/v72/checkout/session"

	"github.com/go-chi/render"
)

// modified from https://github.com/stripe-samples/checkout-single-subscription/blob/master/server/go/server.go

const (
	billingIDKey      = "billing_id"
	currentPriceIDKey = "current_price_id"
)

type errResponse struct {
	Error string `json:"error"`
}

func handleCreateCheckoutSession(appCtx Context) http.HandlerFunc {

	type req struct {
		Price string `json:"price"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(req)
		account, err := appCtx.authn.CurrentAccount(r)
		if err != nil {
			log.Printf("authn.CurrentAccount: %v", err)
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, &errResponse{"unauthorized"})
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("json.NewDecoder.Decode: %v", err)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, &errResponse{err.Error()})
			return
		}

		params := &stripe.CheckoutSessionParams{
			CustomerEmail: stripe.String(account.Email()),
			SuccessURL:    stripe.String(appCtx.cfg.Domain + "/account/checkout/success?session_id={CHECKOUT_SESSION_ID}"),
			CancelURL:     stripe.String(appCtx.cfg.Domain + "/account/checkout/cancel"),
			PaymentMethodTypes: stripe.StringSlice([]string{
				"card",
			}),
			Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					Price:    stripe.String(req.Price),
					Quantity: stripe.Int64(1),
				},
			},
		}

		s, err := session.New(params)
		if err != nil {
			log.Printf("session.New: %v", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, &errResponse{err.Error()})
			return
		}

		render.JSON(w, r, struct {
			SessionID string `json:"sessionId"`
		}{SessionID: s.ID})

	}
}

func handleCheckoutSuccess(appCtx Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			http.Redirect(w, r, "/account", http.StatusSeeOther)
			return
		}
		s, err := session.Get(sessionID, nil)
		if err != nil {
			log.Printf("session.Get: %v\n", err)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, &errResponse{err.Error()})
			return
		}

		account, err := appCtx.authn.CurrentAccount(r)
		if err != nil {
			log.Printf("authn.CurrentAccount: %v\n", err)
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, &errResponse{"unauthorized"})
			return
		}

		err = account.Attributes().Set(billingIDKey, s.Customer.ID)
		if err != nil {
			log.Printf("UpdateBillingID %v\n", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, &errResponse{err.Error()})
			return
		}

		http.Redirect(w, r, "/account?checkout=success", http.StatusSeeOther)
	}
}

func handleCheckoutCancel(appCtx Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/account?checkout=cancel", http.StatusSeeOther)
	}
}

func handleManageSubscription(appCtx Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, err := appCtx.authn.CurrentAccount(r)
		if err != nil {
			log.Printf("authn.CurrentAccount: %v", err)
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, &errResponse{"unauthorized"})
			return
		}

		// expect plan to be change
		err = account.Attributes().Session().Del(w, currentPriceIDKey)
		if err != nil {
			log.Printf("account.Attributes().Session().Del(), %s failed  err %v\n", currentPriceIDKey, err)
		}
		billingID, ok := account.Attributes().Map().String(billingIDKey)
		if !ok {
			log.Printf(" %s not found \n", billingIDKey)
			http.Redirect(w, r, "/account", http.StatusSeeOther)
			return
		}
		params := &stripe.BillingPortalSessionParams{
			Customer:  stripe.String(billingID),
			ReturnURL: stripe.String(appCtx.cfg.Domain + "/account"),
		}

		ps, err := portalsession.New(params)
		if err != nil {
			log.Printf("portalsession.New: %v", err)
			http.Redirect(w, r, "/account", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, ps.URL, http.StatusSeeOther)

	}
}
