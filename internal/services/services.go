package services

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	configs "main.go/internal/config"
	"main.go/internal/consts/errs"
	"main.go/internal/dto"
	sqlc "main.go/internal/sqlc/generate"
	utils "main.go/internal/utils/common"
	"main.go/internal/utils/s3utils"
)

type Services struct {
	Queries *sqlc.Queries
	RedisClient *redis.Client
	DB *pgxpool.Pool
}

func NewService(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool) *Services {
	return &Services{
		Queries: queries,
		RedisClient: redis,
		DB: db,
	}
} 

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Helper Functions

const (
	S3ClientRedisSubKey = "s3client"
)
var (
	SavedS3Clients sync.Map
)
// TODO: auto cleanup and lastused checks are remaining
// also automatic creation if not exists, etc.
func (s *Services) setS3Client(ctx *gin.Context, lakeID int64) error {

	creds, err := s.Queries.GetCredentials(ctx, lakeID)
	if err != nil {
		return fmt.Errorf("failed to get credentials : %v", err)
	}

	lakeKey, err := utils.DecryptStringAESGSM(creds.Key)
	if err != nil {
		return fmt.Errorf("failed to decrypt key : %v", err)
	}

	s3client, err := s.newS3Client(ctx, creds.KeyID, lakeKey, creds.Region, "")
	if err != nil {
		return fmt.Errorf("failed to create new s3 client : %v", err)
	}

	key := fmt.Sprintf("%d-%s", lakeID, S3ClientRedisSubKey)
	data := dto.ClientSave{
		LastUsed: time.Now(),
		S3Client: s3client,
	}

	SavedS3Clients.Store(key, data)

	return nil
}
func (s *Services) getS3Client(ctx *gin.Context, lakeID int64) (*s3.Client, error) {

	key := fmt.Sprintf("%d-%s", lakeID, S3ClientRedisSubKey)

	data, ok := SavedS3Clients.Load(key)
	if !ok || data == nil {
		err := s.setS3Client(ctx, lakeID)
		if err != nil {
			return nil, fmt.Errorf("failed to load S3 client from cache for key : %s", key)
		}
		return s.getS3Client(ctx, lakeID)
	}

	client, ok := data.(dto.ClientSave)
	if !ok {
		return nil, fmt.Errorf("failed to type assert client from cache for key : %s", key)
	}

	return client.S3Client, nil
}
func (s *Services) newS3Client(ctx *gin.Context, keyId, key, region, sessionStr string) (*s3.Client, error) {
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

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>



// NewUser adds a new user to the service. 
// Needs email and password, uses dto.NewUser.
// Email and Password conditions can be found in respective Validation functions.
func (s *Services) NewUser(ctx *gin.Context, data *dto.NewUser) *errs.Errorf {

	prblm, emailOk := utils.ValidateEmail(data.Email)
	if !emailOk {
		return &errs.Errorf{
			Type: errs.ErrInvalidInput,
			Message: "Invalid email : " + prblm,
			ReturnRaw: true,
		}
	}

	prblm, passOk := utils.ValidatePassword(data.Password)
	if !passOk {
		return &errs.Errorf{
			Type: errs.ErrInvalidInput,
			Message: "Invalid password : " + prblm,
			ReturnRaw: true,
		}
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(data.Password), configs.Internal.BcryptPasswordCost)
	if err != nil {
		return &errs.Errorf{
			Type: errs.ErrInternalServer,
			Message: "Failed to generate bcrypt hash from password : " + err.Error(),
		}
	}

	err = s.Queries.InsertNewUser(ctx, sqlc.InsertNewUserParams{
		Email: data.Email,
		Password: string(passHash),
	})
	if err != nil {
		errf := errs.Errorf{
			Type: errs.ErrDBQuery,
			Message: "Failed to insert new user into db : " + err.Error(),
		}

		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.Code == errs.PGErrUniqueViolation {
				errf.Type = errs.ErrConflict
				errf.Message = "User with this email already exists. Log in or use another email."
				errf.ReturnRaw = true
			}
		}
		return &errf
	}

	return nil
} 

// RegisterNewLake registers a new lake, retrieves available buckets.
func (s *Services) RegisterNewLake(ctx *gin.Context, userID int64, data *dto.NewLake) ([]types.Bucket, *errs.Errorf) {
	// TODO: data validation needs to be done here.

	client, err := s.newS3Client(ctx, data.AccessID, data.AccessKey, data.LakeRegion, "")
	if err != nil {
		return nil, &errs.Errorf{
			Type: errs.ErrInvalidCredentials,
			Message: "Invalid S3 lake credentials.",
			ReturnRaw: true,
		}
	}

	buckets := []types.Bucket{}
	var continueToken *string

	for {
		result, err := client.ListBuckets(ctx, &s3.ListBucketsInput{
			ContinuationToken: continueToken,
		})
		if err != nil {
			return nil, &errs.Errorf{
				Type: errs.ErrInternalServer,
				Message: "Failed to list buckets for client : " + err.Error(),
			}
		}

		buckets = append(buckets, result.Buckets...)

		if result.ContinuationToken == nil {
			break
		}

		continueToken = result.ContinuationToken
	}

	cipherKey, err := utils.EncryptStringAESGSM(data.AccessKey)
	if err != nil {
		return nil, &errs.Errorf{
			Type: errs.ErrInternalServer,
			Message: "Failed to encrypt key : " + err.Error(),
		}
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, &errs.Errorf{
			Type: errs.ErrInternalServer,
			Message: "Failed to generate db transaction : " + err.Error(),
		}
	}	
	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)
	
	lakeID, err := qtx.InsertNewLake(ctx, sqlc.InsertNewLakeParams{
		UserID: userID,
		Name: data.Name,
		Region: data.LakeRegion,
	})
	if err != nil {
		return nil, &errs.Errorf{
			Type: errs.ErrDBQuery,
			Message: "Failed to insert new lake in lakes : " + err.Error(),
		}
	}

	err = qtx.InsertNewCredentails(ctx, sqlc.InsertNewCredentailsParams{
		LakeID: lakeID,
		KeyID: data.AccessID,
		Key: cipherKey,
		Region: data.LakeRegion,
	})
	if err != nil {
		errf := errs.Errorf{
			Type: errs.ErrDBQuery,
			Message: "Failed to insert new credentials in db : " + err.Error(),
		}

		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.Code == errs.PGErrUniqueViolation {
				errf.Type = errs.ErrConflict
				errf.Message = "Lake with given Id already exists. Please edit it."
				errf.ReturnRaw = true
			}
		}

		return nil, &errf
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, &errs.Errorf{
			Type: errs.ErrDBConflict,
			Message: "Failed to commit register new lake db transaction : " + err.Error(),
		}
	}

	return buckets, nil
}





func (s *Services) GetLakeMetaData(ctx *gin.Context, userID int64, lakeid string) ([]*dto.BucketDataResponse, error) {

	lakeID, err := strconv.ParseInt(lakeid, 10, 64)
	if err != nil {
		return nil, err
	}

	client, err := s.getS3Client(ctx, lakeID)
	if err != nil {
		return nil, err
	}

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	
	buckets, err := s3utils.ListBuckets(ctx, client)
	if err != nil {
		return nil, err
	}

	response := make([]*dto.BucketDataResponse, 0)

	for _, bucket := range buckets {
		bucData, resp, err := s3utils.GetLocationMetadata(ctx, client, bucket)
		if err != nil {
			fmt.Println(err)
			continue
		}

		response = append(response, &dto.BucketDataResponse{
			Data: bucData,
			Metadata: resp,
		})
	}

	return response, nil
}

func (s *Services) GetLocMetaData(ctx *gin.Context, userID int64, locid string) (*dto.CompleteResponse,error) {

	// locID, err := strconv.ParseInt(locid, 10, 64)
	// if err != nil {
	// 	return nil, err
	// }

	// lakeID, err := s.Queries.GetLakeIDfromLocID(ctx, locID)
	// if err != nil {
	// 	return nil, err
	// }

	// client, err := s.getS3Client(ctx, lakeID)
	// if err != nil {
	// 	return nil, err
	// }

	// // response, err := s3utils.
	// // if err != nil {
	// // 	return nil, err
	// // }

	// return response, nil
	return nil, nil
}
