package authController

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
	"github.com/sojborg/go-todo/internal/cache"
)

// Custom errors
type AuthError struct {
	Message string
	Code    int
}

func (e *AuthError) Error() string {
	return e.Message
}

// Types
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

// Package variables
var (
	tokenCache = cache.NewTokenCache()
	logger     = log.Default() // Changed from structured.NewLogger() to use standard logger
)

// Helper functions
func getProviderFromRequest(r *http.Request) (string, error) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		return "", &AuthError{Message: "Provider is required", Code: http.StatusBadRequest}
	}
	return provider, nil
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Handler functions
func Login(w http.ResponseWriter, r *http.Request) {
	provider, err := getProviderFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
	logger.Printf("auth.login.attempt: provider=%s", provider) // Modified logging

	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		logger.Printf("auth.login.existing_session: user=%s", gothUser.Email) // Modified logging
		return
	}

	gothic.BeginAuthHandler(w, r)
}

func GetAuthCallbackFunction(w http.ResponseWriter, r *http.Request) {
	provider, err := getProviderFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		logger.Printf("auth.callback.error: %v", err) // Modified logging
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Printf("auth.callback.success: user=%s", user.Email) // Modified logging
	http.Redirect(w, r, fmt.Sprintf("http://localhost:5173?access_token=%s", user.AccessToken), http.StatusFound)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	tokenString := extractBearerToken(r)
	if tokenString == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	userInfo, err := verifyAccessToken(tokenString, "google")
	if err != nil {
		logger.Printf("auth.verify_token.error: %v", err) // Modified logging
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	sendJSONResponse(w, userInfo)
}

// Token verification
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	return strings.TrimPrefix(authHeader, "Bearer ")
}

func verifyAccessToken(accessToken, provider string) (*UserInfo, error) {
	if cachedInfo, exists := tokenCache.Get(accessToken); exists {
		return &UserInfo{ID: cachedInfo.UserID, Email: cachedInfo.Email}, nil
	}

	verifyURL := getVerifyURL(provider, accessToken)
	if verifyURL == "" {
		return nil, &AuthError{Message: "unsupported provider", Code: http.StatusBadRequest}
	}

	result, err := verifyTokenWithProvider(verifyURL)
	if err != nil {
		return nil, err
	}

	userInfo := extractUserInfo(result)
	if userInfo == nil {
		return nil, &AuthError{Message: "missing required user information", Code: http.StatusUnauthorized}
	}

	cacheTokenInfo(accessToken, result, userInfo)
	return userInfo, nil
}

func verifyTokenWithProvider(verifyURL string) (map[string]interface{}, error) {
	resp, err := http.Get(verifyURL)
	if err != nil {
		return nil, &AuthError{Message: "failed to verify token with provider", Code: http.StatusInternalServerError}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &AuthError{Message: "invalid token", Code: http.StatusUnauthorized}
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, &AuthError{Message: "failed to decode response from provider", Code: http.StatusInternalServerError}
	}

	return result, nil
}

func getVerifyURL(provider, accessToken string) string {
	switch provider {
	case "google":
		return "https://www.googleapis.com/oauth2/v1/tokeninfo?access_token=" + accessToken
	case "facebook":
		return "https://graph.facebook.com/debug_token?input_token=" + accessToken + "&access_token=" + accessToken
	default:
		return ""
	}
}

func cacheTokenInfo(accessToken string, result map[string]interface{}, userInfo *UserInfo) {
	expiresIn := time.Hour // Default 1 hour expiration

	// Try to get expiration from result
	if exp, ok := result["expires_in"].(float64); ok {
		expiresIn = time.Duration(exp) * time.Second
	}

	tokenCache.Set(accessToken, cache.TokenInfo{
		UserID:    userInfo.ID,
		Email:     userInfo.Email,
		ExpiresAt: time.Now().Add(expiresIn),
	})
}

func extractUserInfo(result map[string]interface{}) *UserInfo {
	// Handle Facebook-style response where data is nested
	if data, ok := result["data"].(map[string]interface{}); ok {
		result = data
	}

	// Extract user ID
	var userID string
	if id, ok := result["user_id"].(string); ok {
		userID = id
	} else if id, ok := result["sub"].(string); ok { // Some providers use "sub"
		userID = id
	} else {
		return nil
	}

	// Extract email
	var email string
	if e, ok := result["email"].(string); ok {
		email = e
	} else {
		return nil
	}

	// Extract name (optional)
	var name string
	if n, ok := result["name"].(string); ok {
		name = n
	}

	return &UserInfo{
		ID:    userID,
		Email: email,
		Name:  name,
	}
}

// ...existing helper functions for token verification...
