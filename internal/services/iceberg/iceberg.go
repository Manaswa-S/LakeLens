package iceberg

import (
	"lakelens/internal/consts"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	sqlc "lakelens/internal/sqlc/generate"
	"lakelens/internal/stash"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type IcebergService struct {
	Queries     *sqlc.Queries
	RedisClient *redis.Client
	DB          *pgxpool.Pool

	Stash *stash.StashService
}

func NewIcebergService(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool, stash *stash.StashService) *IcebergService {
	return &IcebergService{
		Queries:     queries,
		RedisClient: redis,
		DB:          db,

		Stash: stash,
	}
}

func (s *IcebergService) fetchCache(ctx *gin.Context, userID int64, locid string) (*stash.CacheMetadata, *errs.Errorf) {

	locID, err := strconv.ParseInt(locid, 10, 64)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrInvalidInput,
			Message: "Failed to parse location id to int64 : " + err.Error(),
		}
	}

	locData, err := s.Queries.GetLocationData(ctx, locID)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get location data : " + err.Error(),
		}
	}

	if locData.UserID != userID {
		return nil, &errs.Errorf{
			Type:      errs.ErrUnauthorized,
			Message:   "Requested resource does not belong to you.",
			ReturnRaw: true,
		}
	}

	lakeData, err := s.Queries.GetLakeData(ctx, locData.LakeID)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get lake data : " + err.Error(),
		}
	}

	cache := new(stash.CacheMetadata)
	var exists bool

	switch lakeData.Ptype {
	case consts.AWSS3:
		cache, exists = s.Stash.GetBucketS3(locData.BucketName)

	default:
		// ?
	}

	if !exists {
		return nil, &errs.Errorf{
			Type:      errs.ErrNotFound,
			Message:   "Requested resource not found. Please rescan to fetch data.",
			ReturnRaw: true,
		}
	}

	if cache.Bucket.Data.TableType != consts.IcebergTable {
		return nil, &errs.Errorf{
			Type:      errs.ErrResourceLocked,
			Message:   "Requested resource is not of expected table type (iceberg).",
			ReturnRaw: true,
		}
	}

	return cache, nil

}

func (s *IcebergService) AllData(ctx *gin.Context, userID int64, locid string) (*dto.IsIceberg, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	return &cache.Bucket.Iceberg, nil
}

func (s *IcebergService) Metadata(ctx *gin.Context, userID int64, locid string) (*dto.IcebergMetadata, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	return cache.Bucket.Iceberg.Metadata, nil
}

func (s *IcebergService) Snapshot(ctx *gin.Context, userID int64, locid string) (*dto.IcebergSnapshot, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	return cache.Bucket.Iceberg.Snapshot, nil
}

func (s *IcebergService) Manifest(ctx *gin.Context, userID int64, locid string) ([]*dto.IcebergManifest, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	return cache.Bucket.Iceberg.Manifest, nil
}







