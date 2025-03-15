package dto

import (
	"github.com/xitongsys/parquet-go/parquet"
)

type NewLake struct {
	AccessID string
	AccessKey string
	LakeRegion string
}

type CompleteResponse struct {
	Parquet []*ParquetClean
	Iceberg []*IcebergClean
}

type ParquetClean struct {
	Schema []*parquet.SchemaElement
}

type IcebergClean struct {
}


// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Lake Specific

type BucketList map[*BucketData]*AllFilesMp

type AllFilesMp map[string]interface{}

type BucketData struct {
	Name string
	StorageType string
	LeafFilePaths []string
	Parquet bool // true if the bucket is a parquet bucket
	Iceberg bool // true if the bucket is an iceberg table bucket

	Unknown bool // true if unindentified type
}