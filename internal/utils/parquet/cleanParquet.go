package parqutils

import (
	formats "lakelens/internal/dto/formats/parquet"

	"github.com/xitongsys/parquet-go/reader"
)

// CleanParquet extracts only the required data from entire structures of metadata.
//
// The things to extract is fixed for now.
func CleanParquet(parquet *reader.ParquetReader) *formats.ParquetClean {

	cleanParq := formats.ParquetClean{
		Schema:    parquet.Footer.Schema,
		CreatedBy: parquet.Footer.CreatedBy,
		Version:   parquet.Footer.Version,
		NumRows:   parquet.Footer.NumRows,

		EncryptionAlgo: parquet.Footer.EncryptionAlgorithm,
	}

	return &cleanParq
}
