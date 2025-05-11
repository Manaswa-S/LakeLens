package stash

import (
	"lakelens/internal/consts"
	"lakelens/internal/dto"
	sqlc "lakelens/internal/sqlc/generate"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type CacheMetadata struct {

	// trial:
	KeyCount int64

	//

	Bucket *dto.NewBucket
	CreatedAt int64
	UpdatedAt time.Time
}

type buckets struct {
	s3 map[string]*CacheMetadata 
	
	// every provider has a separate pool for bucket caching
}


type clients struct {
	S3 map[string]*dto.S3ClientSave

	// each provider type has its own client store
}

type StashService struct {
	// < General
	Queries *sqlc.Queries
	RedisClient *redis.Client
	DB *pgxpool.Pool
	// >

	// < Bucket saves
	buckets *buckets
	bucMU sync.Mutex
	// >

	// < Client saves
	clients *clients
	cliMU sync.Mutex
	// >
}

func NewStashService(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool) *StashService {
	return &StashService{
		Queries: queries,
		RedisClient: redis,
		DB: db,

		buckets: &buckets{
			s3: make(map[string]*CacheMetadata),
		},
		bucMU: sync.Mutex{},

		clients: &clients{
			S3: make(map[string]*dto.S3ClientSave),
		},
		cliMU: sync.Mutex{},
	}
} 

func (c *StashService) SetBucket(bucket *dto.NewBucket) {

	c.bucMU.Lock()
	switch bucket.Data.StorageType {
	case consts.AWSS3:
		// cache in s3
		c.DelBucketS3(bucket.Data.Name)
		c.buckets.s3[bucket.Data.Name] = &CacheMetadata{
			Bucket: bucket,
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: bucket.Data.UpdatedAt,
			KeyCount: bucket.Data.KeyCount,
		}
	// case 
	default:
		// unknown, err out
	}
	c.bucMU.Unlock()
}



func (c *StashService) GetBucketS3(bucketName string) (*CacheMetadata, bool) {
	c.bucMU.Lock()
	bucData, ok := c.buckets.s3[bucketName]
	c.bucMU.Unlock()
	return bucData, ok
}

func (c *StashService) DelBucketS3(bucketName string) {
	delete(c.buckets.s3, bucketName)
}



