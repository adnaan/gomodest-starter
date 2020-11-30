package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/mholt/binding"

	"github.com/foolin/goview"
)

func setDefaultPageData(appCtx AppContext) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pageData := map[string]interface{}{
				"route": r.URL.Path,
			}

			_, email, metadata, err := appCtx.users.LoggedInUser(r)
			if err == nil {
				pageData["email"] = email
				pageData["metadata"] = metadata
				pageData["is_logged_in"] = true
				fmt.Println("setDefaultPageData", pageData)
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, appCtxDataKey, pageData)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func signupPageSubmit(appCtx AppContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
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

	err := appCtx.users.Signup(email, password, metadata)
	if err != nil {
		return goview.M{}, nil
	}

	http.Redirect(w, r, "/login?confirmation_sent=true", http.StatusSeeOther)

	return goview.M{}, nil
}

type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
			Required: true,
		},
	}
}

func loginPageSubmit(appCtx AppContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	loginForm := new(LoginForm)
	if errs := binding.Bind(r, loginForm); errs != nil {
		return nil, fmt.Errorf("%v, %w", errs, fmt.Errorf("missing email or password"))
	}

	err := appCtx.users.Login(w, r, loginForm.Email, loginForm.Password)
	if err != nil {
		return nil, err
	}

	http.Redirect(w, r, "/app", http.StatusSeeOther)
	return goview.M{}, nil
}

func loginPage(_ AppContext, _ http.ResponseWriter, r *http.Request) (goview.M, error) {
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

	return goview.M{}, nil
}

func accountPage(_ AppContext, _ http.ResponseWriter, r *http.Request) (goview.M, error) {
	emailChanged := r.URL.Query().Get("email_changed")
	if emailChanged == "true" {
		return goview.M{
			"email_changed": true,
		}, nil
	}

	return goview.M{}, nil
}

func confirmEmailChangePage(appCtx AppContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	token := chi.URLParam(r, "token")
	err := appCtx.users.ConfirmEmailChange(token)
	if err != nil {
		return nil, err
	}
	http.Redirect(w, r, "/account?email_changed=true", http.StatusSeeOther)

	return goview.M{}, nil
}

func confirmEmailPage(appCtx AppContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	token := chi.URLParam(r, "token")
	err := appCtx.users.ConfirmEmail(token)
	if err != nil {
		return nil, err
	}
	http.Redirect(w, r, "/login?confirmed=true", http.StatusSeeOther)
	return goview.M{}, nil
}

func appPage(_ AppContext, _ http.ResponseWriter, _ *http.Request) (goview.M, error) {
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

type AccountForm struct {
	Name  string
	Email string
}

// Fieldmap for the accountform. extend it for more fields
func (af *AccountForm) FieldMap(_ *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&af.Name:  "name",
		&af.Email: "email",
	}
}

func accountPageSubmit(appCtx AppContext, _ http.ResponseWriter, r *http.Request) (goview.M, error) {
	accountForm := new(AccountForm)
	if errs := binding.Bind(r, accountForm); errs != nil {
		return nil, fmt.Errorf("%v, %w", errs, fmt.Errorf("missing name or email"))
	}

	pageData := goview.M{}

	id, email, metaData, err := appCtx.users.LoggedInUser(r)
	if err != nil {
		return nil, err
	}

	if accountForm.Email != "" && accountForm.Email != email {
		err = appCtx.users.ChangeEmail(id, accountForm.Email)
		if err != nil {
			return nil, err
		}
		pageData["change_email"] = "requested"
	}

	var name string
	var ok bool
	if metaData["name"] == nil {
		name = ""
	} else {
		name, ok = metaData["name"].(string)
		if !ok {
			return nil, err
		}
	}

	if name != accountForm.Name {
		err = appCtx.users.UpdateMetaData(id, map[string]interface{}{
			"name": accountForm.Name,
		})

		if err != nil {
			return nil, err
		}
	}

	pageData["email"] = email

	metaData["name"] = accountForm.Name
	pageData["metadata"] = metaData
	return pageData, nil
}

func forgotPageSubmit(appCtx AppContext, _ http.ResponseWriter, r *http.Request) (goview.M, error) {
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

func resetPageSubmit(appCtx AppContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	token := chi.URLParam(r, "token")
	resetForm := new(ResetForm)
	if errs := binding.Bind(r, resetForm); errs != nil {
		return nil, fmt.Errorf("%v, %w", errs, fmt.Errorf("missing password"))
	}

	pageData := goview.M{}

	err := appCtx.users.ConfirmRecovery(token, resetForm.Password)
	if err != nil {
		return pageData, err
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)

	return pageData, nil
}
