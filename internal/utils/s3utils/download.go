package s3utils

import (
	"encoding/binary"
	"fmt"
	"io"
	configs "lakelens/internal/config"
	"lakelens/internal/consts/errs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// DownloadParquetS3 downloads all parquet files given in leafFilePaths concurrently.
//
// By default fetches the last 48 KB.
func DownloadParquetS3(ctx *gin.Context, client *s3.Client, bucketName string, leafFilePaths []string) ([]string, error) {

	dwnldPaths := make([]string, 0)

	// TODO: replace this to be dynamic
	rangeHeader := "bytes=-48000"

	var wg sync.WaitGroup

	for i, path := range leafFilePaths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			obj, err := client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: &bucketName,
				Key: &path,
				Range: &rangeHeader,
			})
			if err != nil {
				dwnldPaths = append(dwnldPaths, "")
				return
			}
			defer obj.Body.Close()

			filePath := fmt.Sprintf("%d", i)

			outfile, err := os.Create(filePath)
			if err != nil {
				dwnldPaths = append(dwnldPaths, "")
				return
			}

			_, err = outfile.ReadFrom(obj.Body)
			if err != nil {
				dwnldPaths = append(dwnldPaths, "")
				return
			}			

			dwnldPaths = append(dwnldPaths, filePath)
		} (i, path)
	}
	wg.Wait()
	return dwnldPaths, nil
}

// DownloadSingleParquetS3 downloads  a parquet file from given URI.
//
// It first fetches last 8 bytes to determine the footer length and 
// then downloads the footer accordingly. 
// It increases latency and API calls but zeros the chances of incomplete footer fetching.
// This shouldn't be an issue as a limit is already in place to limit the number of downloads per request.
func DownloadSingleParquetS3(ctx *gin.Context, client *s3.Client, bucketName, uri string) (string, *errs.Errorf) {

	rangeHeader := "bytes=-8"

	objFooter, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key: &uri,
		Range: &rangeHeader,
	})
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrServiceUnavailable,
			Message: "Failed to get parquet file :  " + err.Error(),
		}
	}

	objBody := make([]byte, 8)
	_, err = objFooter.Body.Read(objBody)
	if err != nil {
		if err.Error() == "EOF" {
			// TODO:
		} else {
			fmt.Printf("%s : %v", "error : ", err)
		}
	}
	footerLen := int64(binary.LittleEndian.Uint32(objBody[:4]))
	objFooter.Body.Close()

	rangeHeader = fmt.Sprintf("bytes=-%d", footerLen + 8) 
	completeObj, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key: &uri,
		Range: &rangeHeader,
	})
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrServiceUnavailable,
			Message: "Failed to get parquet file :  " + err.Error(),
		}
	}
	defer completeObj.Body.Close()

	dirPath := filepath.Join(configs.ParquetDownloadS3Path, bucketName)
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrStorageFailed,
			Message: "Failed to create parquet download directory : " + err.Error(),
		}
	}

	pathSplits := strings.Split(uri, "/")
	filePath := filepath.Join(dirPath, pathSplits[len(pathSplits)-1])
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrStorageFailed,
			Message: "Failed to create parquet file : " + err.Error(),
		}
	}

	_, err = io.Copy(outFile, completeObj.Body)
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrStorageFailed,
			Message: "Failed to copy/write to outFile : " + err.Error(),
		}
	}

	return filePath, nil
}

// DownloadIcebergS3 downloads iceberg metadata files from S3
// , stores them locally and returns their local filepaths.
func DownIcebergS3(ctx *gin.Context, client *s3.Client, bucketName, key string) (string, *errs.Errorf) {

	basePath := configs.IcebergDownloadS3Path
	
	obj, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key: &key,
	})
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrServiceUnavailable,
			Message: "Failed to get object : " + err.Error(),
		}
	}
	defer obj.Body.Close()

	dirPath := filepath.Join(basePath, bucketName)
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrStorageFailed,
			Message: "Failed to make directory : " + err.Error(),
		}
	}

	pathSplits := strings.Split(key, "/")
	filePath := filepath.Join(dirPath, pathSplits[len(pathSplits)-1])

	outFile, err := os.Create(filePath)
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrStorageFailed,
			Message: "Failed to create file : " + err.Error(),
		}
	}

	_, err = io.Copy(outFile, obj.Body)
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrStorageFailed,
			Message: "Failed to copy/write to outFile : " + err.Error(),
		}
	}

	return filePath, nil
}


// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>


func DownAvroS3(ctx *gin.Context, client *s3.Client, bucketName, key string) (string, *errs.Errorf) {

	basePath := configs.IcebergDownloadS3Path
	
	obj, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key: &key,
	})
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrServiceUnavailable,
			Message: "Failed to get object : " + err.Error(),
		}
	}
	defer obj.Body.Close()

	dirPath := filepath.Join(basePath, bucketName)
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrStorageFailed,
			Message: "Failed to make directory : " + err.Error(),
		}
	}

	pathSplits := strings.Split(key, "/")
	filePath := filepath.Join(dirPath, pathSplits[len(pathSplits)-1])

	outFile, err := os.Create(filePath)
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrStorageFailed,
			Message: "Failed to create file : " + err.Error(),
		}
	}

	_, err = io.Copy(outFile, obj.Body)
	if err != nil {
		return "", &errs.Errorf{
			Type: errs.ErrStorageFailed,
			Message: "Failed to copy/write to outFile : " + err.Error(),
		}
	}

	return filePath, nil
}


