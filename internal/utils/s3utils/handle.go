package s3utils

import (
	"fmt"
	"path"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	configs "main.go/internal/config"
	"main.go/internal/consts/errs"
	"main.go/internal/dto"
	iceutils "main.go/internal/utils/iceberg"
	parqutils "main.go/internal/utils/parquet"
)

// HandleIceberg handles downloading, reading and extraction of metadata from given bucket containing Iceberg.
func HandleIceberg(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) (*dto.IcebergClean, *errs.Errorf) {

	resp, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: newBucket.Data.Name,
		Prefix: &newBucket.Iceberg.URI,
	})
	if err != nil {
		return nil, &errs.Errorf{
			Type: errs.ErrServiceUnavailable,
			Message: "Failed to list objects : " + err.Error(),
		}
	}

	for _, obj := range resp.Contents {
		key := *obj.Key
		ext := path.Ext(key)
		switch ext {
		case ".json":
			newBucket.Iceberg.JSONFilePaths = append(newBucket.Iceberg.JSONFilePaths, key)
		case ".avro":
			newBucket.Iceberg.AvroFilePaths = append(newBucket.Iceberg.AvroFilePaths, key)
		}
	}

	jsonPaths := newBucket.Iceberg.JSONFilePaths
	sort.Strings(jsonPaths)

	jsonLen := len(jsonPaths)

	if jsonLen <= 0 {
		return nil, &errs.Errorf{
			Type: errs.ErrInvalidInput,
			Message: "No .json metadata files were found.",
			ReturnRaw: true,
		}
	}

	latestPath := jsonPaths[len(jsonPaths) - 1]

	filePath, errf := DownloadIcebergS3(ctx, client, *newBucket.Data.Name, latestPath)
	if errf != nil {
		return nil, errf
	}

	metadata, errf := iceutils.ReadIcebergJSON(filePath)
	if errf != nil {
		return nil, errf
	}

	cleanIceberg := iceutils.CleanIceberg(metadata)

	return cleanIceberg, nil
}

func HandleParquet(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) ([]*dto.ParquetClean, *errs.Errorf) {

	resp, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: newBucket.Data.Name,
	})
	if err != nil {
        return nil, &errs.Errorf{
			Type: errs.ErrServiceUnavailable,
			Message: "Failed to list parquet objects : " + err.Error(),
		}
    }

	limit := configs.ParquetFilesLimit

	for _, obj := range resp.Contents {
		if limit <= 0 {
			break
		}
		if obj.Key != nil {
			key := *obj.Key
			if key[len(key) - 1] != '/' {
				newBucket.Parquet.AllFilePaths = append(newBucket.Parquet.AllFilePaths, key)
				limit--
			}
		}
	}

	var wg sync.WaitGroup
	cleanParquets := make([]*dto.ParquetClean, 0)

	for _, path := range newBucket.Parquet.AllFilePaths {
		wg.Add(1)

		go func(path string) {
			defer wg.Done()

			filePath, errf := DownloadSingleParquetS3(ctx, client, *newBucket.Data.Name, path)
			if errf != nil {
				// TODO: handle error, retry logic
				return
			}

			cleanParquet, errf := parqutils.ReadParquet(filePath)
			if errf != nil {
				fmt.Println(errf.Message)
				return
			}

			cleanParquet.URI = path

			cleanParquets = append(cleanParquets, cleanParquet)
		} (path)
	}

	wg.Wait()
	
	return cleanParquets, nil
}