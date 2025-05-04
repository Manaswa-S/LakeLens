package db

import (
	"context"
	"fmt"
	sqlc "lakelens/internal/sqlc/generate"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)


var	Pool *pgxpool.Pool
var	QueriesPool *sqlc.Queries
var	RedisClient *redis.Client
var err error

func InitDB() (error) {
	fmt.Println("Connecting to Databases and Cache...")
	// create context object
	ctx := context.Background()
	
	// initialize database
	dbConn := os.Getenv("DBLoginCredentials")
	Pool, err = pgxpool.New(ctx, dbConn)
	if err != nil {
		return fmt.Errorf("error creating database pool: %s", err)
	}

	// inittialize queries pool
	QueriesPool = sqlc.New(Pool)



	// sqlc.New(pool).WithTx()

	
	// // Connect to redis client
	// RedisClient = redis.NewClient(&redis.Options{
	// 	Addr: os.Getenv("RedisAddress"),
	// 	Password: os.Getenv("RedisPassword"),
	// 	DB: 0,
	// 	Protocol: 2,
	// })
	// _, err = RedisClient.Ping(ctx).Result()
	// if err != nil {
	// 	return fmt.Errorf("failed to connect to Redis: %v", err)
	// }
	// fmt.Println("Redis connection is alive!")

	conn, err := Pool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("pgx Pool connection failed: %v", err)
	}
	err = conn.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("database connection failed: %v", err)
	} else {
		fmt.Println("Database connection is alive!")
	}


	return nil
}

// Close DB and Redis connections
func Close() (error) {
	fmt.Println("Closing connections to Databases and Cache...")
	if Pool != nil {
		Pool.Close()
	}
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
