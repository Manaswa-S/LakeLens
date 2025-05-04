package services

import (
	"errors"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	"lakelens/internal/services/iceberg"
	sqlc "lakelens/internal/sqlc/generate"
	"lakelens/internal/stash"
	utils "lakelens/internal/utils/common"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ManagerService struct {
	Queries     *sqlc.Queries
	RedisClient *redis.Client
	DB          *pgxpool.Pool

	Stash *stash.StashService

	// all individual table type services are injected in services/manager too for simpler interconnectivity.
	Iceberg *iceberg.IcebergService
}

func NewManagerService(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool, stash *stash.StashService, iceberg *iceberg.IcebergService) *ManagerService {
	return &ManagerService{
		Queries:     queries,
		RedisClient: redis,
		DB:          db,

		Stash:   stash,
		Iceberg: iceberg,
	}
}















func (s *ManagerService) processNewS3(ctx *gin.Context, data *dto.NewLakeS3) {

	client, err := s.newS3Client(ctx, data.AccessID, data.AccessKey, data.LakeRegion, "")
	if err != nil {
		return nil, &errs.Errorf{
			Type:      errs.ErrInvalidCredentials,
			Message:   "Invalid S3 lake credentials.",
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
				Type:    errs.ErrInternalServer,
				Message: "Failed to list buckets for client : " + err.Error(),
			}
		}

		buckets = append(buckets, result.Buckets...)

		if result.ContinuationToken == nil {
			break
		}

		continueToken = result.ContinuationToken
	}

}

func processNewAzure() {

}




// RegisterNewLake registers a new lake, retrieves available buckets.
func (s *ManagerService) RegisterNewLake(ctx *gin.Context, userID int64, data *dto.NewLake) ([]types.Bucket, *errs.Errorf) {
	// TODO: data validation needs to be done here.


	// < Trial

	s.processNewS3(ctx, data.S3)

	// >



	cipherKey, err := utils.EncryptStringAESGSM(data.AccessKey)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to encrypt key : " + err.Error(),
		}
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to generate db transaction : " + err.Error(),
		}
	}
	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)

	lakeID, err := qtx.InsertNewLake(ctx, sqlc.InsertNewLakeParams{
		UserID: userID,
		Name:   data.Name,
		Region: data.LakeRegion,
	})
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to insert new lake in lakes : " + err.Error(),
		}
	}

	err = qtx.InsertNewCredentails(ctx, sqlc.InsertNewCredentailsParams{
		LakeID: lakeID,
		KeyID:  data.AccessID,
		Key:    cipherKey,
		Region: data.LakeRegion,
	})
	if err != nil {
		errf := errs.Errorf{
			Type:    errs.ErrDBQuery,
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
			Type:    errs.ErrDBConflict,
			Message: "Failed to commit register new lake db transaction : " + err.Error(),
		}
	}

	return buckets, nil
}
