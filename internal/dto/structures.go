package dto

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
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
type BucketDataResponse struct {
	Data *BucketData
	Metadata *CompleteResponse
}

type CompleteResponse struct {
	Parquet []*ParquetClean
	Iceberg []*IcebergClean

	Unknown any
}


type BucketData struct {
	Name string
	StorageType string

	Parquet IsParquet 
	Iceberg IsIceberg 

	Unknown bool // true if unindentified type
}




// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Internal Operations

type BucketDataList map[*BucketData]*AllFilesMp

type AllFilesMp map[string]interface{}



