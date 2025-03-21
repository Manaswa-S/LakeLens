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

func GetBucket(ctx *gin.Context, client *s3.Client, bucketName string) (*types.Bucket, error) {

	headBuc, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &bucketName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket metadata: %v", err)
	}

	bucket := types.Bucket{
		Name: &bucketName,
		CreationDate: nil,
		BucketRegion: headBuc.BucketRegion,
	}

	return &bucket, nil
}













func GetLocationMetadata(ctx *gin.Context, client *s3.Client, bucket types.Bucket) (*dto.BucketData, *dto.BucketMetadata, error) {

	fileTree := make(dto.FileTreeMap)
	newBucket := dto.BucketData{
		Name: bucket.Name,
		StorageType: consts.AWSS3,
		Region: bucket.BucketRegion,
		CreationDate: bucket.CreationDate,
		FileTree: &fileTree,
	}

	response, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(*bucket.Name),
	})
	if err != nil {
		return nil, nil, err
	}

	InsertIntoTree(fileTree, &response.Contents)

	newBucket.TableType = DetermineType(fileTree)

	respMetadata := new(dto.BucketMetadata)

	switch newBucket.TableType {
	case consts.ParquetFile:
		newBucket.Parquet.Present = true

		newBucket.Parquet.AllFilePaths = GetAllFilePaths(fileTree, "")

		parClean, err := HandleParquet(ctx, client, &newBucket)
		if err != nil {
			return nil, nil, err
		}

		respMetadata.Parquet = parClean
	case consts.IcebergFile:
		newBucket.Iceberg.Present = true

		newBucket.Iceberg.JSONFilePaths, newBucket.Iceberg.AvroFilePaths = GetIcebergFilePaths(fileTree, "")

		cleanIcebergs, err := HandleIceberg(ctx, client, &newBucket)
		if err != nil {
			fmt.Println(err)
		}

		respMetadata.Iceberg = cleanIcebergs
	default:
		newBucket.Unknown = true
		respMetadata.Unknown = nil
	}

	return &newBucket, respMetadata, nil
}

