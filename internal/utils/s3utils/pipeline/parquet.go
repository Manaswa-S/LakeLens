package pipeline

import (
	"fmt"
	configs "lakelens/internal/config"
	"lakelens/internal/consts"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	parqutils "lakelens/internal/utils/parquet"
	fetcher "lakelens/internal/utils/s3utils/engine/fetcher"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

func HandleParquet(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) ([]*dto.ParquetClean, bool, *errs.Errorf) {

	resp, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: newBucket.Data.Name,
	})
	if err != nil {
		return nil, false, &errs.Errorf{
			Type:    errs.ErrServiceUnavailable,
			Message: "Failed to list parquet objects : " + err.Error(),
		}
	}

	limit := configs.ParquetFilesLimit
	latestUpdate := time.Time{}

	for _, obj := range resp.Contents {
		if limit <= 0 {
			break
		}

		//
		// fmt.Printf("%s : %s : %s\n", *obj.Key, obj.LastModified, newBucket.Data.UpdatedAt)
		if obj.LastModified.After(latestUpdate) {
			latestUpdate = *obj.LastModified
		}
		//

		if obj.Key != nil {
			key := *obj.Key
			if key[len(key)-1] != '/' && strings.HasSuffix(key, consts.ParquetFileExt) {
				newBucket.Parquet.AllFilePaths = append(newBucket.Parquet.AllFilePaths, key)
				limit--
			}
		}
	}

	if !latestUpdate.After(newBucket.Data.UpdatedAt) && !latestUpdate.IsZero() {
		return nil, true, nil
	}
	newBucket.Data.UpdatedAt = latestUpdate

	var wg sync.WaitGroup
	cleanParquets := make([]*dto.ParquetClean, 0)

	for _, path := range newBucket.Parquet.AllFilePaths {
		wg.Add(1)

		go func(path string) {
			defer wg.Done()

			filePath, errf := fetcher.DownloadSingleParquetS3(ctx, client, *newBucket.Data.Name, path)
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
		}(path)
	}

	wg.Wait()

	return cleanParquets, false, nil
}
