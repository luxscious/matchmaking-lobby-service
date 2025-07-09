package internal

import (
	"context"
	"log"
	"sort"
	"time"
)

func StartMatchmakingLoop(ctx context.Context, redisClient *RedisClient) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Matchmaking loop stopped gracefully.")
				return
			default:
				if err := MatchPlayers(ctx, redisClient); err != nil {
					log.Printf("Matchmaking error: %v", err)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

// Finds a group of 5 players with close skill ratings and creates a lobby for them.
func MatchPlayers(ctx context.Context, redisClient *RedisClient) error {
	players, err := redisClient.GetQueuedPlayers()
	if err != nil {
		return err
	}
	if len(players) == 0 {
		return nil
	}

	selected := SelectPlayersForLobby(players, 5, 100)
	if selected == nil {
		log.Println("No suitable group found, waiting...")
		return nil
	}

	// Start transaction
	pipe := redisClient.Client.TxPipeline()

	// Remove selected IDs and metadata
	for _, p := range selected {
		pipe.LRem(ctx, "matchmaking_queue", 1, p.PlayerID)
		pipe.Del(ctx, "player:"+p.PlayerID)
	}

	// Execute the transaction
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	// Create lobby
	lobby, err := redisClient.CreateLobby(ctx, selected)
	if err != nil {
		return err
	}

	log.Printf("Created lobby %s with players: %v", lobby.LobbyID, lobby.PlayerIDs)

	for _, p := range selected {
		if err := NotifyPlayerLobby(p.PlayerID, lobby.LobbyID); err != nil {
			log.Printf("Failed to notify player %s: %v", p.PlayerID, err)
		}
	}

	return nil
}

// Selects a group of players whose skill ratings
// are within the specified threshold.
// Returns nil if no suitable group was found.
func SelectPlayersForLobby(players []*Player, groupSize int, threshold int) []*Player {
	if len(players) < groupSize {
		return nil
	}

	// Sort players by SkillRating ascending
	sort.Slice(players, func(i, j int) bool {
		return players[i].SkillRating < players[j].SkillRating
	})

	for i := 0; i+groupSize-1 < len(players); i++ {
		window := players[i : i+groupSize]
		spread := window[groupSize-1].SkillRating - window[0].SkillRating
		if spread <= threshold {
			return window
		}
	}
	return nil
}
