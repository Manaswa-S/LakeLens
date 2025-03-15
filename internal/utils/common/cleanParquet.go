package cutils

import (
	"github.com/xitongsys/parquet-go/reader"
	"main.go/internal/dto"
)


func CleanParquet(parqData *reader.ParquetReader) *dto.ParquetClean {
	cleanParq := dto.ParquetClean{
		Schema: parqData.Footer.Schema,
	}

	return &cleanParq
}