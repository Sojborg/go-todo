package authController

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
)

func GetAuthCallbackFunction(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	fmt.Println("### CALLBACK ###", provider)

	if provider == "" {
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}

	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Println("### ERROR ###", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("### USER ###", user)
}

func Logout(res http.ResponseWriter, req *http.Request) {
	gothic.Logout(res, req)
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func Login(res http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	fmt.Println("### AUTH PROVIDER ###", provider)

	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(res, r); err == nil {
		fmt.Fprintln(res, gothUser)
	} else {
		gothic.BeginAuthHandler(res, r)
	}
}
