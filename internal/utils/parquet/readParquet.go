package parqutils

import (
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"main.go/internal/dto"
)

// ReadParquet reads parquet files.
// It directly returns the cleansed version because of how the readers work.
func ReadParquet(filePaths []string) ([]*dto.ParquetClean, error) {

	cleanParquets := make([]*dto.ParquetClean, 0)

	for _, path := range filePaths {

		fileReader, err := local.NewLocalFileReader(path)
		if err != nil {
			continue
		} 

		parqReader, err := reader.NewParquetReader(fileReader, nil, 4)
		if err != nil {
			continue
		}

		cleanParq, err := CleanParquet(parqReader)
		if err != nil {
			continue
		}
		cleanParquets = append(cleanParquets, cleanParq)

		parqReader.ReadStop()
		fileReader.Close()
	}

	return cleanParquets, nil
}