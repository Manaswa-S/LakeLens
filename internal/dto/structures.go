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
type LocationResp struct {
	Data *BucketData
	Metadata *BucketMetadata
}

type BucketData struct {
	Name *string
	StorageType string
	Region *string
	CreationDate *time.Time
	FileTree *FileTreeMap
	TableType string

	Parquet IsParquet 
	Iceberg IsIceberg 

	Unknown bool // true if unindentified type
}

type BucketMetadata struct {
	Parquet []*ParquetClean
	Iceberg []*IcebergClean

	Unknown any
}







// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Internal Operations

type BucketDataList map[*BucketData]*FileTreeMap

type FileTreeMap map[string]interface{}



