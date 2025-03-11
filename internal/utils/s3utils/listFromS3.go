package s3utils

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	fileList = make([]string, 0)

	parquetList = make([]string, 0)

)

var (
	bucketList = make([]*bucket, 0)
)

type file struct {
	path string
	url string
}

type bucket struct {
	name string
	storagetype string
	parquet []*file
	// iceberg []*file
}















func ListFromS3() {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := s3.NewFromConfig(cfg)
	
	output, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, obj := range output.Buckets {
		fmt.Printf("%s, : ", aws.ToString(obj.Name))
		newBucket := bucket{
			name: aws.ToString(obj.Name),
			storagetype: "s3",
			parquet: []*file{},
		}
		bucketList = append(bucketList, &newBucket)
	}

	for _, bName := range fileList {
		output, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bName),
		})
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Printf("%d / %d \n", *output.KeyCount, *output.MaxKeys)
		
		for _, obj := range output.Contents {
			fmt.Printf("%s\n", aws.ToString(obj.Key))
		}	

	}

	

}

const (
	ParquetFile = ".parquet"
)


func getFileExtension(key string) (error) {
	base := path.Base(key) 
	ext := strings.ToLower(path.Ext(key))

	switch ext {
	case ParquetFile:
		parquetList = append(parquetList, )
	}





	return nil
}