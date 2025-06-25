package parqutils

import (
	"lakelens/internal/consts/errs"
	formats "lakelens/internal/dto/formats/parquet"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
)

// ReadParquet reads parquet files.
// It directly returns the cleansed version because of how the readers work.
func ReadParquet(filePath string) (*formats.ParquetClean, *errs.Errorf) {

	fileReader, err := local.NewLocalFileReader(filePath)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrStorageFailed,
			Message: "Failed to open parquet file : " + err.Error(),
		}
	}

	parqReader, err := reader.NewParquetReader(fileReader, nil, 4)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to read parquet file : " + err.Error(),
		}
	}

	cleanParquet := CleanParquet(parqReader)

	parqReader.ReadStop()
	fileReader.Close()

	return cleanParquet, nil
}
