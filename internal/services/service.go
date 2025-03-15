package services

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"main.go/internal/dto"
	sqlc "main.go/internal/sqlc/generate"
	"main.go/internal/utils/s3utils"
)

type Services struct {
	Queries *sqlc.Queries
	RedisClient *redis.Client
}

func NewService(queries *sqlc.Queries, redis *redis.Client) *Services {
	return &Services{
		Queries: queries,
		RedisClient: redis,
	}
} 

func (s *Services) RegisterNewLake(ctx *gin.Context, data *dto.NewLake) error {
	// TODO:
	// encryptedKey, err := utils.EncryptKey(data.AccessKey)
	// if err != nil {
	// 	return err
	// }

	err := s.Queries.InsertNewCredentails(ctx, sqlc.InsertNewCredentailsParams{
		KeyID: data.AccessID,
		Key: data.AccessKey,
		Region: data.LakeRegion,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Services) GetMetaData(ctx *gin.Context, userID int64, lakeid string) (*dto.CompleteResponse,error) {

	lakeID, err := strconv.ParseInt(lakeid, 10, 64)
	if err != nil {
		return nil, err
	}


	// TODO: get lake creds from db, decrypt, make a client using ctx

	creds, err := s.Queries.GetCredentials(ctx, lakeID)
	if err != nil {
		return nil, err
	}

	// TODO: this is not how we do this 
	// everytime a user comes online we create a client and store it, reuse it
	config, err := config.LoadDefaultConfig(ctx,
	config.WithRegion(creds.Region),
	config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(creds.KeyID, creds.Key, "")))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(config, func(o *s3.Options) {
		o.ResponseChecksumValidation = aws.ResponseChecksumValidation(0)
	})

	response, err := s3utils.ListFromS3(ctx, client)
	if err != nil {
		return nil, err
	}

	return response, nil
}
