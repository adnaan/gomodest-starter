package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

func loginPage(appCtx AppContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
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

func accountPage(appCtx AppContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	emailChanged := r.URL.Query().Get("email_changed")
	if emailChanged == "true" {
		return goview.M{
			"email_changed": true,
		}, nil
	}

	return goview.M{}, nil
}

func appPage(appCtx AppContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	dummy := struct {
		Title string `json:"title"`
	}{
		Title: "Hello Props",
	}

	d, err := json.Marshal(&dummy)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, HTTPErr{
			UserMessage: "Couldn't find data",
			Code:        404,
		})
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
func (af *AccountForm) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&af.Name:  "name",
		&af.Email: "email",
	}
}

func accountPageSubmit(appCtx AppContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	accountForm := new(AccountForm)
	if errs := binding.Bind(r, accountForm); errs != nil {
		return nil, fmt.Errorf("%v, %w", errs, BadRequest)
	}

	pageData := goview.M{}

	id, email, metaData, err := appCtx.users.LoggedInUser(r)
	if err != nil {
		return nil, fmt.Errorf("%v, %w", err, InternalErr)
	}

	if accountForm.Email != "" && accountForm.Email != email {
		err = appCtx.users.ChangeEmail(id, accountForm.Email)
		if err != nil {
			return nil, fmt.Errorf("%v, %w", err, InternalErr)
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
			return nil, fmt.Errorf("%v, %w", fmt.Errorf("invalid name"), InternalErr)
		}
	}

	if name != accountForm.Name {
		err = appCtx.users.UpdateMetaData(id, map[string]interface{}{
			"name": accountForm.Name,
		})

		if err != nil {
			return nil, fmt.Errorf("%v, %w", err, InternalErr)
		}
	}

	pageData["email"] = email

	metaData["name"] = accountForm.Name
	pageData["metadata"] = metaData
	fmt.Println("accountPageSubmit", pageData)
	return pageData, nil
}
