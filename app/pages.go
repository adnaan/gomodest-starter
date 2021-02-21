package app

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/adnaan/gomodest/app/internal/models/task"
	"github.com/lithammer/shortuuid/v3"

	"github.com/adnaan/users"

	rl "github.com/adnaan/renderlayout"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/sub"

	"github.com/google/uuid"

	"github.com/go-chi/chi"

	"github.com/mholt/binding"
)

func defaultPageHandler(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		pageData := map[string]interface{}{}
		pageData["route"] = r.URL.Path
		pageData["app_name"] = strings.Title(strings.ToLower(appCtx.cfg.Name))
		pageData["feature_groups"] = appCtx.cfg.FeatureGroups

		user, err := appCtx.users.LoggedInUser(r)
		if err != nil {
			return pageData, nil
		}

		pageData["is_logged_in"] = true
		pageData["email"] = user.Email
		pageData["metadata"] = user.Metadata
		pageData["workspaces"] = user.Workspaces
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

		return pageData, nil
	}
}

func signupPageSubmit(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		var email, password string
		metadata := make(map[string]interface{})
		_ = r.ParseForm()
		for k, v := range r.Form {

			if k == "email" && len(v) == 0 {
				return rl.M{}, fmt.Errorf("email is required")
			}

			if k == "password" && len(v) == 0 {
				return rl.M{}, fmt.Errorf("password is required")
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
			return rl.M{}, err
		}

		http.Redirect(w, r, "/login?confirmation_sent=true", http.StatusSeeOther)

		return rl.M{}, nil
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

func loginPageSubmit(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		loginForm := new(LoginForm)
		if errs := binding.Bind(r, loginForm); errs != nil {
			return nil, fmt.Errorf("%v, %w",
				errs, fmt.Errorf("missing email"))
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

		return rl.M{}, nil
	}
}

func magicLinkLoginConfirm(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		otp := chi.URLParam(r, "otp")
		err := appCtx.users.LoginWithOTP(w, r, otp)
		if err != nil {
			return nil, err
		}

		redirectTo := "/app"

		http.Redirect(w, r, redirectTo, http.StatusSeeOther)

		return rl.M{}, nil
	}
}

func loginPage(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {

		confirmed := r.URL.Query().Get("confirmed")
		if confirmed == "true" {
			return rl.M{
				"confirmed": true,
			}, nil
		}

		notConfirmed := r.URL.Query().Get("not_confirmed")
		if notConfirmed == "true" {
			return rl.M{
				"not_confirmed": true,
			}, nil
		}

		confirmationSent := r.URL.Query().Get("confirmation_sent")
		if confirmationSent == "true" {
			return rl.M{
				"confirmation_sent": true,
			}, nil
		}

		emailChanged := r.URL.Query().Get("email_changed")
		if emailChanged == "true" {
			return rl.M{
				"email_changed": true,
			}, nil
		}

		from := r.URL.Query().Get("from")
		if from != "" {
			// store from in session to be used by external login(goth)
			appCtx.users.SetSessionVal(r, w, "from", from)
		}

		return rl.M{}, nil
	}
}

func gothAuthCallbackPage(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		err := appCtx.users.HandleGothCallback(w, r, "owner", nil)
		if err != nil {
			return rl.M{}, err
		}
		redirectTo := "/app"

		fromVal, err := appCtx.users.GetSessionVal(r, "from")
		if err == nil && fromVal != nil {
			redirectTo = fromVal.(string)
			appCtx.users.DelSessionVal(r, w, "from")
		}

		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return rl.M{}, nil
	}
}

func gothAuthPage(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		err := appCtx.users.HandleGothLogin(w, r)
		if err != nil {
			return rl.M{}, err
		}
		redirectTo := "/app"
		from := r.URL.Query().Get("from")
		if from != "" {
			redirectTo = from
		}

		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return rl.M{}, nil
	}
}

func confirmEmailChangePage(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		token := chi.URLParam(r, "token")
		err := appCtx.users.ConfirmEmailChange(token)
		if err != nil {
			return nil, err
		}
		http.Redirect(w, r, "/account?email_changed=true", http.StatusSeeOther)

		return rl.M{}, nil
	}
}

func confirmEmailPage(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		token := chi.URLParam(r, "token")
		err := appCtx.users.ConfirmEmail(token)
		if err != nil {
			return nil, err
		}

		http.Redirect(w, r, "/login?confirmed=true", http.StatusSeeOther)
		return rl.M{}, nil
	}
}
func appPage(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		return rl.M{}, nil
	}
}

func accountPage(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		emailChanged := r.URL.Query().Get("email_changed")
		if emailChanged == "true" {
			return rl.M{
				"form_token":    uuid.New(),
				"email_changed": true,
			}, nil
		}

		checkout := r.URL.Query().Get("checkout")
		if checkout == "success" || checkout == "cancel" {
			return rl.M{
				"checkout": checkout,
				"plans":    appCtx.cfg.Plans,
			}, nil
		}

		return rl.M{
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

func accountPageSubmit(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		accountForm := new(AccountForm)
		binding.Bind(r, accountForm)

		pageData := make(map[string]interface{})

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
						return rl.M{}, nil
					}
				}
			}
			apiToken, err := appCtx.users.ResetAPIToken(r)
			if err != nil {
				return rl.M{}, err
			}

			appCtx.users.SetSessionVal(r, w, "form_token", accountForm.FormToken)
			return rl.M{
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
}

func forgotPageSubmit(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		accountForm := new(AccountForm)
		if errs := binding.Bind(r, accountForm); errs != nil {
			return nil, fmt.Errorf("%v, %w", errs, "email or password missing")
		}

		pageData := make(map[string]interface{})

		err := appCtx.users.Recovery(accountForm.Email)
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

func resetPageSubmit(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		token := chi.URLParam(r, "token")
		resetForm := new(ResetForm)
		if errs := binding.Bind(r, resetForm); errs != nil {
			return nil, fmt.Errorf("%v, %w", errs, fmt.Errorf("missing password"))
		}

		err := appCtx.users.ConfirmRecovery(token, resetForm.Password)
		if err != nil {
			return rl.M{}, err
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return rl.M{}, nil
	}
}

func deleteAccount(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		err := appCtx.users.DeleteUser(r)
		if err != nil {
			return rl.M{}, err
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return rl.M{}, nil
	}
}

func listTasks(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {

		userID := r.Context().Value(users.CtxUserIdKey).(string)
		tasks, err := appCtx.db.Task.Query().Where(task.Owner(userID)).All(appCtx.ctx)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return rl.M{
			"tasks": tasks,
		}, nil
	}
}

func createNewTask(appCtx Context) rl.ViewHandlerFunc {
	type req struct {
		Text string `json:"text"`
	}

	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		userID := r.Context().Value(users.CtxUserIdKey).(string)

		req := new(req)
		err := r.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		err = appCtx.formDecoder.Decode(req, r.Form)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		if req.Text == "" {
			return nil, fmt.Errorf("%w", fmt.Errorf("empty task"))
		}

		_, err = appCtx.db.Task.Create().
			SetID(shortuuid.New()).
			SetStatus(task.StatusInprogress).
			SetOwner(userID).
			SetText(req.Text).
			Save(appCtx.ctx)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return nil, nil
	}
}

func deleteTask(appCtx Context) rl.ViewHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		id := chi.URLParam(r, "id")

		userID := r.Context().Value(users.CtxUserIdKey).(string)

		_, err := appCtx.db.Task.Delete().Where(task.And(
			task.Owner(userID), task.ID(id),
		)).Exec(appCtx.ctx)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return nil, nil
	}
}

func editTask(appCtx Context) rl.ViewHandlerFunc {
	type req struct {
		Text string `json:"text"`
	}
	return func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
		id := chi.URLParam(r, "id")

		userID := r.Context().Value(users.CtxUserIdKey).(string)

		req := new(req)
		err := r.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		err = appCtx.formDecoder.Decode(req, r.Form)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		if req.Text == "" {
			return nil, fmt.Errorf("%w", fmt.Errorf("empty task"))
		}

		err = appCtx.db.Task.Update().Where(task.And(
			task.Owner(userID), task.ID(id),
		)).SetText(req.Text).Exec(appCtx.ctx)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return nil, nil
	}
}
