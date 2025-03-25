package dto

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"main.go/internal/consts/errs"
)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Internal

type ClientSave struct {
	LastUsed time.Time
	S3Client *s3.Client
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// User Requests
type NewUser struct {
	Email string
	Password string
}

type NewLake struct {
	Name string

	AccessID string
	AccessKey string
	LakeRegion string
}


// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// User Responses

type BucketData struct {
	Name *string
	StorageType string
	Region *string
	CreationDate *time.Time
	TableType string
}

type NewBucket struct {
	Data BucketData
	Parquet IsParquet
	Iceberg IsIceberg
	Delta IsDelta
	Hudi IsHudi
	Errors []*errs.Errorf
}


// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Internal Operations

type Latency struct {
	Start int64
	ListBuckets int64
	DetermineTableType int64
	Handle int64
}