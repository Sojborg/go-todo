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

func Login(response http.ResponseWriter, request *http.Request) {
	provider := chi.URLParam(request, "provider")
	fmt.Println("### AUTH PROVIDER ###", provider)

	request = request.WithContext(context.WithValue(request.Context(), "provider", provider))

	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(response, request); err == nil {
		// Store the access token in a cookie
		cookie := &http.Cookie{
			Name:     "access_token",
			Value:    gothUser.AccessToken,
			Path:     "/",
			HttpOnly: false, // Allow access from client-side JavaScript
		}
		http.SetCookie(response, cookie)

		response.Header().Set("Location", "/")
		response.WriteHeader(http.StatusTemporaryRedirect)
		return
	} else {
		gothic.BeginAuthHandler(response, request)
	}
}
