package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/crypto/bcrypt"

	"humanguard/auth"
	"humanguard/storage"

	"github.com/google/uuid"
)

type UserHandler struct {
	storage        storage.Storage
	jwt            *auth.JWTService
	totp           *auth.TOTPService
	oauth          *auth.OAuthService
	googleOAuth    *auth.OAuthService
	githubOAuth    *auth.OAuthService
	sessionManager *auth.UserSessionManager
}

func NewUserHandler(
	store storage.Storage,
	jwtService *auth.JWTService,
	totpService *auth.TOTPService,
	oauthService *auth.OAuthService,
	googleOAuth *auth.OAuthService,
	githubOAuth *auth.OAuthService,
	sm *auth.UserSessionManager,
) *UserHandler {
	return &UserHandler{
		storage:        store,
		jwt:            jwtService,
		totp:           totpService,
		oauth:          oauthService,
		googleOAuth:    googleOAuth,
		githubOAuth:    githubOAuth,
		sessionManager: sm,
	}
}

// GET /api/users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.storage.ListUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, user := range users {
		user.PasswordHash = ""
		user.TOTPSecret = nil
	}

	if users == nil {
		users = []*storage.User{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// GET /api/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.storage.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	user.PasswordHash = ""
	user.TOTPSecret = nil
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	if req.Email == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "email and password required"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	if len(req.Password) < 8 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "password too short (min 8)"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	exists, err := h.storage.CheckEmailExists(r.Context(), req.Email)
	if err != nil {
		exists = false
	}
	if exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "email already exists"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "failed to hash password"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	totpSecret, err := h.totp.GenerateSecret()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate TOTP secret"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	user := &storage.User{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hashedPassword),
		Role:         "user",
		TOTPSecret:   &totpSecret,
	}

	if err := h.storage.CreateUser(r.Context(), user); err != nil {
		log.Printf("CreateUser error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "failed to create user"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"user":        user,
		"totp_secret": totpSecret,
		"qr_code_url": h.totp.GenerateQRURL(req.Email, totpSecret),
	}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		TOTPCode string `json:"totp_code,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	if req.Email == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "email and password required"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	user, err := h.storage.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	if user.TOTPSecret != nil && *user.TOTPSecret != "" {
		if req.TOTPCode == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "totp_code required"}); err != nil {
				log.Printf("Failed to encode response: %v", err)
			}
			return
		}
		if !h.totp.ValidateCode(*user.TOTPSecret, req.TOTPCode) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "invalid totp code"}); err != nil {
				log.Printf("Failed to encode response: %v", err)
			}
			return
		}
	}
	sessionID := uuid.New().String()
	realIP := getRealIP(r) 
	h.sessionManager.Create(sessionID, user.ID, user.Email, user.Role, realIP, r.UserAgent())
	token, err := h.jwt.GenerateTokenWithSessionID(user.ID, user.Role, sessionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate token"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	if err := h.storage.UpdateLastLogin(r.Context(), user.ID); err != nil {
		log.Printf("Failed to update last login for user %s: %v", user.ID, err)
	}

	user.PasswordHash = ""
	user.TOTPSecret = nil

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  user,
	}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID := auth.GetSessionID(r.Context())
	log.Printf("[DEBUG] Logout called with sessionID: %s", sessionID)

	if sessionID != "" {
		if sess, ok := h.sessionManager.Get(sessionID); ok {
			log.Printf("[DEBUG] Session found for user: %s, deleting...", sess.UserID)
			h.sessionManager.Delete(sessionID)
			log.Printf("[DEBUG] Session deleted")
		} else {
			log.Printf("[DEBUG] Session NOT found: %s", sessionID)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func generateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *UserHandler) KeycloakLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateOAuthState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   true,
	})
	url := h.oauth.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *UserHandler) KeycloakCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := h.oauth.ExchangeCode(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := h.oauth.GetUserInfo(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := h.storage.GetOrCreateUserByOAuth(r.Context(), "keycloak", userInfo.ID, userInfo.Email, userInfo.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionID := uuid.New().String()
	h.sessionManager.Create(sessionID, user.ID, user.Email, user.Role, r.RemoteAddr, r.UserAgent())
	jwtToken, err := h.jwt.GenerateTokenWithSessionID(user.ID, user.Role, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""
	user.TOTPSecret = nil

	h.redirectOAuthSuccess(w, r, jwtToken, user)
}

func (h *UserHandler) redirectOAuthSuccess(w http.ResponseWriter, r *http.Request, token string, user *storage.User) {
	frontendURL := getEnvOrDefault("FRONTEND_URL", "http://localhost:5173")
	encodedUser, err := json.Marshal(map[string]string{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectTo := fmt.Sprintf("%s/auth/oauth/callback?token=%s&user=%s", frontendURL, url.QueryEscape(token), url.QueryEscape(string(encodedUser)))
	http.Redirect(w, r, redirectTo, http.StatusTemporaryRedirect)
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	user, err := h.storage.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if err := h.storage.UpdateUser(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.PasswordHash = ""
	user.TOTPSecret = nil
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.storage.DeleteUser(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.PathValue("email")
	user, err := h.storage.GetUserByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	user.PasswordHash = ""
	user.TOTPSecret = nil
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		http.Error(w, "Old password and new password are required", http.StatusBadRequest)
		return
	}
	user, err := h.storage.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		http.Error(w, "Invalid old password", http.StatusUnauthorized)
		return
	}
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash new password", http.StatusInternalServerError)
		return
	}
	if err := h.storage.UpdatePassword(r.Context(), id, string(hashedNewPassword)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) CheckEmailExists(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}
	exists, err := h.storage.CheckEmailExists(r.Context(), email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"exists": exists}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *UserHandler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.AvatarURL == "" {
		http.Error(w, "avatar_url is required", http.StatusBadRequest)
		return
	}
	if err := h.storage.UpdateAvatar(r.Context(), id, req.AvatarURL); err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetUserByOAuth(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	oauthId := r.PathValue("oauthId")
	if provider == "" || oauthId == "" {
		http.Error(w, "provider and oauthId are required", http.StatusBadRequest)
		return
	}
	user, err := h.storage.GetUserByOAuth(r.Context(), provider, oauthId)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.PasswordHash = ""
	user.TOTPSecret = nil
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := h.storage.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.PasswordHash = ""
	user.TOTPSecret = nil
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *UserHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateOAuthState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   true,
	})
	url := h.googleOAuth.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *UserHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := h.googleOAuth.ExchangeCode(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := h.googleOAuth.GetUserInfo(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := h.storage.GetOrCreateUserByOAuth(r.Context(), "google", userInfo.ID, userInfo.Email, userInfo.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionID := uuid.New().String()
	h.sessionManager.Create(sessionID, user.ID, user.Email, user.Role, r.RemoteAddr, r.UserAgent())
	jwtToken, err := h.jwt.GenerateTokenWithSessionID(user.ID, user.Role, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""
	user.TOTPSecret = nil

	h.redirectOAuthSuccess(w, r, jwtToken, user)
}

func (h *UserHandler) GithubLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateOAuthState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   true,
	})
	url := h.githubOAuth.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *UserHandler) GithubCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := h.githubOAuth.ExchangeCode(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := h.githubOAuth.GetUserInfo(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := h.storage.GetOrCreateUserByOAuth(r.Context(), "github", userInfo.ID, userInfo.Email, userInfo.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionID := uuid.New().String()
	h.sessionManager.Create(sessionID, user.ID, user.Email, user.Role, r.RemoteAddr, r.UserAgent())
	jwtToken, err := h.jwt.GenerateTokenWithSessionID(user.ID, user.Role, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	user.PasswordHash = ""
	user.TOTPSecret = nil
	h.redirectOAuthSuccess(w, r, jwtToken, user)
}
