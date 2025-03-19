package s3utils

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"main.go/internal/dto"
	iceutils "main.go/internal/utils/iceberg"
	parqutils "main.go/internal/utils/parquet"
)

func HandleIceberg(ctx *gin.Context, client *s3.Client, bucData *dto.BucketData) ([]*dto.IcebergClean, error) {


	filePaths, errs := DownloadIcebergS3(ctx, client, bucData)
	if len(errs) != 0 {
		fmt.Println(errs)
		return nil, errs[0]
	}


	icebergs, errs := iceutils.ReadIcebergJSON(filePaths)
	if len(errs) != 0 {
		fmt.Println(errs)
		return nil, errs[0]
	}

	cleanIcebergs, err := iceutils.CleanIceberg(icebergs)
	if err != nil {
		return nil, err
	}

	return cleanIcebergs, nil
}



func HandleParquet(ctx *gin.Context, client *s3.Client, bucData *dto.BucketData) ([]*dto.ParquetClean, error) {

	var cleanParquet []*dto.ParquetClean

	filePaths, err := DownloadParquetS3(ctx, client, bucData.Name, bucData.Parquet.AllFilePaths)
	if err != nil {
		return nil, err
	}

	cleanParquet, err = parqutils.ReadParquet(filePaths)
	if err != nil {
		return nil, err
	}

	return cleanParquet, nil
}