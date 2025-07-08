package pipeline

import (
	"lakelens/internal/adapters/s3/engine/fetcher"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	deltautils "lakelens/internal/utils/delta"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

func HandleDelta(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) (bool, *errs.Errorf) {

	resp, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &newBucket.Data.Name,
		Prefix: &newBucket.Iceberg.URI,
	})
	if err != nil {
		return false, &errs.Errorf{
			Type:    errs.ErrServiceUnavailable,
			Message: "Failed to list objects : " + err.Error(),
		}
	}

	for _, obj := range resp.Contents {

		key := *obj.Key
		if strings.HasSuffix(key, ".json") {
			newBucket.Delta.LogFPaths = append(newBucket.Delta.LogFPaths, key)
		} else if strings.HasSuffix(key, ".crc") {
			newBucket.Delta.CRCFPaths = append(newBucket.Delta.CRCFPaths, key)
		}
	}

	errf := logOps(ctx, client, newBucket)
	if errf != nil {
		return false, errf
	}

	return false, nil
}

func logOps(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) *errs.Errorf {

	slices.Sort(newBucket.Delta.LogFPaths)
	deltaMetaFilesLimit := 3

	for i := len(newBucket.Delta.LogFPaths) - 1; i >= 0; i-- {

		fPath, errf := fetcher.FetchNdSave(ctx, client, newBucket.Data.Name, newBucket.Delta.LogFPaths[i], "")
		if errf != nil {
			return errf
		}

		log, errf := deltautils.ReadMetadata(fPath)
		if errf != nil {
			return errf
		}

		if log.Metadata.SchemaString != "" {

			newBucket.Delta.Log = append(newBucket.Delta.Log, log)
			// newBucket.Delta.Log = append([]*formats.DeltaMetadata{meta}, newBucket.Delta.Log...)

			deltaMetaFilesLimit--
			if deltaMetaFilesLimit == 0 {
				break
			}
		}
	}

	return nil
}
