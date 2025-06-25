package dto

import (
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto/formats"
	"time"
)

type Latency struct {
	Start              int64
	ListBuckets        int64
	DetermineTableType int64
	Handle             int64
}

// used to send responses when all went right.
type GoodResp struct {
	Message string
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// These are structs that are not intended to be ever put out to the user, neither as req nor as resp.

type BucketData struct {
	Name         string
	StorageType  string
	Region       *string
	CreationDate *time.Time
	TableType    string
	UpdatedAt    time.Time
	//
	KeyCount int64
	//
	LocationID int64
}

type NewBucket struct {
	Data    BucketData
	Parquet formats.IsParquet
	Iceberg formats.IsIceberg
	Delta   formats.IsDelta
	Hudi    formats.IsHudi
	Errors  []*errs.Errorf
}
