package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"humanguard/storage"
)

func main() {
	cfg := &storage.Config{
		DBURL:       getEnv("DATABASE_URL", "postgres://postgres:123@localhost:5432/humanguard?sslmode=disable"),
		UploadDir:   getEnv("UPLOAD_DIR", "./data/uploads"),
		MaxFileSize: 100 * 1024 * 1024,
	}

	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer store.Close()

	log.Println("Connected to database")

	if err := store.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}
	log.Println(" Database ping successful")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}