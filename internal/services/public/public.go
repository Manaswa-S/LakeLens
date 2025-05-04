package public

import (
	sqlc "lakelens/internal/sqlc/generate"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)



type PublicService struct {
	Queries *sqlc.Queries
	RedisClient *redis.Client
	DB *pgxpool.Pool
}

func NewPublicService(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool) *PublicService {
	return &PublicService{
		Queries: queries,
		RedisClient: redis,
		DB: db,
	}
} 