package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/rdelpret/music-release-planner/backend/internal/handler"
)

func main() {
	godotenv.Load()
	godotenv.Load("../.env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := handler.NewServer()

	fmt.Printf("Server running at http://localhost:%s\n", port)
	if err := srv.Start(port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
