package iceberg

import (
	sqlc "lakelens/internal/sqlc/generate"
	"lakelens/internal/stash"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)



type IcebergService struct {
	Queries *sqlc.Queries
	RedisClient *redis.Client
	DB *pgxpool.Pool

	Stash *stash.StashService
}

func NewIcebergService(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool, stash *stash.StashService) *IcebergService {
	return &IcebergService{
		Queries: queries,
		RedisClient: redis,
		DB: db,
		
		Stash: stash,
	}
} 