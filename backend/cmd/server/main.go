package main

import (
	"humanguard/handlers"
	"humanguard/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	store := connectToDatabase()
	defer store.Close()
	startHTTPServer(store)
	waitForShutdown()
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

func startHTTPServer(store storage.Storage) {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// User endpoints
	{
		userHandler := handlers.NewUserHandler(store)

		mux.HandleFunc("GET /api/users/{id}", userHandler.GetUser)
		mux.HandleFunc("GET /api/users/email/{email}", userHandler.GetUserByEmail)
		mux.HandleFunc("GET /api/users/exists", userHandler.CheckEmailExists)
		mux.HandleFunc("POST /api/users", userHandler.CreateUser)
		mux.HandleFunc("PUT /api/users/{id}", userHandler.UpdateUser)
		mux.HandleFunc("DELETE /api/users/{id}", userHandler.DeleteUser)
		mux.HandleFunc("POST /api/users/{id}/password", userHandler.ChangePassword)
		mux.HandleFunc("POST /api/login", userHandler.Login)
	}

	// Site endpoints
	{
		siteHandler := handlers.NewSiteHandler(store)

		mux.HandleFunc("POST /api/sites", siteHandler.CreateSite)
		mux.HandleFunc("GET /api/sites", siteHandler.ListSites)
		mux.HandleFunc("GET /api/sites/{id}", siteHandler.GetSite)
		mux.HandleFunc("PUT /api/sites/{id}", siteHandler.UpdateSite)
		mux.HandleFunc("DELETE /api/sites/{id}", siteHandler.DeleteSite)
		mux.HandleFunc("POST /api/sites/{id}/activate", siteHandler.ActivateSite)
		mux.HandleFunc("POST /api/sites/{id}/suspend", siteHandler.SuspendSite)
		mux.HandleFunc("GET /api/sites/{id}/settings", siteHandler.GetSiteSettings)
		mux.HandleFunc("PUT /api/sites/{id}/settings", siteHandler.UpdateSiteSettings)
	}

	// Session endpoints
	{
		sessionHandler := handlers.NewSessionHandler(store)

		mux.HandleFunc("POST /api/sessions", sessionHandler.CreateSession)
		mux.HandleFunc("GET /api/sessions/{id}", sessionHandler.GetSession)
		mux.HandleFunc("DELETE /api/sessions/{id}", sessionHandler.DeactivateSession)
		mux.HandleFunc("POST /api/sessions/{id}/block", sessionHandler.BlockSession)
		mux.HandleFunc("POST /api/sessions/{id}/unblock", sessionHandler.UnblockSession)
		mux.HandleFunc("PATCH /api/sessions/{id}/risk", sessionHandler.UpdateRiskScore)
		mux.HandleFunc("GET /api/sites/{id}/sessions", sessionHandler.GetSessionsBySite)
		mux.HandleFunc("GET /api/sites/{id}/sessions/suspicious", sessionHandler.GetSuspiciousSessions)
		mux.HandleFunc("GET /api/sites/{id}/stats", sessionHandler.GetSessionStats)
	}

	// create server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// start server
	go func() {
		log.Println("Server starting on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
}

// for gracefully shutdown
func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Received shutdown signal")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
