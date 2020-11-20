package pkg

import (
	"context"
	"fmt"
	"net/http"
)

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "auth-session")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	session.Values = nil
	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func setDefaultPageData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		pageData := map[string]interface{}{
			"route": r.URL.Path,
		}

		session, err := store.Get(r, "auth-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if profileData, ok := session.Values["profile"]; ok {
			pageData["is_logged_in"] = "true"
			profile, ok := profileData.(map[string]interface{})
			if ok {
				fmt.Printf("%+v\n", profile)
				pageData["email"] = profile["email"]
			}
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, appCtxDataKey, pageData)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func isAuthenticatedAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "auth-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, ok := session.Values["profile"]; !ok {
			http.Error(w, "Unauthorised", http.StatusUnauthorized)
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func isAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "auth-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, ok := session.Values["profile"]; !ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
