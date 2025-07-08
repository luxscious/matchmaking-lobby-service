package internal

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Player struct {
	PlayerID    string `json:"player_id"`
	SkillRating int    `json:"skill_rating"`
}
type RedisClient struct {
	Client *redis.Client
	Ctx    context.Context
}
