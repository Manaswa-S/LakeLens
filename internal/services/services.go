package services

import (
	"errors"
	"fmt"
	configs "lakelens/internal/config"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	sqlc "lakelens/internal/sqlc/generate"
	utils "lakelens/internal/utils/common"
	"lakelens/internal/utils/s3utils"
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


// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>



// NewUser adds a new user to the service. 
// Needs email and password, uses dto.NewUser.
// Email and Password conditions can be found in respective Validation functions.
func (s *Services) NewUser(ctx *gin.Context, data *dto.NewUser) *errs.Errorf {
	// TODO: need to confirm email, send email
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







func (s *Services) GetLakeData(ctx *gin.Context, userID int64, lakeid string) ([]*dto.NewBucket, []*errs.Errorf) {

	lakeID, err := strconv.ParseInt(lakeid, 10, 64)
	if err != nil {
		return nil, nil
	}

	client, err := s.getS3Client(ctx, lakeID)
	if err != nil {
		return nil, nil
	}

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	buckets, err := s3utils.ListBuckets(ctx, client)
	if err != nil {
		return nil, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	response := make([]*dto.NewBucket, 0)
	errorfs := make([]*errs.Errorf, 0)

	for _, bucket := range buckets {
		wg.Add(1)

		go func(bucket types.Bucket)  {
			defer wg.Done()
			newBucket, errf := s3utils.GetLocationMetadata(ctx, client, &bucket)
			if errf != nil {
				if errf.ReturnRaw {
					newBucket.Errors = append(newBucket.Errors, errf)
				} else {
					mu.Lock()
					errorfs = append(errorfs, errf)
					mu.Unlock()
				}
			}
			mu.Lock()
			response = append(response, newBucket)
			mu.Unlock()
		} (bucket)
	}

	wg.Wait()

	return response, errorfs
}

func (s *Services) GetLocData(ctx *gin.Context, userID int64, lakeid, locid string) (*dto.NewBucket, *errs.Errorf) {

	lakeID, err := strconv.ParseInt(lakeid, 10, 64)
	if err != nil {
		return nil, nil
	}

	locID, err := strconv.ParseInt(locid, 10, 64)
	if err != nil {
		return nil, nil
	}

	locData, err := s.Queries.GetLocationData(ctx, locID)
	if err != nil {
		return nil, nil
	}

	if lakeID != locData.LakeID {
		return nil, nil
	}

	client, err := s.getS3Client(ctx, locData.LakeID)
	if err != nil {
		return nil, nil
	}

	bucket, errf := s3utils.GetBucket(ctx, client, locData.BucketName)
	if errf != nil {
		return nil, errf
	}

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>..

	// bucData, exists := cacheutils.GetCacheS3Bucket(*bucket.Name)
	// if exists {
	// 	return bucData.Bucket, nil
	// }

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>..

	newBucket, errf := s3utils.GetLocationMetadata(ctx, client, bucket)
	if errf != nil {
		return nil, errf
	}

	return newBucket, nil
}





