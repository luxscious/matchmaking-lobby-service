package internal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func (r *RedisClient) PushPlayerToQueue(p *Player) error {
	// Store player metadata
	playerKey := "player:" + p.PlayerID
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if err := r.Client.Set(r.Ctx, playerKey, data, 0).Err(); err != nil {
		return err
	}

	// Push playerID into the queue
	if err := r.Client.RPush(r.Ctx, "matchmaking_queue", p.PlayerID).Err(); err != nil {
		return err
	}

	return nil
}

func NewRedisClient(addr string, password string, db int) *RedisClient {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,     // e.g., "localhost:6379"
		Password: password, // "" if no password
		DB:       db,       // default 0
	})

	// Test the connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	fmt.Println("Connected to Redis:", addr)

	return &RedisClient{
		Client: rdb,
		Ctx:    ctx,
	}
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}

func (r *RedisClient) PopPlayersFromQueue(count int) ([]*Player, error) {
	ctx := r.Ctx
	var players []*Player

	for i := 0; i < count; i++ {
		data, err := r.Client.LPop(ctx, "matchmaking_queue").Result()
		if err == redis.Nil {
			// no more players in the queue
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to pop player from queue: %w", err)
		}
		var p Player
		if err := json.Unmarshal([]byte(data), &p); err != nil {
			return nil, fmt.Errorf("failed to deserialize player: %w", err)
		}
		players = append(players, &p)
	}
	return players, nil

}

func (r *RedisClient) GetQueuedPlayers() ([]*Player, error) {
	ids, err := r.Client.LRange(r.Ctx, "matchmaking_queue", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var players []*Player
	for _, id := range ids {
		playerKey := "player:" + id
		data, err := r.Client.Get(r.Ctx, playerKey).Result()
		if err != nil {
			return nil, err
		}
		var p Player
		if err := json.Unmarshal([]byte(data), &p); err != nil {
			return nil, err
		}
		players = append(players, &p)
	}

	return players, nil
}
