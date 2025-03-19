package parqutils

import (
	"github.com/xitongsys/parquet-go/reader"
	"main.go/internal/dto"
)

// CleanParquet extracts only the required data from entire structures of metadata.
//
// The things to extract is fixed for now.
func CleanParquet(parquet *reader.ParquetReader) (*dto.ParquetClean, error) {

	cleanParq := dto.ParquetClean{
		Schema: parquet.Footer.Schema,
	}

	return &cleanParq, nil
}

