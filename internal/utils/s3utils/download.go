package s3utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	configs "main.go/internal/config"
	"main.go/internal/dto"
)


func DownloadParquetS3(ctx *gin.Context, client *s3.Client, bucName string, leafFilePaths []string) ([]string, error) {

	dwnldPaths := make([]string, 0)

	// header to get only last 8KB of parquet files
	// TODO: replace this to be dynamic
	rangeHeader := "bytes=-46384"

	for i, path := range leafFilePaths {

		obj, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: &bucName,
			Key: &path,
			Range: &rangeHeader,
		})
		if err != nil {
			continue
		}
		defer obj.Body.Close()

		filePath := fmt.Sprintf("%d", i)

		fmt.Printf("%s : %d\n", path, i)

		outfile, err := os.Create(filePath)
		if err != nil {
			continue
		}

		_, err = outfile.ReadFrom(obj.Body)
		if err != nil {
			continue
		}			

		dwnldPaths = append(dwnldPaths, filePath)
	}

	return dwnldPaths, nil
}



// DownloadIcebergS3 downloads iceberg metadata files from S3
// , stores them locally and returns their local filepaths.
func DownloadIcebergS3(ctx *gin.Context, client *s3.Client, bucData *dto.BucketData) (filePaths []string, errs []error) {

	basePath := configs.IcebergDownloadS3Path

	for _, path := range bucData.Iceberg.JSONFilePaths {
		obj, err := client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &bucData.Name,
			Key: &path,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get object from s3 : %v", err))
			continue
		}
		defer obj.Body.Close()

		dirPath := filepath.Join(basePath, bucData.Name)
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to make directory : %v", err))
			continue
		}

		pathSplits := strings.Split(path, "/")
		filePath := filepath.Join(dirPath, pathSplits[len(pathSplits)-1])

		outFile, err := os.Create(filePath)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create file : %v", err))
			continue
		}

		_, err = io.Copy(outFile, obj.Body)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to copy object : %v", err))
			continue
		}

		filePaths = append(filePaths, filePath)
	}

	return
}


