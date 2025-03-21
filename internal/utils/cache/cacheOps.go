package cacheutils

import "main.go/internal/dto"

var (
	S3BucketCache = make(map[string]*dto.BucketData, 0)
)

func SetCacheS3Bucket(bucket *dto.BucketData) {

	delete(S3BucketCache, *bucket.Name)

	S3BucketCache[*bucket.Name] = bucket
}

func GetCacheS3Bucket(bucketName string) (*dto.BucketData, bool) {

	bucData, ok := S3BucketCache[bucketName]

	return bucData, ok
}

func DelCacheS3Bucket(bucketName string) {
	delete(S3BucketCache, bucketName)
}