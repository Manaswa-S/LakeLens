package s3utils

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"main.go/internal/consts"
	"main.go/internal/dto"
)


func ListBuckets(ctx *gin.Context, client *s3.Client) ([]types.Bucket, error) {

	response, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	return response.Buckets, nil
}

func GetLocationMetadata(ctx *gin.Context, client *s3.Client, bucket types.Bucket) (*dto.BucketData, *dto.CompleteResponse, error) {

	newBucket := dto.BucketData{
		Name: aws.ToString(bucket.Name),
		StorageType: consts.AWSS3,
	}
	fileTree := dto.AllFilesMp{}

	response, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(*bucket.Name),
	})
	if err != nil {
		return nil, nil, err
	}

	for _, obj := range response.Contents {
		InsertIntoTree(fileTree, *obj.Key)
	}	

	fileTypeReturn := DetermineType(fileTree)

	switch fileTypeReturn {
	case consts.ParquetFile:
		newBucket.Parquet.Present = true

		newBucket.Parquet.AllFilePaths = GetAllFilePaths(fileTree, "")

		parClean, err := HandleParquet(ctx, client, &newBucket)
		if err != nil {
			return nil, nil, err
		}
		
		return &newBucket, &dto.CompleteResponse{
			Parquet: parClean,
		}, nil
	case consts.IcebergFile:
		newBucket.Iceberg.Present = true

		newBucket.Iceberg.JSONFilePaths, newBucket.Iceberg.AvroFilePaths = GetIcebergFilePaths(fileTree, "")

		cleanIcebergs, err := HandleIceberg(ctx, client, &newBucket)
		if err != nil {
			fmt.Println(err)
		}

		return &newBucket, &dto.CompleteResponse{
			Iceberg: cleanIcebergs,
		}, nil
	default:
		newBucket.Unknown = true

		return &newBucket, &dto.CompleteResponse{
			Unknown: nil,
		}, nil
	}
}