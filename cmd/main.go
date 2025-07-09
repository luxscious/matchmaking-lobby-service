package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/luxscious/matchmaking-lobby-service/internal"
)

func main() {
	log.Println("Go matchmaking service starting...")

	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, relying on environment variables.")
	}

	// Parse Redis DB
	redisDBStr := os.Getenv("REDIS_DB")
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		log.Fatalf("Invalid REDIS_DB: %v", err)
	}

	// Create Redis client
	redisClient := internal.NewRedisClient(
		os.Getenv("REDIS_ADDR"),
		os.Getenv("REDIS_PASSWORD"),
		redisDB,
	)
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis client: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the matchmaking loop
	log.Println("Starting matchmaking loop...")
	internal.StartMatchmakingLoop(ctx, redisClient)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/ws/{playerID}", internal.WebSocketHandler)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Matchmaking API is up!")); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	})
	r.Post("/queue", internal.JoinQueueHandler(redisClient))

	// Get server port
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	log.Printf("Listening on :%s...", serverPort)

	if err := http.ListenAndServe(":"+serverPort, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
