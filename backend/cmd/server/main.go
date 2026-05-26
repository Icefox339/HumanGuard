package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"humanguard/auth"
	"humanguard/handlers"
	"humanguard/metrics"
	"humanguard/middleware"
	"humanguard/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"
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

	ensureDefaultAdmin(store)

	return store
}

func ensureDefaultAdmin(store storage.Storage) {
	adminEmail := strings.TrimSpace(getEnv("DEFAULT_ADMIN_EMAIL", ""))
	adminPassword := getEnv("DEFAULT_ADMIN_PASSWORD", "")
	adminName := strings.TrimSpace(getEnv("DEFAULT_ADMIN_NAME", "System Admin"))
	if adminEmail == "" || adminPassword == "" {
		log.Println("Default admin bootstrap skipped: DEFAULT_ADMIN_EMAIL or DEFAULT_ADMIN_PASSWORD is empty")
		return
	}

	if len(adminPassword) < 8 {
		log.Printf("Default admin bootstrap skipped: password for %s is shorter than 8 chars", adminEmail)
		return
	}

	ctx := context.Background()
	if _, err := store.GetUserByEmail(ctx, adminEmail); err == nil {
		log.Printf("Default admin already exists: %s", adminEmail)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Default admin bootstrap failed: password hash error: %v", err)
		return
	}

	adminUser := &storage.User{
		Email:        adminEmail,
		Name:         adminName,
		Role:         "admin",
		PasswordHash: string(passwordHash),
		IsVerified:   true,
	}
	if err := store.CreateUser(ctx, adminUser); err != nil {
		log.Printf("Default admin bootstrap failed: create user error: %v", err)
		return
	}

	log.Printf("Default admin created: %s", adminEmail)
}

func startHTTPServer(store storage.Storage) *http.Server {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			log.Printf("Failed to write health response: %v", err)
		}
	})

	// Core services
	jwtService := auth.NewJWTService(getEnv("JWT_SECRET", "super-secret-key"))
	totpService := auth.NewTOTPService()
	userSessionManager := auth.NewUserSessionManager(24 * time.Hour)

	oauthService := auth.NewOAuthService(
		"keycloak",
		getEnv("OAUTH_CLIENT_ID", "humanguard"),
		getEnv("OAUTH_CLIENT_SECRET", "1meWH6qPeEhd17APBADgo20Mth1J5pzP"),
		getEnv("KEYCLOAK_REDIRECT_URL", "http://localhost:8080/api/auth/keycloak/callback"),
	)

	googleOAuth := auth.NewOAuthService(
		"google",
		getEnv("GOOGLE_CLIENT_ID", ""),
		getEnv("GOOGLE_CLIENT_SECRET", ""),
		"http://localhost:8080/api/auth/google/callback",
	)

	githubOAuth := auth.NewOAuthService(
		"github",
		getEnv("GITHUB_CLIENT_ID", ""),
		getEnv("GITHUB_CLIENT_SECRET", ""),
		"http://localhost:8080/api/auth/github/callback",
	)
	// User handler
	userHandler := handlers.NewUserHandler(
		store,
		jwtService,
		totpService,
		oauthService, // Keycloak
		googleOAuth,  // Google
		githubOAuth,  // GitHub
		userSessionManager,
	)
	authMiddleware := auth.NewAuthMiddleware(jwtService, userSessionManager, store)

	// Role middleware
	adminOnly := auth.RequireAdmin()
	behaviorHandler := handlers.NewBehaviorHandler(store)
	visitorSessionHandler := handlers.NewVisitorSessionHandler(store)

	mux.Handle("GET /metrics", promhttp.Handler())
	// Public user endpoints (no auth required)
	mux.HandleFunc("POST /api/users", userHandler.CreateUser)
	mux.HandleFunc("POST /api/login", userHandler.Login)
	mux.HandleFunc("GET /api/auth/keycloak/login", userHandler.KeycloakLogin)
	mux.HandleFunc("GET /api/auth/keycloak/callback", userHandler.KeycloakCallback)
	mux.HandleFunc("GET /api/auth/google/login", userHandler.GoogleLogin)
	mux.HandleFunc("GET /api/auth/google/callback", userHandler.GoogleCallback)
	mux.HandleFunc("GET /api/auth/github/login", userHandler.GithubLogin)
	mux.HandleFunc("GET /api/auth/github/callback", userHandler.GithubCallback)
	mux.HandleFunc("POST /api/check", visitorSessionHandler.CheckRequest)     // nginx
	mux.HandleFunc("POST /api/behavior/{id}", behaviorHandler.SubmitBehavior) // JS
	mux.HandleFunc("GET /api/csrf", middleware.CSRFTokenHandler)

	// Authenticated endpoints (any valid JWT or API key)
	mux.Handle("POST /api/logout", authMiddleware.Middleware(http.HandlerFunc(userHandler.Logout)))
	mux.Handle("GET /api/me", authMiddleware.Middleware(http.HandlerFunc(userHandler.GetCurrentUser)))
	mux.Handle("GET /api/users/email/{email}", authMiddleware.Middleware(http.HandlerFunc(userHandler.GetUserByEmail)))
	mux.Handle("GET /api/users/exists", authMiddleware.Middleware(http.HandlerFunc(userHandler.CheckEmailExists)))
	mux.Handle("GET /api/users/oauth/{provider}/{oauthId}", authMiddleware.Middleware(http.HandlerFunc(userHandler.GetUserByOAuth)))
	mux.Handle("POST /api/users/{id}/password", authMiddleware.Middleware(http.HandlerFunc(userHandler.ChangePassword)))
	mux.Handle("POST /api/users/{id}/avatar", authMiddleware.Middleware(http.HandlerFunc(userHandler.UpdateAvatar)))
	mux.Handle("GET /api/users/{id}", authMiddleware.Middleware(http.HandlerFunc(userHandler.GetUser)))
	mux.Handle("PUT /api/users/{id}", authMiddleware.Middleware(http.HandlerFunc(userHandler.UpdateUser)))

	// Admin only user management
	mux.Handle("GET /api/users", authMiddleware.Middleware(adminOnly(http.HandlerFunc(userHandler.ListUsers))))
	mux.Handle("DELETE /api/users/{id}", authMiddleware.Middleware(adminOnly(http.HandlerFunc(userHandler.DeleteUser))))

	// Admin user sessions management
	userSessionHandler := handlers.NewUserSessionHandler(userSessionManager)
	mux.Handle("GET /api/admin/users/sessions", authMiddleware.Middleware(adminOnly(http.HandlerFunc(userSessionHandler.ListAllUserSessions))))
	mux.Handle("GET /api/admin/users/sessions/stats", authMiddleware.Middleware(adminOnly(http.HandlerFunc(userSessionHandler.GetSessionsStats))))
	mux.Handle("DELETE /api/admin/users/sessions/{session_id}", authMiddleware.Middleware(adminOnly(http.HandlerFunc(userSessionHandler.ForceRevokeSession))))

	// Sites (authenticated users can manage their own sites)
	siteHandler := handlers.NewSiteHandler(store)
	mux.Handle("POST /api/sites", authMiddleware.Middleware(http.HandlerFunc(siteHandler.CreateSite)))
	mux.Handle("GET /api/sites", authMiddleware.Middleware(http.HandlerFunc(siteHandler.ListSites)))
	mux.Handle("GET /api/sites/{id}", authMiddleware.Middleware(http.HandlerFunc(siteHandler.GetSite)))
	mux.Handle("PUT /api/sites/{id}", authMiddleware.Middleware(http.HandlerFunc(siteHandler.UpdateSite)))
	mux.Handle("DELETE /api/sites/{id}", authMiddleware.Middleware(http.HandlerFunc(siteHandler.DeleteSite)))
	mux.Handle("POST /api/sites/{id}/activate", authMiddleware.Middleware(http.HandlerFunc(siteHandler.ActivateSite)))
	mux.Handle("POST /api/sites/{id}/suspend", authMiddleware.Middleware(http.HandlerFunc(siteHandler.SuspendSite)))
	mux.Handle("GET /api/sites/{id}/settings", authMiddleware.Middleware(http.HandlerFunc(siteHandler.GetSiteSettings)))
	mux.Handle("PUT /api/sites/{id}/settings", authMiddleware.Middleware(http.HandlerFunc(siteHandler.UpdateSiteSettings)))

	// Visitor sessions
	mux.Handle("GET /api/sites/{id}/sessions", authMiddleware.Middleware(http.HandlerFunc(visitorSessionHandler.GetSessionsBySite)))
	mux.Handle("GET /api/sites/{id}/sessions/suspicious", authMiddleware.Middleware(http.HandlerFunc(visitorSessionHandler.GetSuspiciousSessions)))
	mux.Handle("GET /api/sites/{id}/stats", authMiddleware.Middleware(http.HandlerFunc(visitorSessionHandler.GetSessionStats)))

	var fs storage.S3Client
	storageType := getEnv("STORAGE_TYPE", "local")
	if storageType == "minio" {
		endpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
		accessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
		secretKey := getEnv("MINIO_SECRET_KEY", "minioadmin123")
		bucket := getEnv("MINIO_BUCKET", "humanguard")
		useSSL := getEnv("MINIO_USE_SSL", "false") == "true"

		minioClient, err := storage.NewMinIOClient(endpoint, accessKey, secretKey, bucket, useSSL)
		if err != nil {
			log.Printf("Warning: Failed to connect to MinIO: %v, falling back to local storage", err)
			fs, err = storage.NewLocalS3("./data/uploads")
			if err != nil {
				log.Printf("Warning: Failed to create local storage: %v", err)
			}
		} else {
			fs = minioClient
			log.Println("Connected to MinIO storage")
		}
	} else {
		var err error
		fs, err = storage.NewLocalS3("./data/uploads")
		if err != nil {
			log.Printf("Warning: Failed to create local storage: %v", err)
		}
		log.Println("Using local file storage")
	}

	fileHandler := handlers.NewFileHandler(store, fs)

	mux.Handle("POST /api/sessions/{id}/analyze", authMiddleware.Middleware(http.HandlerFunc(behaviorHandler.TriggerAnalysis)))
	mux.Handle("POST /api/files/upload", authMiddleware.Middleware(http.HandlerFunc(fileHandler.Upload)))
	mux.Handle("GET /api/files/{id}", authMiddleware.Middleware(http.HandlerFunc(fileHandler.Download)))
	mux.Handle("DELETE /api/files/{id}", authMiddleware.Middleware(http.HandlerFunc(fileHandler.Delete)))
	mux.Handle("GET /api/files", authMiddleware.Middleware(http.HandlerFunc(fileHandler.List)))
	mux.Handle("POST /api/files/share", authMiddleware.Middleware(http.HandlerFunc(fileHandler.CreateShare)))
	mux.HandleFunc("GET /api/files/share/{token}", fileHandler.GetByShareToken)
	mux.Handle("/api/files/upload/progress", authMiddleware.Middleware(http.HandlerFunc(fileHandler.UploadProgressWS)))
	// API keys
	apiKeyHandler := handlers.NewAPIKeyHandler(store)
	mux.Handle("POST /api/keys", authMiddleware.Middleware(http.HandlerFunc(apiKeyHandler.CreateAPIKey)))
	mux.Handle("GET /api/keys", authMiddleware.Middleware(http.HandlerFunc(apiKeyHandler.ListAPIKeys)))
	mux.Handle("DELETE /api/keys/{id}", authMiddleware.Middleware(http.HandlerFunc(apiKeyHandler.RevokeAPIKey)))
	mux.Handle("DELETE /api/keys/{id}/permanent", authMiddleware.Middleware(adminOnly(http.HandlerFunc(apiKeyHandler.DeleteAPIKey))))

	// Global middleware chain
	handler := http.Handler(mux)
	handler = loggingMiddleware(handler)
	handler = corsMiddleware(handler)
	handler = middleware.CSPMiddleware(handler)
	handler = middleware.RequestIDMiddleware(handler)
	handler = middleware.CSRFMiddleware([]string{
		"/api/login",
		"/api/users",
		"/api/check",
		"/api/behavior/",
		"/api/auth/",
	})(handler)

	// Rate limiting rules
	rules := []middleware.Rule{
		{Pattern: regexp.MustCompile(`^/api/login$`), Limit: 5.0 / 60.0, Burst: 5},
		{Pattern: regexp.MustCompile(`^/api/users$`), Limit: 10.0 / 60.0, Burst: 10},
		{Pattern: regexp.MustCompile(`^/api/check$`), Limit: 100.0 / 60.0, Burst: 50},
		{Pattern: regexp.MustCompile(`^/api/behavior/`), Limit: 300.0 / 60.0, Burst: 100},
		{Pattern: regexp.MustCompile(`^/api/files/upload`), Limit: 10.0 / 60.0, Burst: 5},
		{Pattern: regexp.MustCompile(`^/api/files/`), Limit: 30.0 / 60.0, Burst: 20},
		{Pattern: regexp.MustCompile(`^/api/sites`), Limit: 30.0 / 60.0, Burst: 15},
		{Pattern: regexp.MustCompile(`^/api/keys`), Limit: 10.0 / 60.0, Burst: 5},
		{Pattern: regexp.MustCompile(`^/api/admin/`), Limit: 20.0 / 60.0, Burst: 10},
		{Pattern: regexp.MustCompile(`^/api/me`), Limit: 30.0 / 60.0, Burst: 15},
	}
	rateLimiter := middleware.NewRateLimiter(rules)
	handler = rateLimiter.Middleware(handler)

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

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			stats, err := store.GetSessionStats(context.Background(), "")
			if err == nil {
				metrics.ActiveSessions.Set(float64(stats.Active))
				metrics.AverageRiskScore.Set(stats.AvgRisk)
			}

			highRisk, err := store.GetSuspiciousSessions(context.Background(), "", 80)
			if err == nil {
				metrics.HighRiskSessions.Set(float64(len(highRisk)))
			}
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
		authMethod := "none"
		if r.Header.Get("Authorization") != "" {
			authMethod = "jwt"
		} else if r.Header.Get("X-API-Key") != "" {
			authMethod = "api_key"
		}
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s %s (auth: %s)", requestID, r.Method, r.URL.Path, time.Since(start), authMethod)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigin := getEnv("CORS_ORIGIN", "http://localhost:5173")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-CSRF-Token")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining")
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
