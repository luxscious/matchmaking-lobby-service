package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
)

// Creates a new lobby with the given players.
// Returns the Lobby object as a pointer or an error if creation fails.
func (r *RedisClient) CreateLobby(ctx context.Context, players []*Player) (*Lobby, error) {
	lobbyID := uuid.NewString()

	playerIDs := []string{}

	for _, player := range players {
		playerIDs = append(playerIDs, player.PlayerID)
	}

	lobby := Lobby{
		LobbyID:   lobbyID,
		PlayerIDs: playerIDs,
	}

	data, err := json.Marshal(lobby)
	if err != nil {
		log.Printf("failed to serialize lobby: %v", err)
		return nil, err
	}

	key := fmt.Sprintf("lobby:%s", lobbyID)
	if err := r.Client.Set(ctx, key, data, 0).Err(); err != nil {
		log.Printf("failed to store lobby in Redis: %v", err)
		return nil, err
	}
	return &lobby, nil
}
