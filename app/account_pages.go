package app

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/lithammer/shortuuid/v3"

	rl "github.com/adnaan/renderlayout"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/sub"

	"github.com/google/uuid"

	"github.com/go-chi/chi"

	"github.com/mholt/binding"
)

func defaultPageHandler(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		pageData := map[string]interface{}{}
		pageData["route"] = r.URL.Path
		pageData["app_name"] = strings.Title(strings.ToLower(appCtx.cfg.Name))
		pageData["feature_groups"] = appCtx.cfg.FeatureGroups

		account, err := appCtx.authn.CurrentAccount(r)
		if err != nil {
			return pageData, nil
		}

		accAttributes := account.Attributes().Map()
		if _, ok := accAttributes["api_key"]; ok {
			pageData["is_api_token_set"] = true
		}

		pageData["is_logged_in"] = true
		pageData["email"] = account.Email()
		pageData["metadata"] = accAttributes

		currentPriceID, _ := account.Attributes().Session().Get(currentPriceIDKey)
		// get currentPriceID using stripe customer ID
		billingId, billingIDExists := accAttributes.String(billingIDKey)
		if billingIDExists && currentPriceID == nil {
			params := &stripe.SubscriptionListParams{
				Customer: billingId,
				Status:   string(stripe.SubscriptionStatusActive),
			}
			params.AddExpand("data.items.data.price")
			params.Filters.AddFilter("limit", "", "1")

			i := sub.List(params)
			for i.Next() {
				s := i.Subscription()
				if s.Status == stripe.SubscriptionStatusActive {
					for _, pr := range s.Items.Data {
						currentPriceID = pr.Price.ID
					}
				}

			}
		}

		if currentPriceID != nil {
			err = account.Attributes().Session().Set(w, currentPriceIDKey, currentPriceID)
			if err != nil {
				log.Println("SetSessionVal", err)
			}

			for _, plan := range appCtx.cfg.Plans {
				if plan.PriceID == currentPriceID.(string) {
					pageData["current_plan"] = Plan{
						Current:   true,
						PriceID:   plan.PriceID,
						Name:      plan.Name,
						Price:     plan.Price,
						Details:   plan.Details,
						StripeKey: plan.StripeKey,
					}
				}
			}

		}

		return pageData, nil
	}
}

func signupPageSubmit(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		var email, password string
		metadata := make(map[string]interface{})
		_ = r.ParseForm()
		for k, v := range r.Form {

			if k == "email" && len(v) == 0 {
				return rl.D{}, fmt.Errorf("email is required")
			}

			if k == "password" && len(v) == 0 {
				return rl.D{}, fmt.Errorf("password is required")
			}

			if len(v) == 0 {
				continue
			}

			if k == "email" && len(v) > 0 {
				email = v[0]
				continue
			}

			if k == "password" && len(v) > 0 {
				password = v[0]
				continue
			}

			if len(v) == 1 {
				metadata[k] = v[0]
				continue
			}
			if len(v) > 1 {
				metadata[k] = v
			}
		}

		err := appCtx.authn.Signup(r.Context(), email, password, metadata)
		if err != nil {
			return rl.D{}, err
		}

		http.Redirect(w, r, "/login?confirmation_sent=true", http.StatusSeeOther)

		return rl.D{}, nil
	}
}

type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Magic    string `json:"magic"`
}

func (l *LoginForm) Bind(_ *http.Request) error {
	return nil
}

// Fieldmap for the LoginData
func (l *LoginForm) FieldMap(_ *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&l.Email: binding.Field{
			Form:     "email",
			Required: true,
		},
		&l.Password: binding.Field{
			Form:     "password",
			Required: false,
		},
		&l.Magic: binding.Field{
			Form:     "magic",
			Required: false,
		},
	}
}

func loginPageSubmit(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		loginForm := new(LoginForm)
		if errs := binding.Bind(r, loginForm); errs != nil {
			return nil, fmt.Errorf("%v, %w",
				errs, fmt.Errorf("missing email"))
		}

		if loginForm.Magic == "magic" {
			err := appCtx.authn.SendPasswordlessToken(r.Context(), loginForm.Email)
			if err != nil {
				return nil, err
			}
			http.Redirect(w, r, "/magic-link-sent", http.StatusSeeOther)
		} else {
			err := appCtx.authn.Login(w, r, loginForm.Email, loginForm.Password)
			if err != nil {
				return nil, err
			}

			redirectTo := "/app"
			from := r.URL.Query().Get("from")
			if from != "" {
				redirectTo = from
			}

			http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		}

		return rl.D{}, nil
	}
}

func magicLinkLoginConfirm(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		otp := chi.URLParam(r, "otp")
		err := appCtx.authn.LoginWithPasswordlessToken(w, r, otp)
		if err != nil {
			return nil, err
		}

		redirectTo := "/app"

		http.Redirect(w, r, redirectTo, http.StatusSeeOther)

		return rl.D{}, nil
	}
}

func loginPage(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {

		confirmed := r.URL.Query().Get("confirmed")
		if confirmed == "true" {
			return rl.D{
				"confirmed": true,
			}, nil
		}

		notConfirmed := r.URL.Query().Get("not_confirmed")
		if notConfirmed == "true" {
			return rl.D{
				"not_confirmed": true,
			}, nil
		}

		confirmationSent := r.URL.Query().Get("confirmation_sent")
		if confirmationSent == "true" {
			return rl.D{
				"confirmation_sent": true,
			}, nil
		}

		emailChanged := r.URL.Query().Get("email_changed")
		if emailChanged == "true" {
			return rl.D{
				"email_changed": true,
			}, nil
		}

		return rl.D{}, nil
	}
}

func gothAuthCallbackPage(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		err := appCtx.authn.LoginProviderCallback(w, r, nil)
		if err != nil {
			return rl.D{}, err
		}
		redirectTo := "/app"

		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return rl.D{}, nil
	}
}

func gothAuthPage(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		err := appCtx.authn.LoginWithProvider(w, r)
		if err != nil {
			return rl.D{}, err
		}
		redirectTo := "/app"
		from := r.URL.Query().Get("from")
		if from != "" {
			redirectTo = from
		}

		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return rl.D{}, nil
	}
}

func confirmEmailChangePage(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		token := chi.URLParam(r, "token")
		acc, err := appCtx.authn.CurrentAccount(r)
		if err != nil {
			return nil, err
		}
		err = acc.ConfirmEmailChange(token)
		if err != nil {
			return nil, err
		}
		http.Redirect(w, r, "/account?email_changed=true", http.StatusSeeOther)
		return rl.D{}, nil
	}
}

func confirmEmailPage(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		token := chi.URLParam(r, "token")
		err := appCtx.authn.ConfirmSignupEmail(r.Context(), token)
		if err != nil {
			return nil, err
		}

		http.Redirect(w, r, "/login?confirmed=true", http.StatusSeeOther)
		return rl.D{}, nil
	}
}
func appPage(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		return rl.D{}, nil
	}
}

func accountPage(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		emailChanged := r.URL.Query().Get("email_changed")
		if emailChanged == "true" {
			return rl.D{
				"form_token":    uuid.New(),
				"email_changed": true,
			}, nil
		}

		checkout := r.URL.Query().Get("checkout")
		if checkout == "success" || checkout == "cancel" {
			return rl.D{
				"checkout": checkout,
				"plans":    appCtx.cfg.Plans,
			}, nil
		}

		return rl.D{
			"form_token": uuid.New(),
			"plans":      appCtx.cfg.Plans,
		}, nil
	}
}

type AccountForm struct {
	Name          string
	Email         string
	ResetAPIToken bool
	FormToken     string
}

// Fieldmap for the accountform. extend it for more fields
func (af *AccountForm) FieldMap(_ *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&af.Name:          "name",
		&af.Email:         "email",
		&af.ResetAPIToken: "reset_api_token",
		&af.FormToken:     "form_token",
	}
}

func accountPageSubmit(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		accountForm := new(AccountForm)
		binding.Bind(r, accountForm)

		pageData := make(map[string]interface{})

		account, err := appCtx.authn.CurrentAccount(r)
		if err != nil {
			return nil, err
		}

		if accountForm.ResetAPIToken {
			// check if the form has been previously submitted
			if accountForm.FormToken != "" {
				formTokenVal, err := account.Attributes().Session().Get("form_token")
				if err == nil && formTokenVal != nil {
					formToken := formTokenVal.(string)
					if formToken == accountForm.FormToken {
						return rl.D{}, nil
					}
				}
			}

			apiKey := shortuuid.New()
			token, err := appCtx.branca.EncodeToString(apiKey)
			if err != nil {
				return nil, fmt.Errorf("%v %w", err, ErrInternal)
			}

			err = account.Attributes().Set("api_key", apiKey)
			if err != nil {
				return nil, fmt.Errorf("%v %w", err, ErrInternal)
			}

			account.Attributes().Session().Set(w, "form_token", accountForm.FormToken)
			return rl.D{
				"is_api_token_set": true,
				"api_token":        token,
			}, nil
		}

		if accountForm.Email != "" && accountForm.Email != account.Email() {
			err = account.ChangeEmail(accountForm.Email)
			if err != nil {
				return nil, err
			}
			pageData["change_email"] = "requested"
		}

		name, _ := account.Attributes().Map().String("name")
		if name != accountForm.Name {
			err = account.Attributes().Set("name", accountForm.Name)
			if err != nil {
				return nil, err
			}
		}

		account.Attributes().Session().Set(w, "form_token", accountForm.FormToken)

		pageData["email"] = account.Email()
		pageData["metadata"] = account.Attributes().Map()
		return pageData, nil
	}
}

func forgotPageSubmit(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		accountForm := new(AccountForm)
		if errs := binding.Bind(r, accountForm); errs != nil {
			return nil, fmt.Errorf("%v, %w", errs, "email or password missing")
		}

		pageData := make(map[string]interface{})

		err := appCtx.authn.Recovery(r.Context(), accountForm.Email)
		if err != nil {
			return pageData, err
		}

		pageData["recovery_sent"] = true

		return pageData, nil
	}
}

type ResetForm struct {
	Password string
}

// Fieldmap for the ResetForm. extend it for more fields
func (rf *ResetForm) FieldMap(_ *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&rf.Password: "password",
	}
}

func resetPageSubmit(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		token := chi.URLParam(r, "token")
		resetForm := new(ResetForm)
		if errs := binding.Bind(r, resetForm); errs != nil {
			return nil, fmt.Errorf("%v, %w", errs, fmt.Errorf("missing password"))
		}

		err := appCtx.authn.ConfirmRecovery(r.Context(), token, resetForm.Password)
		if err != nil {
			return rl.D{}, err
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return rl.D{}, nil
	}
}

func deleteAccount(appCtx Context) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		account, err := appCtx.authn.CurrentAccount(r)
		if err != nil {
			return nil, err
		}
		err = account.Delete()
		if err != nil {
			return nil, err
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return rl.D{}, nil
	}
}
