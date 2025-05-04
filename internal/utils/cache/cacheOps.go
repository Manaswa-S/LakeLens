package cacheutils

import (
	"lakelens/internal/dto"
	"sync"
	"time"
)


type CacheMetadata struct {

	// trial:
	KeyCount int64

	//

	Bucket *dto.NewBucket
	CreatedAt int64
	UpdatedAt time.Time
}

var (
	S3BucketCache = make(map[string]*CacheMetadata, 0)

	mu = sync.Mutex{}
)

func SetCacheS3Bucket(bucket *dto.NewBucket) {
	DelCacheS3Bucket(*bucket.Data.Name)
	
	mu.Lock()
	S3BucketCache[*bucket.Data.Name] = &CacheMetadata{
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: bucket.Data.UpdatedAt,
		Bucket: bucket,
		//
		KeyCount: bucket.Data.KeyCount,
		//
	}
	mu.Unlock()
}

func GetCacheS3Bucket(bucketName string) (*CacheMetadata, bool) {
	mu.Lock()
	bucData, ok := S3BucketCache[bucketName]
	mu.Unlock()
	return bucData, ok
}

func DelCacheS3Bucket(bucketName string) {
	delete(S3BucketCache, bucketName)
}