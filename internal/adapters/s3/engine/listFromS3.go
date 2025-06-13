package engine

import (
	"errors"
	"lakelens/internal/adapters/s3/pipeline"
	configs "lakelens/internal/config"
	"lakelens/internal/consts"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/gin-gonic/gin"
)

// ListBuckets lists all buckets for a given client.
func ListBuckets(ctx *gin.Context, client *s3.Client) ([]types.Bucket, error) {

	response, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	return response.Buckets, nil
}

// GetBucket determines if bucket exists and if access is allowed, returns some metadata too.
func GetBucket(ctx *gin.Context, client *s3.Client, bucketName string) (*types.Bucket, *errs.Errorf) {

	headBuc, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &bucketName,
	})
	if err != nil {
		var erraws smithy.APIError
		if errors.As(err, &erraws) {
			switch erraws.ErrorCode() {
			case "NotFound":
				return nil, &errs.Errorf{
					Type:      errs.ErrNotFound,
					Message:   "Bucket not found : " + bucketName,
					ReturnRaw: true,
				}
			case "Forbidden":
				return nil, &errs.Errorf{
					Type:      errs.ErrForbidden,
					Message:   "Bucket access is forbidden : " + bucketName + " : " + err.Error(),
					ReturnRaw: true,
				}
			}
		}
		return nil, &errs.Errorf{
			Type:    errs.ErrServiceUnavailable,
			Message: "Failed to get bucket metadata (head) : " + err.Error(),
		}
	}

	return &types.Bucket{
		Name:         &bucketName,
		CreationDate: nil,
		BucketRegion: headBuc.BucketRegion,
	}, nil
}

// GetLocationMetadata handles the metadata extraction of the given bucket.
func ScrapeLoc(ctx *gin.Context, client *s3.Client, bucket *types.Bucket) (*dto.NewBucket, *errs.Errorf) {

	newBucket := new(dto.NewBucket)
	newBucket.Data.Name = *bucket.Name
	newBucket.Data.StorageType = consts.AWSS3
	newBucket.Data.Region = bucket.BucketRegion
	newBucket.Data.CreationDate = bucket.CreationDate

	errf, defaultTo := DetermineTableTypeBFS(ctx, client, newBucket)
	if errf != nil {
		if defaultTo {
			newBucket.Errors = append(newBucket.Errors, errf)
		} else {
			return newBucket, errf
		}
	}

	switch {
	case newBucket.Iceberg.Present:
		{
			newBucket.Data.TableType = consts.IcebergTable
			_, errf := pipeline.HandleIceberg(ctx, client, newBucket)
			if errf != nil {
				return newBucket, errf
			}
		}
	case newBucket.Delta.Present:
		{
			newBucket.Data.TableType = consts.DeltaTable
			_, errf := pipeline.HandleDelta(ctx, client, newBucket)
			if errf != nil {
				return newBucket, errf
			}

			// TODO: coming soon !
		}
	case newBucket.Hudi.Present:
		{
			newBucket.Data.TableType = consts.HudiTable
			// TODO: coming soon !
		}
	default:
		{
			newBucket.Data.TableType = consts.ParquetFile
			newBucket.Parquet.Present = true
			_, errf := pipeline.HandleParquet(ctx, client, newBucket)

			if errf != nil {
				return newBucket, errf
			}
		}
	}

	return newBucket, nil
}

// DetermineTableType determines/detects the table type in a given bucket by recursively listing nested folders.
//
// This is the DFS based approach.
func DetermineTableType(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket, prefix string, depth int) *errs.Errorf {

	if depth <= 0 {
		return &errs.Errorf{
			Type:      errs.ErrNotFound,
			Message:   "Maximum allowed depth reached but no table type found.",
			ReturnRaw: true,
		}
	}

	rootFolders, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    &newBucket.Data.Name,
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrServiceUnavailable,
			Message: "Unable to list objects (folders) : " + err.Error(),
		}
	}

	for _, prefix := range rootFolders.CommonPrefixes {
		pre := *prefix.Prefix

		if strings.HasSuffix(pre, consts.IcebergMetaFolder) &&
			strings.HasSuffix(pre, consts.IcebergDataFolder) {
			newBucket.Iceberg.Present = true
			newBucket.Iceberg.URI = pre
		} else if strings.HasSuffix(pre, consts.DeltaLogFolder) {
			newBucket.Delta.Present = true
		} else if strings.HasSuffix(pre, consts.HudiMetaFolder) {
			newBucket.Hudi.Present = true
		}
	}

	if !(newBucket.Parquet.Present || newBucket.Delta.Present || newBucket.Iceberg.Present) {
		for _, prefix := range rootFolders.CommonPrefixes {
			pre := *prefix.Prefix
			errf := DetermineTableType(ctx, client, newBucket, pre, depth-1)
			if errf != nil {
				return errf
			}
		}
	}

	return nil
}

// DetermineTableTypeBFS determines/detects the table type in a given bucket.
//
// This is the BFS based approach. This should perform better for most cases.
func DetermineTableTypeBFS(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) (*errs.Errorf, bool) {

	queue := []string{""}
	maxDepth := configs.Extras.DetermineTableTypeMaxDepth
	subQueue := []string{}

	for maxDepth > 0 && len(queue) > 0 {
		subQueue = subQueue[:0]

		for _, prefix := range queue {

			rootFolders, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
				Bucket:    &newBucket.Data.Name,
				Prefix:    &prefix,
				Delimiter: aws.String("/"),
			})
			if err != nil {
				return &errs.Errorf{
					Type:    errs.ErrServiceUnavailable,
					Message: "Unable to list objects (folders) : " + err.Error(),
				}, false
			}

			for _, prefix := range rootFolders.CommonPrefixes {
				pre := *prefix.Prefix
				slashPre := "/" + pre

				switch {
				case strings.HasSuffix(slashPre, consts.IcebergMetaFolder):
					for _, prefix2 := range rootFolders.CommonPrefixes {
						pre2 := "/" + *prefix2.Prefix
						if strings.HasSuffix(pre2, consts.IcebergDataFolder) {
							newBucket.Iceberg.Present = true
							newBucket.Iceberg.URI = pre
							return nil, false
						}
					}

				case strings.HasSuffix(slashPre, consts.DeltaLogFolder):
					newBucket.Delta.Present = true
					newBucket.Delta.URI = pre
					return nil, false

				case strings.HasSuffix(slashPre, consts.HudiMetaFolder):
					newBucket.Hudi.Present = true
					return nil, false

				default:
				}

				subQueue = append(subQueue, pre)
			}
		}

		queue = subQueue
		maxDepth--
	}

	return &errs.Errorf{
		Type:      errs.ErrNotFound,
		Message:   "Maximum allowed depth reached but no table type found. Defaulting to extract few .parquet files if found.",
		ReturnRaw: true,
	}, true
}
