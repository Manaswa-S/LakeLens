package parqutils

import (
	"lakelens/internal/dto"

	"github.com/xitongsys/parquet-go/reader"
)

// CleanParquet extracts only the required data from entire structures of metadata.
//
// The things to extract is fixed for now.
func CleanParquet(parquet *reader.ParquetReader) (*dto.ParquetClean) {

	cleanParq := dto.ParquetClean{
		Schema: parquet.Footer.Schema,
		CreatedBy: parquet.Footer.CreatedBy,
		Version: parquet.Footer.Version,
		NumRows: parquet.Footer.NumRows,
		
		EncryptionAlgo: parquet.Footer.EncryptionAlgorithm,
	}	

	return &cleanParq
}

