package internal

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func GetLobbyHandler(redisClient *RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lobbyID := chi.URLParam(r, "lobbyID")
		if lobbyID == "" {
			http.Error(w, "lobby ID is required", http.StatusBadRequest)
			return
		}

		lobby, err := redisClient.GetLobby(lobbyID)
		if err != nil {
			log.Printf("Failed to fetch lobby: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if lobby == nil {
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(map[string]string{
				"error": "lobby not found",
			}); err != nil {
				log.Printf("Failed to encode not found response: %v", err)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(lobby); err != nil {
			log.Printf("Failed to encode lobby response: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}
