package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func JoinQueueHandler(redisClient *RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p Player

		// Parse the JSON request body
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Validate fields
		if p.PlayerID == "" {
			http.Error(w, "player_id is required", http.StatusBadRequest)
			return
		}

		if p.SkillRating <= 0 {
			http.Error(w, "skill_rating must be > 0", http.StatusBadRequest)
			return
		}

		// Store in Redis queue
		if err := redisClient.PushPlayerToQueue(&p); err != nil {
			http.Error(w, fmt.Sprintf("Failed to enqueue player: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Player enqueued successfully"))
	}
}
