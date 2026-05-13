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

    // Health check
    mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"ok"}`))
    })

    // Core services
    jwtService := auth.NewJWTService(getEnv("JWT_SECRET", "super-secret-key"))
    totpService := auth.NewTOTPService()
    userSessionManager := auth.NewUserSessionManager(24 * time.Hour)

    oauthService := auth.NewOAuthService(
        "humanguard",
        getEnv("OAUTH_CLIENT_SECRET", "1meWH6qPeEhd17APBADgo20Mth1J5pzP"),
        "http://localhost:8080/api/auth/keycloak/callback",
        getEnv("KEYCLOAK_URL", "http://localhost:8081"),
    )

    // User handler
    userHandler := handlers.NewUserHandler(store, jwtService, totpService, oauthService, userSessionManager)
    authMiddleware := auth.NewAuthMiddleware(jwtService, userSessionManager, store)
    
    // Role middleware
    adminOnly := auth.RequireAdmin()

    // Public user endpoints (no auth required)
    mux.HandleFunc("POST /api/users", userHandler.CreateUser)
    mux.HandleFunc("POST /api/login", userHandler.Login)
    mux.HandleFunc("GET /api/auth/keycloak/login", userHandler.KeycloakLogin)
    mux.HandleFunc("GET /api/auth/keycloak/callback", userHandler.KeycloakCallback)

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
    visitorSessionHandler := handlers.NewVisitorSessionHandler(store)
    mux.Handle("POST /api/sessions", authMiddleware.Middleware(http.HandlerFunc(visitorSessionHandler.CreateSession)))
    mux.Handle("GET /api/sessions/{id}", authMiddleware.Middleware(http.HandlerFunc(visitorSessionHandler.GetSession)))
    mux.Handle("DELETE /api/sessions/{id}", authMiddleware.Middleware(http.HandlerFunc(visitorSessionHandler.DeactivateSession)))
    mux.Handle("POST /api/sessions/{id}/block", authMiddleware.Middleware(http.HandlerFunc(visitorSessionHandler.BlockSession)))
    mux.Handle("POST /api/sessions/{id}/unblock", authMiddleware.Middleware(http.HandlerFunc(visitorSessionHandler.UnblockSession)))
    mux.Handle("PATCH /api/sessions/{id}/risk", authMiddleware.Middleware(http.HandlerFunc(visitorSessionHandler.UpdateRiskScore)))
    mux.Handle("POST /api/sessions/{id}/activity", authMiddleware.Middleware(http.HandlerFunc(visitorSessionHandler.UpdateSessionActivity)))
    mux.Handle("POST /api/sessions/{id}/captcha", authMiddleware.Middleware(http.HandlerFunc(visitorSessionHandler.MarkCaptchaShown)))
    mux.Handle("POST /api/sessions/cleanup", authMiddleware.Middleware(http.HandlerFunc(visitorSessionHandler.CleanupExpiredSessions)))
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
            fs = storage.NewLocalS3("./data/uploads")
        } else {
            fs = minioClient
            log.Println("Connected to MinIO storage")
        }
    } else {
        fs = storage.NewLocalS3("./data/uploads")
        log.Println("Using local file storage")
    }
    
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

    // Rate limiting rules
    rules := []middleware.Rule{
        {Pattern: regexp.MustCompile(`^/api/login$`), Limit: 10.0 / 60.0, Burst: 5},
        {Pattern: regexp.MustCompile(`^/api/users$`), Limit: 10.0 / 60.0, Burst: 5},
        {Pattern: regexp.MustCompile(`^/api/files/upload`), Limit: 20.0 / 60.0, Burst: 10},
        {Pattern: regexp.MustCompile(`^/api/sessions/.*/behavior$`), Limit: 300.0 / 60.0, Burst: 100},
        {Pattern: regexp.MustCompile(`^/api/sessions/.*/analyze$`), Limit: 30.0 / 60.0, Burst: 10},
        {Pattern: regexp.MustCompile(`^/api/sessions$`), Limit: 30.0 / 60.0, Burst: 10},
        {Pattern: regexp.MustCompile(`^/api/sites`), Limit: 30.0 / 60.0, Burst: 10},
        {Pattern: regexp.MustCompile(`^/api/keys`), Limit: 10.0 / 60.0, Burst: 5},
        {Pattern: regexp.MustCompile(`^/api/admin/users/sessions`), Limit: 20.0 / 60.0, Burst: 10},
    }
    rateLimiter := middleware.NewRateLimiter(rules)
    handler = rateLimiter.Middleware(handler)

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
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
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