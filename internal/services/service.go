package services

import (
	"github.com/redis/go-redis/v9"
	sqlc "main.go/internal/sqlc/generate"
)

type Services struct {
	Queries *sqlc.Queries
	RedisClient *redis.Client
}

func NewService(queries *sqlc.Queries, redis *redis.Client) *Services {
	return &Services{
		Queries: queries,
		RedisClient: redis,
	}
} 

func (s *Services) GetList() {
	
}
