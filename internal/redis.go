package internal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func (r *RedisClient) PushPlayerToQueue(player *Player) any {
	data, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("failed to serialize player: %w", err)
	}

	if err := r.Client.RPush(r.Ctx, "matchmaking_queue", data).Err(); err != nil {
		return fmt.Errorf("failed to push player to queue: %w", err)
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
