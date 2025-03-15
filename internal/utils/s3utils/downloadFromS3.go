package s3utils

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
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
 


