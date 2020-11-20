package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/mholt/binding"

	"github.com/foolin/goview"
)

type LoginForm struct {
	Email    string
	Password string
}

// Fieldmap for the LoginForm. extend it for more fields
func (lf *LoginForm) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&lf.Email:    "email",
		&lf.Password: "password",
	}
}

func loginPageSubmit(w http.ResponseWriter, r *http.Request) (goview.M, error) {
	loginForm := new(LoginForm)
	if errs := binding.Bind(r, loginForm); errs != nil {
		return nil, fmt.Errorf("%v, %w", errs, BadRequest)
	}

	session, err := store.Get(r, "auth-session")
	if err != nil {
		return nil, fmt.Errorf("%v, %w", err, InternalErr)
	}
	session.Values["token"] = uuid.New().String()
	profile := make(map[string]interface{})
	profile["email"] = loginForm.Email
	profile["name"] = ""

	session.Values["profile"] = profile
	err = session.Save(r, w)
	if err != nil {
		return nil, fmt.Errorf("%v, %w", err, InternalErr)
	}

	http.Redirect(w, r, "/app", http.StatusSeeOther)
	return goview.M{}, nil
}

func appPage(w http.ResponseWriter, r *http.Request) (goview.M, error) {
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

func accountPage(w http.ResponseWriter, r *http.Request) (goview.M, error) {

	session, err := store.Get(r, "auth-session")
	if err != nil {
		return nil, fmt.Errorf("%v, %w", err, InternalErr)
	}

	profileData, ok := session.Values["profile"]
	if !ok {
		return nil, fmt.Errorf("%v, %w", err, InternalErr)
	}

	profile, ok := profileData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%v, %w", err, InternalErr)
	}

	return goview.M{
		"name": profile["name"],
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

func accountPageSubmit(w http.ResponseWriter, r *http.Request) (goview.M, error) {
	accountForm := new(AccountForm)
	if errs := binding.Bind(r, accountForm); errs != nil {
		return nil, fmt.Errorf("%v, %w", errs, BadRequest)
	}

	session, err := store.Get(r, "auth-session")
	if err != nil {
		return nil, fmt.Errorf("%v, %w", err, InternalErr)
	}

	profileData, ok := session.Values["profile"]
	if !ok {
		return nil, fmt.Errorf("%v, %w", err, InternalErr)
	}

	profile, ok := profileData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%v, %w", err, InternalErr)
	}

	profile["name"] = accountForm.Name
	session.Values["profile"] = profile
	err = session.Save(r, w)
	if err != nil {
		return nil, fmt.Errorf("%v, %w", err, InternalErr)
	}

	return goview.M{
		"name": accountForm.Name,
	}, nil
}
