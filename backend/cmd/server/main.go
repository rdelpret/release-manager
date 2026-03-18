package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/rdelpret/music-release-planner/backend/internal/auth"
	"github.com/rdelpret/music-release-planner/backend/internal/handler"
	"github.com/rdelpret/music-release-planner/backend/internal/store"
)

func main() {
	godotenv.Load()
	godotenv.Load("../.env")

	// Initialize auth
	auth.Initialize()

	// Initialize database
	db, err := store.New()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Wire up user upsert for OAuth callback
	auth.SetUserUpsertFunc(func(ctx context.Context, email, name string, avatarURL *string) (string, error) {
		user, err := db.UpsertUser(ctx, email, name, avatarURL)
		if err != nil {
			return "", err
		}
		return user.ID, nil
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := handler.NewServer(db)
	fmt.Printf("Server running at http://localhost:%s\n", port)
	if err := srv.Start(port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
