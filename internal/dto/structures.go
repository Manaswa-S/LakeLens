package dto

import (
	"lakelens/internal/consts/errs"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Internal

type S3ClientSave struct {
	LastUsed time.Time
	S3Client *s3.Client
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// User Requests
type NewUser struct {
	Email    string
	Password string
}

type NewLake struct {
	Name string // the lake project name, whatever the user wants.

	// only one is valid, others remain nil.
	S3    *NewLakeS3
	Azure *NewLakeAzure
	GCP   *NewLakeGCP
}

type NewLakeS3 struct {
	AccessID   string
	AccessKey  string
	LakeRegion string
}

type NewLakeAzure struct {
	// TODO:
}

type NewLakeGCP struct {
	// TODO:
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// User Responses

type BucketData struct {
	Name         *string
	StorageType  string
	Region       *string
	CreationDate *time.Time
	TableType    string
	UpdatedAt    time.Time
	//
	KeyCount int64
	//
}

type NewBucket struct {
	Data    BucketData
	Parquet IsParquet
	Iceberg IsIceberg
	Delta   IsDelta
	Hudi    IsHudi
	Errors  []*errs.Errorf
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Internal Operations

type Latency struct {
	Start              int64
	ListBuckets        int64
	DetermineTableType int64
	Handle             int64
}
