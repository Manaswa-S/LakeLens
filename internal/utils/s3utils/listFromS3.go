package s3utils

import (
	"fmt"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"main.go/internal/consts"
	"main.go/internal/dto"
	cutils "main.go/internal/utils/common"
)




func ListFromS3(ctx *gin.Context, client *s3.Client) (*dto.CompleteResponse, error) {

	bucList := dto.BucketList{}

	output, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	for _, obj := range output.Buckets {
		newBucket := dto.BucketData{
			Name: aws.ToString(obj.Name),
			StorageType: consts.AWSS3,
		}
		bucList[&newBucket] = &dto.AllFilesMp{}
	}

	for bucData, fileTree := range bucList {

		output, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucData.Name),
		})
		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, obj := range output.Contents {
			key := *obj.Key
			InsertIntoTree(*fileTree, key)
		}	
	}

	response := dto.CompleteResponse{}

	for bucData, fileTree := range bucList {

		fileTypeReturn := DetermineType(*fileTree)

		switch fileTypeReturn {
		case consts.ParquetFile:
			bucData.Parquet = true

			fmt.Println("Parquet Handling")
			bucData.LeafFilePaths = GetAllFilePaths(*fileTree, "")

			pc := HandleParquet(ctx, client, bucList)
			
			fmt.Println(pc)

			response.Parquet = append(response.Parquet, pc...)

		case consts.IcebergFile:
			bucData.Iceberg = true
		default:
			bucData.Unknown = true
		}

	}


	// for bucData, fileTree := range bucList {
	// 	fmt.Println(*bucData)
	// 	jsonData, _ := json.MarshalIndent(*fileTree, "", "  ")
	// 	fmt.Println(string(jsonData))
	// }

	return &response, nil 
}

func HandleParquet(ctx *gin.Context, client *s3.Client, bucList map[*dto.BucketData]*dto.AllFilesMp) ([]*dto.ParquetClean) {

	var cleanParquet []*dto.ParquetClean

	for bucData, _ := range bucList {
		switch bucData.StorageType {
		case consts.AWSS3:
			filePaths, err := DownloadParquetS3(ctx, client, bucData.Name, bucData.LeafFilePaths)
			if err != nil {
				fmt.Println(err)
				continue
			}

			cleanParquet, err = cutils.ReadParquetFileMetadata(filePaths)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println(cleanParquet)

		default:
			return nil
		}
	}
	return cleanParquet
}

func GetAllFilePaths(tree dto.AllFilesMp, currentPath string) []string {
	var filePaths []string

	for name, subTree := range tree {
		fullPath := path.Join(currentPath, name)

		if nested, ok := subTree.(dto.AllFilesMp); ok {
			filePaths = append(filePaths, GetAllFilePaths(nested, fullPath)...)
		} else {
			filePaths = append(filePaths, fullPath)
		}
	}

	return filePaths
}

func DetermineType(tree dto.AllFilesMp) (string) {

	// typeToPath := make(map[string]string, 0)

	// identify Iceberg files
	if _, exists := tree["data"]; exists {
		if _, exists := tree["metadata"]; exists {
			return consts.IcebergFile
		}
	}

	// jsonData, _ := json.MarshalIndent(tree, "", "  ")
	// fmt.Println(string(jsonData))

	for objName, subt := range tree {

		if subt, ok := subt.(dto.AllFilesMp); ok {
			result := DetermineType(subt)
			if result == consts.IcebergFile {
				return consts.IcebergFile
			} else if result == consts.ParquetFile {
				return consts.ParquetFile
			} 
		} else {
			if path.Ext(objName) == consts.ParquetFile {
				return consts.ParquetFile
			} else {
				return consts.UnknownFile
			}
		}
	}

	return consts.UnknownFile
}

func InsertIntoTree(tree dto.AllFilesMp, path string) {
	parts := strings.Split(path, "/")
	current := tree

	l := len(parts) - 1

	for i, part := range parts {

		if part == "" {
			continue
		}

		if i == l {
			current[part] = nil
		} else {
			if _, exists := current[part]; !exists {
				current[part] = dto.AllFilesMp{}
			}
			current = current[part].(dto.AllFilesMp)
		}
	}
}