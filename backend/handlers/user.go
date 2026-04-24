// backend/handlers/user.go
package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"humanguard/auth"
	"humanguard/storage"
)

type UserHandler struct {
	storage storage.Storage
	jwt     *auth.JWTService
	totp    *auth.TOTPService
	oauth   *auth.OAuthService
}

func NewUserHandler(store storage.Storage, jwtService *auth.JWTService, totpService *auth.TOTPService, oauthService *auth.OAuthService) *UserHandler {
	return &UserHandler{
		storage: store,
		jwt:     jwtService,
		totp:    totpService,
		oauth:   oauthService,
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
	json.NewEncoder(w).Encode(users)
}

// GET /api/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.storage.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
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
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if req.Email == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "email and password required"})
		return
	}

	if len(req.Password) < 8 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "password too short (min 8)"})
		return
	}

	exists, _ := h.storage.CheckEmailExists(r.Context(), req.Email)
	if exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "email already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to hash password"})
		return
	}

	totpSecret := h.totp.GenerateSecret()

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
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create user"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":        user,
		"totp_secret": totpSecret,
		"qr_code_url": h.totp.GenerateQRURL(req.Email, totpSecret),
	})
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
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if req.Email == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "email and password required"})
		return
	}

	user, err := h.storage.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
		return
	}

	if user.TOTPSecret != nil && *user.TOTPSecret != "" {
		if req.TOTPCode == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "totp_code required"})
			return
		}
		if !h.totp.ValidateCode(*user.TOTPSecret, req.TOTPCode) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid totp code"})
			return
		}
	}

	token, err := h.jwt.GenerateToken(user.ID, user.Role)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate token"})
		return
	}

	h.storage.UpdateLastLogin(r.Context(), user.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func generateOAuthState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *UserHandler) KeycloakLogin(w http.ResponseWriter, r *http.Request) {
	state := generateOAuthState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
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

	jwtToken, err := h.jwt.GenerateToken(user.ID, user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": jwtToken,
		"user":  user,
	})
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
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
	json.NewEncoder(w).Encode(map[string]bool{"exists": exists})
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
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
	json.NewEncoder(w).Encode(user)
}