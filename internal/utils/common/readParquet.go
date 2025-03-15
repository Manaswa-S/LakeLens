package cutils

import (
	"fmt"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"main.go/internal/dto"
)

func ReadParquetFileMetadata(filePaths []string) ([]*dto.ParquetClean, error) {
	
	cleanParquet := make([]*dto.ParquetClean, 0)

	for _, path := range filePaths {

		fr, err := local.NewLocalFileReader(path)
		if err != nil {
			fmt.Println(err)
			continue
		}

		pr, err := reader.NewParquetReader(fr, nil, 4)
		if err != nil {
			fmt.Printf("%v ::: %s\n", err, path)
			continue
		}

		cleanParq := CleanParquet(pr)
		cleanParquet = append(cleanParquet, cleanParq)
		

		pr.ReadStop()
		fr.Close()
	}

	return cleanParquet, nil
}

