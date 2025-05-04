package stash

import (
	"fmt"
	"lakelens/internal/dto"
	utils "lakelens/internal/utils/common"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// TODO: auto cleanup and lastused checks are remaining
// also automatic creation if not exists, etc.
func (s *StashService) SetS3Client(ctx *gin.Context, lakeID int64) error {

	creds, err := s.Queries.GetCredentials(ctx, lakeID)
	if err != nil {
		return fmt.Errorf("failed to get credentials : %v", err)
	}

	lakeKey, err := utils.DecryptStringAESGSM(creds.Key)
	if err != nil {
		return fmt.Errorf("failed to decrypt key : %v", err)
	}

	s3client, err := s.NewS3Client(ctx, creds.KeyID, lakeKey, creds.Region, "")
	if err != nil {
		return fmt.Errorf("failed to create new s3 client : %v", err)
	}

	data := dto.S3ClientSave{
		LastUsed: time.Now(),
		S3Client: s3client,
	}

	s.cliMU.Lock()
	s.clients.S3[fmt.Sprintf("%d", lakeID)] = &data
	s.cliMU.Unlock()

	return nil
}

func (s *StashService) GetS3Client(ctx *gin.Context, lakeID int64) (*s3.Client, error) {

	client, ok := s.clients.S3[fmt.Sprintf("%d", lakeID)]
	if !ok || client == nil {
		err := s.SetS3Client(ctx, lakeID)
		if err != nil {
			return nil, fmt.Errorf("failed to load S3 client from cache for key : %d", lakeID)
		}
		return s.GetS3Client(ctx, lakeID)
	}

	return client.S3Client, nil
}

func (s *StashService) NewS3Client(ctx *gin.Context, keyId, key, region, sessionStr string) (*s3.Client, error) {
	// TODO: check and validate args
	config, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(keyId, key, sessionStr),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load a default config : %v", err)
	}

	client := s3.NewFromConfig(config, func(o *s3.Options) {
		o.ResponseChecksumValidation = aws.ResponseChecksumValidation(0)
	})

	return client, nil
}
