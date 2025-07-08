package internal

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
	Ctx    context.Context
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
