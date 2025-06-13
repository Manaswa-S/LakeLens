package manager

import (
	"errors"
	"fmt"
	s3engine "lakelens/internal/adapters/s3/engine"
	"lakelens/internal/consts"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	"lakelens/internal/services/iceberg"
	sqlc "lakelens/internal/sqlc/generate"
	"lakelens/internal/stash"
	utils "lakelens/internal/utils/common"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

func (s *ManagerService) fetchCache(ctx *gin.Context, userID int64, locid string) (*stash.CacheMetadata, *errs.Errorf) {

	locID, err := strconv.ParseInt(locid, 10, 64)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrInvalidInput,
			Message: "Failed to parse location id to int64 : " + err.Error(),
		}
	}

	locData, err := s.Queries.GetLocationData(ctx, locID)
	if err != nil {
		if err.Error() == errs.PGErrNoRowsFound {
			return nil, &errs.Errorf{
				Type:      errs.ErrNotFound,
				Message:   "Requested resource not found, no such location registered.",
				ReturnRaw: true,
			}
		}
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

	return cache, nil
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Registering a new lake

func (s *ManagerService) processNewS3(ctx *gin.Context, userID int64, name string, data *dto.NewLakeS3) ([]types.Bucket, *errs.Errorf) {

	// TODO:
	// this gets weird here.
	// 1) check if we got access for all required policies
	// 2) fetch all buckets for them, irrespective of how much time/resources it takes.
	// 3) the user sends what all buckets he wants to be analyzed.
	// 4) if the number(buckets.requested_by_user) < limit > OK > do it.
	//    else tell the user that this is too heavy of a task and that he will need to lessen the number of locations.


	client, err := s.Stash.NewS3Client(ctx, data.AccessID, data.AccessKey, data.LakeRegion, "")
	if err != nil {
		return nil, &errs.Errorf{
			Type:      errs.ErrInvalidCredentials,
			Message:   "Invalid S3 lake credentials.",
			ReturnRaw: true,
		}
	}

	buckets := []types.Bucket{}
	var continueToken *string

	result, err := client.ListBuckets(ctx, &s3.ListBucketsInput{
		ContinuationToken: continueToken,
	})
	if err != nil {
		var serr *smithy.OperationError
		if errors.As(err, &serr) {
			return nil, &errs.Errorf{
				Type:      errs.ErrInvalidCredentials,
				Message:   "Invalid ID, Key or Region were provided. Please check your inputs.",
				ReturnRaw: true,
			}
		}
		return nil, &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to list buckets for client : " + err.Error(),
		}
	}

	buckets = append(buckets, result.Buckets...)

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
		Name:   name,
		Region: data.LakeRegion,
		Ptype:  consts.AWSS3,
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

func processNewAzure() {

}

// RegisterNewLake registers a new lake, retrieves available buckets.
func (s *ManagerService) RegisterNewLake(ctx *gin.Context, userID int64, data *dto.NewLake) ([]dto.NewLakeResp, *errs.Errorf) {
	// TODO: data validation needs to be done here.

	// TODO: its difficult to gauge requirements when provider types increase, lets solve it later.

	var lakeResp []dto.NewLakeResp

	// < Trial

	if data.S3 != nil {
		// process s3
		fmt.Println("Adding new lake")
		buckets, errf := s.processNewS3(ctx, userID, data.Name, data.S3)
		if errf != nil {
			return nil, errf
		}

		for _, buc := range buckets {
			lakeResp = append(lakeResp, dto.NewLakeResp{
				Name:         buc.Name,
				Region:       buc.BucketRegion,
				CreationDate: buc.CreationDate,
			})
		}

	} else if data.Azure != nil {
		// process azure
	} else {
		return nil, &errs.Errorf{
			Type:      errs.ErrBadForm,
			Message:   "All provider types are nil.",
			ReturnRaw: true,
		}
	}

	// >

	return lakeResp, nil
}



// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
func (s *ManagerService) AddLocations(ctx *gin.Context, userID int64, data *dto.AddLocsReq) (*errs.Errorf) {

	





	return nil
}
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// 1) sends the basic account details
// 2) sends the billing details
// 3) sends the lakes and locations details
// 4) sends the user settings and preferences

// TODO:
// 5) also need collaborators and other details

func (s *ManagerService) AccDetails(ctx *gin.Context, userID int64) (*dto.AccDetailsResp, *errs.Errorf) {

	userDetails, err := s.Queries.AccDetails(ctx, userID)
	if err != nil {
		if err.Error() == errs.PGErrNoRowsFound {
			return nil, &errs.Errorf{
				Type:      errs.ErrBadRequest,
				Message:   "Invalid user tokens. Please login again.",
				ReturnRaw: true,
			}
		}
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get account/user details : " + err.Error(),
		}
	}

	switch userDetails.AuthType {

	case consts.EPassAuth:
		{
			accDetails, err := s.Queries.GetEPDetails(ctx, userID)
			if err != nil {
				return nil, &errs.Errorf{
					Type:    errs.ErrDBQuery,
					Message: "Failed to get EP account details : " + err.Error(),
				}
			}

			return &dto.AccDetailsResp{
				Email:     userDetails.Email,
				CreatedAt: userDetails.CreatedAt.Time,
				Confirmed: userDetails.Confirmed,
				UUID:      userDetails.UserUuid.String(),

				Name:    accDetails.Name.String,
				Picture: accDetails.Picture.String,

				AuthType: consts.EPassAuth,
			}, nil
		}
	case consts.GoogleOAuth:
		{
			accDetails, err := s.Queries.GetGODetails(ctx, userID)
			if err != nil {
				return nil, &errs.Errorf{
					Type:    errs.ErrDBQuery,
					Message: "Failed to get GO account details : " + err.Error(),
				}
			}

			return &dto.AccDetailsResp{
				Email:     userDetails.Email,
				CreatedAt: userDetails.CreatedAt.Time,
				Confirmed: userDetails.Confirmed,
				UUID:      userDetails.UserUuid.String(),

				Name:    accDetails.Name.String,
				Picture: accDetails.Picture.String,

				AuthType: consts.GoogleOAuth,
			}, nil
		}

	default:
		return nil, &errs.Errorf{
			Type:    errs.ErrOutOfRange,
			Message: "Couldn't determine the auth type for account details. Can be a critical error.",
		}
	}
}

func (s *ManagerService) AccBilling(ctx *gin.Context, userID int64) (*dto.AccBillingResp, *errs.Errorf) {

	return &dto.AccBillingResp{
		Type:       "Free",
		Applicable: true,
	}, nil
}

func (s *ManagerService) AccProjects(ctx *gin.Context, userID int64) (*dto.AccProjectsResp, *errs.Errorf) {

	lakesList, err := s.Queries.GetLakesList(ctx, userID)
	if err != nil {
		if err.Error() != errs.PGErrNoRowsFound {
			return nil, &errs.Errorf{
				Type:    errs.ErrDBQuery,
				Message: "Failed to get lakes list from db : " + err.Error(),
			}
		}
	}

	locsList, err := s.Queries.GetLocsList(ctx, userID)
	if err != nil {
		if err.Error() != errs.PGErrNoRowsFound {
			return nil, &errs.Errorf{
				Type:    errs.ErrDBQuery,
				Message: "Failed to get locations list from db : " + err.Error(),
			}
		}
	}

	combos := make(map[int64]*dto.LocsForLake)
	for _, lake := range lakesList {
		a := dto.LakeResp{
			LakeID:    lake.LakeID,
			Name:      lake.Name,
			Ptype:     lake.Ptype,
			CreatedAt: lake.CreatedAt.Time,
			Region:    lake.Region,
		}
		combos[a.LakeID] = &dto.LocsForLake{
			Lake: a,
		}
	}

	for _, loc := range locsList {
		a := dto.LocResp{
			LocID:      loc.LocID,
			LakeID:     loc.LakeID,
			BucketName: loc.BucketName,
			CreatedAt:  loc.CreatedAt.Time,
		}
		combos[a.LakeID].Locs = append(combos[a.LakeID].Locs, a)
	}

	list := make([]*dto.LocsForLake, 0)
	for _, b := range combos {
		list = append(list, b)
	}

	return &dto.AccProjectsResp{
		LocsForLake: list,
	}, nil
}

func (s *ManagerService) AccSettings(ctx *gin.Context, userID int64) (*dto.AccSettingsResp, *errs.Errorf) {

	settings, err := s.Queries.GetSettings(ctx, userID)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get settings for user : " + err.Error(),
		}
	}

	return &dto.AccSettingsResp{
		AdvancedMetaData:    settings.Advmeta,
		CompactView:         settings.Cmptview,
		AutoRefreshInterval: int32(settings.Rfrshint),
		Notifications:       settings.Notif,

		Theme:        settings.Theme,
		FontSize:     int32(settings.Fontsz),
		ShowToolTips: settings.Tooltps,

		KeyboardShortcuts: settings.Shortcuts,
	}, nil
}

func (s *ManagerService) AccSettingsUpdate(ctx *gin.Context, data *dto.AccSettingsResp, userID int64) *errs.Errorf {

	if data.AutoRefreshInterval > 60 || data.AutoRefreshInterval <= 15 {
		return &errs.Errorf{
			Type:      errs.ErrInvalidInput,
			Message:   "The auto refresh interval should be between 15 and 60 only.",
			ReturnRaw: true,
		}
	}

	if data.FontSize > 35 || data.AutoRefreshInterval <= 5 {
		return &errs.Errorf{
			Type:      errs.ErrInvalidInput,
			Message:   "The font size should be between 5 and 35 only.",
			ReturnRaw: true,
		}
	}

	err := s.Queries.UpdateSettings(ctx, sqlc.UpdateSettingsParams{
		UserID:      userID,
		LastUpdated: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Advmeta:     data.AdvancedMetaData,
		Cmptview:    data.CompactView,
		Rfrshint:    int16(data.AutoRefreshInterval),
		Notif:       data.Notifications,
		Theme:       data.Theme,
		Fontsz:      int16(data.FontSize),
		Tooltps:     data.ShowToolTips,
		Shortcuts:   data.KeyboardShortcuts,
	})
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get settings for user : " + err.Error(),
		}
	}
	return nil
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Analyzing a lake

type CloudClient interface {
	ProcessLake(ctx *gin.Context) ([]*dto.NewBucket, []*errs.Errorf)
	ProcessLoc(ctx *gin.Context, bucName string) (*dto.NewBucket, *errs.Errorf)
}

type S3Client struct {
	client *s3.Client
}

func (c *S3Client) ProcessLake(ctx *gin.Context) ([]*dto.NewBucket, []*errs.Errorf) {

	buckets, err := s3engine.ListBuckets(ctx, c.client)
	if err != nil {
		return nil, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	response := make([]*dto.NewBucket, 0)
	errorfs := make([]*errs.Errorf, 0)

	for _, bucket := range buckets {
		wg.Add(1)

		go func(bucket types.Bucket) {
			defer wg.Done()
			newBucket, errf := s3engine.ScrapeLoc(ctx, c.client, &bucket)
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
		}(bucket)
	}
	wg.Wait()

	return response, errorfs
}
func (c *S3Client) ProcessLoc(ctx *gin.Context, bucName string) (*dto.NewBucket, *errs.Errorf) {

	bucket, errf := s3engine.GetBucket(ctx, c.client, bucName)
	if errf != nil {
		return nil, errf
	}

	newBucket, errf := s3engine.ScrapeLoc(ctx, c.client, bucket)
	if errf != nil {
		return nil, errf
	}

	return newBucket, nil
}

type AzureClient struct {
	// init
}

func (c *AzureClient) ProcessLake(ctx *gin.Context) ([]*dto.NewBucket, []*errs.Errorf) {

	return nil, nil
}
func (c *AzureClient) ProcessLoc(ctx *gin.Context, bucName string) (*dto.NewBucket, *errs.Errorf) {

	return nil, nil
}

func (s *ManagerService) handleLakeAnalysis(ctx *gin.Context, c CloudClient) ([]*dto.NewBucket, []*errs.Errorf) {
	return c.ProcessLake(ctx)
}
func (s *ManagerService) handleLocAnalysis(ctx *gin.Context, bucName string, c CloudClient) (*dto.NewBucket, *errs.Errorf) {
	return c.ProcessLoc(ctx, bucName)
}

func (s *ManagerService) AnalyzeLake(ctx *gin.Context, userID int64, lakeid string) ([]*dto.BucketData, []*errs.Errorf) {

	lakeID, err := strconv.ParseInt(lakeid, 10, 64)
	if err != nil {
		return nil, []*errs.Errorf{
			&errs.Errorf{
				Type:    errs.ErrBadForm,
				Message: "Failed to parse lake id as int64 : " + err.Error(),
			},
		}
	}

	//

	lakeData, err := s.Queries.GetLakeData(ctx, lakeID)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	var client CloudClient

	switch lakeData.Ptype {
	case consts.AWSS3:
		s3Client, err := s.Stash.GetS3Client(ctx, lakeID)
		if err != nil {
			return nil, nil
		}
		client = &S3Client{
			client: s3Client,
		}
	case consts.Azure:
		// init
	case consts.MinIO:
		// init
	default:
		// init
		fmt.Println("no provider match")
		return nil, nil
	}

	//

	buckets, errfs := s.handleLakeAnalysis(ctx, client)
	if len(errfs) != 0 {
		return nil, errfs
	}

	bucsData := make([]*dto.BucketData, 0)
	for _, bucket := range buckets {
		s.Stash.SetBucket(bucket)
		bucsData = append(bucsData, &bucket.Data)
	}

	return bucsData, nil
}

func (s *ManagerService) AnalyzeLoc(ctx *gin.Context, userID int64, lakeid, locid string) (*dto.BucketData, *errs.Errorf) {

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

	lakeData, err := s.Queries.GetLakeData(ctx, lakeID)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	var client CloudClient

	switch lakeData.Ptype {
	case consts.AWSS3:
		s3Client, err := s.Stash.GetS3Client(ctx, lakeID)
		if err != nil {
			return nil, nil
		}
		client = &S3Client{
			client: s3Client,
		}
	case consts.Azure:
		// init
	case consts.MinIO:
		// init
	default:
		// init
		fmt.Println("no provider match")
		return nil, nil
	}

	//

	bucket, errf := s.handleLocAnalysis(ctx, locData.BucketName, client)
	if errf != nil {
		return nil, errf
	}

	s.Stash.SetBucket(bucket)

	return &bucket.Data, nil
}

func (s *ManagerService) FetchLocation(ctx *gin.Context, userID int64, locid string) (*dto.NewBucket, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	return cache.Bucket, nil
}
