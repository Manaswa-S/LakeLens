package dto

import "github.com/xitongsys/parquet-go/parquet"



type ParquetClean struct {
	Schema []*parquet.SchemaElement
}

type IsParquet struct {
	Present bool
	AllFilePaths []string
}
