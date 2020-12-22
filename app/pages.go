package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/sub"

	"github.com/google/uuid"

	"github.com/go-chi/chi"

	"github.com/mholt/binding"

	"github.com/foolin/goview"
)

func setDefaultPageData(appCtx Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			pageData := map[string]interface{}{}
			pageData["route"] = r.URL.Path
			pageData["app_name"] = strings.Title(strings.ToLower(appCtx.cfg.Name))
			defer func() {
				ctx := r.Context()
				ctx = context.WithValue(ctx, appCtxDataKey, pageData)
				next.ServeHTTP(w, r.WithContext(ctx))
			}()

			user, err := appCtx.users.LoggedInUser(r)
			if err != nil {
				return
			}

			pageData["is_logged_in"] = true
			pageData["email"] = user.Email
			pageData["metadata"] = user.Metadata
			if user.IsAPITokenSet {
				pageData["is_api_token_set"] = true
			}

			currentPriceID := appCtx.users.GetSessionStringVal(r, "current_price_id")
			// get currentPriceID using stripe customer ID
			if user.BillingID != "" && currentPriceID == nil {
				params := &stripe.SubscriptionListParams{
					Customer: user.BillingID,
					Status:   string(stripe.SubscriptionStatusActive),
				}
				params.AddExpand("data.items.data.price")
				params.Filters.AddFilter("limit", "", "1")

				i := sub.List(params)
				for i.Next() {
					s := i.Subscription()
					if s.Status == stripe.SubscriptionStatusActive {
						for _, pr := range s.Items.Data {
							currentPriceID = &pr.Price.ID
						}
					}

				}
			}

			if currentPriceID != nil {
				err = appCtx.users.SetSessionVal(r, w, "current_price_id", *currentPriceID)
				if err != nil {
					log.Println("SetSessionVal", err)
				}

				for _, plan := range appCtx.cfg.Plans {
					if plan.PriceID == *currentPriceID {
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
		})
	}
}

func signupPageSubmit(appCtx Context, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	var email, password string
	metadata := make(map[string]interface{})
	_ = r.ParseForm()
	for k, v := range r.Form {

		if k == "email" && len(v) == 0 {
			return goview.M{}, fmt.Errorf("email is required")
		}

		if k == "password" && len(v) == 0 {
			return goview.M{}, fmt.Errorf("password is required")
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

	err := appCtx.users.Signup(email, password, "owner", metadata)
	if err != nil {
		return goview.M{}, err
	}

	http.Redirect(w, r, "/login?confirmation_sent=true", http.StatusSeeOther)

	return goview.M{}, nil
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

func loginPageSubmit(appCtx Context, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	loginForm := new(LoginForm)
	if errs := binding.Bind(r, loginForm); errs != nil {
		return nil, fmt.Errorf("%v, %w", errs, fmt.Errorf("missing email"))
	}

	if loginForm.Magic == "magic" {
		err := appCtx.users.OTP(loginForm.Email)
		if err != nil {
			return nil, err
		}
		http.Redirect(w, r, "/magic-link-sent", http.StatusSeeOther)
	} else {
		err := appCtx.users.Login(w, r, loginForm.Email, loginForm.Password)
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

	return goview.M{}, nil
}

func magicLinkLoginConfirm(appCtx Context, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	otp := chi.URLParam(r, "otp")
	err := appCtx.users.LoginWithOTP(w, r, otp)
	if err != nil {
		return nil, err
	}

	redirectTo := "/app"

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)

	return goview.M{}, nil
}

func loginPage(appCtx Context, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	confirmed := r.URL.Query().Get("confirmed")
	if confirmed == "true" {
		return goview.M{
			"confirmed": true,
		}, nil
	}

	notConfirmed := r.URL.Query().Get("not_confirmed")
	if notConfirmed == "true" {
		return goview.M{
			"not_confirmed": true,
		}, nil
	}

	confirmationSent := r.URL.Query().Get("confirmation_sent")
	if confirmationSent == "true" {
		return goview.M{
			"confirmation_sent": true,
		}, nil
	}

	emailChanged := r.URL.Query().Get("email_changed")
	if emailChanged == "true" {
		return goview.M{
			"email_changed": true,
		}, nil
	}

	from := r.URL.Query().Get("from")
	if from != "" {
		// store from in session to be used by external login(goth)
		appCtx.users.SetSessionVal(r, w, "from", from)
	}

	return goview.M{}, nil
}

func gothAuthCallbackPage(appCtx Context, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	err := appCtx.users.HandleGothCallback(w, r, "owner", nil)
	if err != nil {
		return goview.M{}, err
	}
	redirectTo := "/app"

	fromVal, err := appCtx.users.GetSessionVal(r, "from")
	if err == nil && fromVal != nil {
		redirectTo = fromVal.(string)
		appCtx.users.DelSessionVal(r, w, "from")
	}

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	return goview.M{}, nil
}

func gothAuthPage(appCtx Context, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	err := appCtx.users.HandleGothLogin(w, r)
	if err != nil {
		return goview.M{}, err
	}
	redirectTo := "/app"
	from := r.URL.Query().Get("from")
	if from != "" {
		redirectTo = from
	}

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	return goview.M{}, nil
}

func confirmEmailChangePage(appCtx Context, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	token := chi.URLParam(r, "token")
	err := appCtx.users.ConfirmEmailChange(token)
	if err != nil {
		return nil, err
	}
	http.Redirect(w, r, "/account?email_changed=true", http.StatusSeeOther)

	return goview.M{}, nil
}

func confirmEmailPage(appCtx Context, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	token := chi.URLParam(r, "token")
	err := appCtx.users.ConfirmEmail(token)
	if err != nil {
		return nil, err
	}

	http.Redirect(w, r, "/login?confirmed=true", http.StatusSeeOther)
	return goview.M{}, nil
}

func appPage(_ Context, _ http.ResponseWriter, _ *http.Request) (goview.M, error) {
	dummy := struct {
		Title string `json:"title"`
	}{
		Title: "Hello Props",
	}

	d, err := json.Marshal(&dummy)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, fmt.Errorf("encoding failed"))
	}

	return goview.M{
		"Data": string(d),
	}, nil
}

func accountPage(appCtx Context, _ http.ResponseWriter, r *http.Request) (goview.M, error) {
	emailChanged := r.URL.Query().Get("email_changed")
	if emailChanged == "true" {
		return goview.M{
			"form_token":    uuid.New(),
			"email_changed": true,
		}, nil
	}

	checkout := r.URL.Query().Get("checkout")
	if checkout == "success" || checkout == "cancel" {
		return goview.M{
			"checkout": checkout,
			"plans":    appCtx.cfg.Plans,
		}, nil
	}

	return goview.M{
		"form_token": uuid.New(),
		"plans":      appCtx.cfg.Plans,
	}, nil
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

func accountPageSubmit(appCtx Context, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	accountForm := new(AccountForm)
	binding.Bind(r, accountForm)

	pageData := goview.M{}

	user, err := appCtx.users.LoggedInUser(r)
	if err != nil {
		return nil, err
	}

	if accountForm.ResetAPIToken {
		// check if the form has been previously submitted
		if accountForm.FormToken != "" {
			formTokenVal, err := appCtx.users.GetSessionVal(r, "form_token")
			if err == nil && formTokenVal != nil {
				formToken := formTokenVal.(string)
				if formToken == accountForm.FormToken {
					return goview.M{}, nil
				}
			}
		}
		apiToken, err := appCtx.users.ResetAPIToken(r)
		if err != nil {
			return goview.M{}, err
		}

		appCtx.users.SetSessionVal(r, w, "form_token", accountForm.FormToken)
		return goview.M{
			"is_api_token_set": true,
			"api_token":        apiToken,
		}, nil
	}

	if accountForm.Email != "" && accountForm.Email != user.Email {
		err = appCtx.users.ChangeEmail(user.ID, accountForm.Email)
		if err != nil {
			return nil, err
		}
		pageData["change_email"] = "requested"
	}

	var name string
	var ok bool
	if user.Metadata["name"] == nil {
		name = ""
	} else {
		name, ok = user.Metadata["name"].(string)
		if !ok {
			return nil, err
		}
	}

	if name != accountForm.Name {
		err = appCtx.users.UpdateMetaData(r, map[string]interface{}{
			"name": accountForm.Name,
		})

		if err != nil {
			return nil, err
		}
	}

	appCtx.users.SetSessionVal(r, w, "form_token", accountForm.FormToken)

	pageData["email"] = user.Email

	user.Metadata["name"] = accountForm.Name
	pageData["metadata"] = user.Metadata
	return pageData, nil
}

func forgotPageSubmit(appCtx Context, _ http.ResponseWriter, r *http.Request) (goview.M, error) {
	accountForm := new(AccountForm)
	if errs := binding.Bind(r, accountForm); errs != nil {
		return nil, fmt.Errorf("%v, %w", errs, "email or password missing")
	}

	pageData := goview.M{}

	err := appCtx.users.Recovery(accountForm.Email)
	if err != nil {
		return pageData, err
	}

	pageData["recovery_sent"] = true

	return pageData, nil
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

func resetPageSubmit(appCtx Context, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	token := chi.URLParam(r, "token")
	resetForm := new(ResetForm)
	if errs := binding.Bind(r, resetForm); errs != nil {
		return nil, fmt.Errorf("%v, %w", errs, fmt.Errorf("missing password"))
	}

	err := appCtx.users.ConfirmRecovery(token, resetForm.Password)
	if err != nil {
		return goview.M{}, err
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)

	return goview.M{}, nil
}

func deleteAccount(appCtx Context, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	err := appCtx.users.DeleteUser(r)
	if err != nil {
		return goview.M{}, err
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return goview.M{}, nil
}
