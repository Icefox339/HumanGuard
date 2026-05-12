package main

import (
	"context"
	"humanguard/auth"
	"humanguard/handlers"
	"humanguard/middleware"
	"humanguard/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

func main() {
	store := connectToDatabase()
	defer store.Close()

	server := startHTTPServer(store)
	waitForShutdown(server)

	log.Println("Shutting down...")
}

func connectToDatabase() storage.Storage {
	cfg := &storage.Config{
		DBURL:       getEnv("DATABASE_URL", "postgres://postgres:123@localhost:5432/humanguard?sslmode=disable"),
		UploadDir:   getEnv("UPLOAD_DIR", "./data/uploads"),
		MaxFileSize: 100 * 1024 * 1024,
	}

	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Connected to database")

	if err := store.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}
	log.Println("Database ping successful")

	return store
}

func startHTTPServer(store storage.Storage) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	jwtService := auth.NewJWTService(getEnv("JWT_SECRET", "super-secret-key"))
	totpService := auth.NewTOTPService()

	oauthService := auth.NewOAuthService(
		"humanguard",
		getEnv("OAUTH_CLIENT_SECRET", "1meWH6qPeEhd17APBADgo20Mth1J5pzP"),
		"http://localhost:8080/api/auth/keycloak/callback",
		getEnv("KEYCLOAK_URL", "http://localhost:8081"),
	)

	authMiddleware := auth.AuthMiddleware(jwtService)
	sessionBlockMiddleware := middleware.SessionBlockMiddleware(store)
	combinedAuthMiddleware := func(next http.Handler) http.Handler {
		return sessionBlockMiddleware(authMiddleware(next))
	}

	{
		userHandler := handlers.NewUserHandler(store, jwtService, totpService, oauthService)

		mux.HandleFunc("POST /api/users", userHandler.CreateUser)
		mux.HandleFunc("POST /api/login", userHandler.Login)

		mux.HandleFunc("GET /api/auth/keycloak/login", userHandler.KeycloakLogin)
		mux.HandleFunc("GET /api/auth/keycloak/callback", userHandler.KeycloakCallback)

		// Protected
		mux.Handle("GET /api/users", combinedAuthMiddleware(http.HandlerFunc(userHandler.ListUsers)))
		mux.Handle("GET /api/me", combinedAuthMiddleware(http.HandlerFunc(userHandler.GetCurrentUser)))
		mux.Handle("GET /api/users/{id}", combinedAuthMiddleware(http.HandlerFunc(userHandler.GetUser)))
		mux.Handle("GET /api/users/email/{email}", combinedAuthMiddleware(http.HandlerFunc(userHandler.GetUserByEmail)))
		mux.Handle("GET /api/users/exists", combinedAuthMiddleware(http.HandlerFunc(userHandler.CheckEmailExists)))
		mux.Handle("GET /api/users/oauth/{provider}/{oauthId}", combinedAuthMiddleware(http.HandlerFunc(userHandler.GetUserByOAuth)))
		mux.Handle("PUT /api/users/{id}", combinedAuthMiddleware(http.HandlerFunc(userHandler.UpdateUser)))
		mux.Handle("DELETE /api/users/{id}", combinedAuthMiddleware(http.HandlerFunc(userHandler.DeleteUser)))
		mux.Handle("POST /api/users/{id}/password", combinedAuthMiddleware(http.HandlerFunc(userHandler.ChangePassword)))
		mux.Handle("POST /api/users/{id}/avatar", combinedAuthMiddleware(http.HandlerFunc(userHandler.UpdateAvatar)))
	}

	{
		siteHandler := handlers.NewSiteHandler(store)

		mux.Handle("POST /api/sites", combinedAuthMiddleware(http.HandlerFunc(siteHandler.CreateSite)))
		mux.Handle("GET /api/sites", combinedAuthMiddleware(http.HandlerFunc(siteHandler.ListSites)))
		mux.Handle("GET /api/sites/{id}", combinedAuthMiddleware(http.HandlerFunc(siteHandler.GetSite)))
		mux.Handle("PUT /api/sites/{id}", combinedAuthMiddleware(http.HandlerFunc(siteHandler.UpdateSite)))
		mux.Handle("DELETE /api/sites/{id}", combinedAuthMiddleware(http.HandlerFunc(siteHandler.DeleteSite)))
		mux.Handle("POST /api/sites/{id}/activate", combinedAuthMiddleware(http.HandlerFunc(siteHandler.ActivateSite)))
		mux.Handle("POST /api/sites/{id}/suspend", combinedAuthMiddleware(http.HandlerFunc(siteHandler.SuspendSite)))
		mux.Handle("GET /api/sites/{id}/settings", combinedAuthMiddleware(http.HandlerFunc(siteHandler.GetSiteSettings)))
		mux.Handle("PUT /api/sites/{id}/settings", combinedAuthMiddleware(http.HandlerFunc(siteHandler.UpdateSiteSettings)))
	}

	{
		sessionHandler := handlers.NewSessionHandler(store)

		mux.Handle("POST /api/sessions", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.CreateSession)))
		mux.Handle("GET /api/sessions/{id}", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.GetSession)))
		mux.Handle("PUT /api/sessions/{id}", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.UpdateSession)))
		mux.Handle("DELETE /api/sessions/{id}", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.DeactivateSession)))
		mux.Handle("POST /api/sessions/{id}/block", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.BlockSession)))
		mux.Handle("POST /api/sessions/{id}/unblock", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.UnblockSession)))
		mux.Handle("PATCH /api/sessions/{id}/risk", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.UpdateRiskScore)))
		mux.Handle("POST /api/sessions/{id}/activity", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.UpdateSessionActivity)))
		mux.Handle("POST /api/sessions/{id}/captcha", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.MarkCaptchaShown)))
		mux.Handle("POST /api/sessions/cleanup", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.CleanupExpiredSessions)))
		mux.Handle("GET /api/sites/{id}/sessions", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.GetSessionsBySite)))
		mux.Handle("GET /api/sites/{id}/sessions/suspicious", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.GetSuspiciousSessions)))
		mux.Handle("GET /api/sites/{id}/stats", combinedAuthMiddleware(http.HandlerFunc(sessionHandler.GetSessionStats)))
	}

	fs := storage.NewLocalS3("./data/uploads")
	fileHandler := handlers.NewFileHandler(store, fs)
	behaviorHandler := handlers.NewBehaviorHandler(store)

	mux.Handle("POST /api/sessions/{id}/behavior", combinedAuthMiddleware(http.HandlerFunc(behaviorHandler.SubmitBehavior)))
	mux.Handle("POST /api/sessions/{id}/analyze", combinedAuthMiddleware(http.HandlerFunc(behaviorHandler.TriggerAnalysis)))
	mux.Handle("POST /api/files/upload", combinedAuthMiddleware(http.HandlerFunc(fileHandler.Upload)))
	mux.Handle("GET /api/files/{id}", combinedAuthMiddleware(http.HandlerFunc(fileHandler.Download)))
	mux.Handle("DELETE /api/files/{id}", combinedAuthMiddleware(http.HandlerFunc(fileHandler.Delete)))
	mux.Handle("GET /api/files", combinedAuthMiddleware(http.HandlerFunc(fileHandler.List)))
	mux.Handle("POST /api/files/share", combinedAuthMiddleware(http.HandlerFunc(fileHandler.CreateShare)))
	mux.HandleFunc("GET /api/files/share/{token}", fileHandler.GetByShareToken)

	// API Key management endpoints
	apiKeyHandler := handlers.NewAPIKeyHandler(store)
	mux.Handle("POST /api/keys", combinedAuthMiddleware(http.HandlerFunc(apiKeyHandler.CreateAPIKey)))
	mux.Handle("GET /api/keys", combinedAuthMiddleware(http.HandlerFunc(apiKeyHandler.ListAPIKeys)))
	mux.Handle("DELETE /api/keys/{id}", combinedAuthMiddleware(http.HandlerFunc(apiKeyHandler.RevokeAPIKey)))
	mux.Handle("DELETE /api/keys/{id}/permanent", combinedAuthMiddleware(http.HandlerFunc(apiKeyHandler.DeleteAPIKey)))

	handler := loggingMiddleware(corsMiddleware(mux))
	handler = middleware.CSPMiddleware(handler)

	// Rate limiting rules: different limits for different endpoints
	rules := []middleware.Rule{
		// Auth endpoints (strict: 10 requests per minute)
		{
			Pattern: regexp.MustCompile(`^/api/login$`),
			Limit:   10.0 / 60.0,
			Burst:   5,
		},
		{
			Pattern: regexp.MustCompile(`^/api/users$`),
			Limit:   10.0 / 60.0,
			Burst:   5,
		},
		// File upload (moderate: 20 requests per minute)
		{
			Pattern: regexp.MustCompile(`^/api/files/upload`),
			Limit:   20.0 / 60.0,
			Burst:   10,
		},
		// Behavior collection (high volume: 300 requests per minute)
		{
			Pattern: regexp.MustCompile(`^/api/sessions/.*/behavior$`),
			Limit:   300.0 / 60.0,
			Burst:   100,
		},
		// Analysis trigger (moderate: 30 requests per minute)
		{
			Pattern: regexp.MustCompile(`^/api/sessions/.*/analyze$`),
			Limit:   30.0 / 60.0,
			Burst:   10,
		},
		// Session creation (moderate: 30 requests per minute)
		{
			Pattern: regexp.MustCompile(`^/api/sessions$`),
			Limit:   30.0 / 60.0,
			Burst:   10,
		},
		// Site operations (moderate: 30 requests per minute)
		{
			Pattern: regexp.MustCompile(`^/api/sites`),
			Limit:   30.0 / 60.0,
			Burst:   10,
		},
	}

	rateLimiter := middleware.NewRateLimiter(rules)
	handler = rateLimiter.Middleware(handler)

	// API Key authentication (tries to authenticate, falls back to JWT)
	apiKeyAuth := middleware.NewAPIKeyAuthenticator(store)
	handler = apiKeyAuth.Middleware(handler)

	handler = middleware.RequestIDMiddleware(handler)

	server := &http.Server{
		Addr:         ":" + getEnv("PORT", "8080"),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Server starting on http://localhost:" + getEnv("PORT", "8080"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	return server
}

func waitForShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server stopped")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := middleware.GetRequestID(r.Context())
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s %s", requestID, r.Method, r.URL.Path, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
