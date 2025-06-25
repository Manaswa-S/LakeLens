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
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/transport/http"
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

func (s *ManagerService) processNewS3(ctx *gin.Context, userID int64, name string, data *dto.NewLakeS3) (*dto.NewLakeResp, *errs.Errorf) {

	// TODO:
	// this gets weird here.
	// 1) check if we got access for all required policies
	// 2) fetch all buckets for them, irrespective of how much time/resources it takes.
	// 3) the user sends what all buckets he wants to be analyzed.
	// 4) if the number(buckets.requested_by_user) < limit > OK > do it.
	//    else tell the user that this is too heavy of a task and that he will need to lessen the number of locations.

	// TODO: this should be gets3client only, the stash internally decides on new/cached ,etc.
	client, err := s.Stash.NewS3Client(ctx, data.AccessID, data.AccessKey, data.LakeRegion, "")
	if err != nil {
		return nil, &errs.Errorf{
			Type:      errs.ErrInvalidCredentials,
			Message:   "Invalid S3 lake credentials.",
			ReturnRaw: true,
		}
	}

	buckets := make([]dto.Locations, 0)
	var continueToken *string

	for {
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

		for _, buc := range result.Buckets {
			buckets = append(buckets, dto.Locations{
				Name:         buc.Name,
				CreationDate: buc.CreationDate,
				Region:       buc.BucketRegion,
			})
		}
		if result.ContinuationToken == nil {
			break
		}
		continueToken = result.ContinuationToken
	}

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

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

	return &dto.NewLakeResp{
		LakeID:    lakeID,
		Locations: buckets,
	}, nil
}

// RegisterNewLake registers a new lake, retrieves available buckets.
func (s *ManagerService) RegisterNewLake(ctx *gin.Context, userID int64, data *dto.NewLake) (*dto.NewLakeResp, *errs.Errorf) {
	// TODO: data validation needs to be done here.

	// TODO: its difficult to gauge requirements when provider types increase, lets solve it later.

	var lakeResp *dto.NewLakeResp
	var errf *errs.Errorf
	// < Trial

	if data.S3 != nil {
		// process s3
		fmt.Println("Adding new lake")
		lakeResp, errf = s.processNewS3(ctx, userID, data.Name, data.S3)
		if errf != nil {
			return nil, errf
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
	GetLocs(ctx *gin.Context) ([]*dto.Locations, *errs.Errorf)
	AddLocs(ctx *gin.Context, locNames []string) (*dto.AddLocsResp, *errs.Errorf)
	ProcessLake(ctx *gin.Context) ([]*dto.NewBucket, []*errs.Errorf)
	ProcessLoc(ctx *gin.Context, bucName string) (*dto.NewBucket, *errs.Errorf)
}

type S3Client struct {
	client *s3.Client
}

func (c *S3Client) GetLocs(ctx *gin.Context) ([]*dto.Locations, *errs.Errorf) {

	bucs, err := s3engine.ListBuckets(ctx, c.client)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to list buckets from s3 : " + err.Error(),
		}
	}

	locs := make([]*dto.Locations, 0)
	for _, buc := range bucs {
		locs = append(locs, &dto.Locations{
			Name:         buc.Name,
			CreationDate: buc.CreationDate,
			Region:       buc.BucketRegion,
		})
	}

	return locs, nil
}
func (c *S3Client) AddLocs(ctx *gin.Context, locNames []string) (*dto.AddLocsResp, *errs.Errorf) {

	resp := new(dto.AddLocsResp)

	// TODO: can and mp should do this in parallel

	for _, locName := range locNames {

		_, err := c.client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: &locName,
		})
		if err != nil {
			var serr *smithy.OperationError
			if errors.As(err, &serr) {
				var httperr *http.ResponseError
				if errors.As(serr.Err, &httperr) && (httperr.HTTPStatusCode() >= 400 && httperr.HTTPStatusCode() < 500) {
					resp.Failed = append(resp.Failed, locName)
					continue
				}
			}
			return nil, &errs.Errorf{
				Type:    errs.ErrDependencyFailed,
				Message: "Failed to get head bucket to add location : " + err.Error(),
			}
		}

		resp.Added = append(resp.Added, locName)
	}

	return resp, nil
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

func (s *ManagerService) handleGetLocs(ctx *gin.Context, c CloudClient) ([]*dto.Locations, *errs.Errorf) {
	return c.GetLocs(ctx)
}
func (s *ManagerService) handleAddLocs(ctx *gin.Context, locNames []string, c CloudClient) (*dto.AddLocsResp, *errs.Errorf) {
	return c.AddLocs(ctx, locNames)
}
func (s *ManagerService) handleLakeAnalysis(ctx *gin.Context, c CloudClient) ([]*dto.NewBucket, []*errs.Errorf) {
	return c.ProcessLake(ctx)
}
func (s *ManagerService) handleLocAnalysis(ctx *gin.Context, bucName string, c CloudClient) (*dto.NewBucket, *errs.Errorf) {
	return c.ProcessLoc(ctx, bucName)
}

func (s *ManagerService) GetLocations(ctx *gin.Context, userID int64, lakeid string) ([]*dto.Locations, *errs.Errorf) {

	lakeID, err := strconv.ParseInt(lakeid, 10, 64)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrBadForm,
			Message: "Failed to parse lake id as int64 : " + err.Error(),
		}
	}

	lakeData, err := s.Queries.GetLakeData(ctx, lakeID)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get lake data : " + err.Error(),
		}
	}

	if lakeData.UserID != userID {
		return nil, &errs.Errorf{
			Type:      errs.ErrUnauthorized,
			Message:   "This lake does not belong to the requesting user.",
			ReturnRaw: true,
		}
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
	default:
		// init
		fmt.Println("no provider match")
		return nil, &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "The lake ptype did not match to any.",
		}
	}

	buckets, errf := s.handleGetLocs(ctx, client)
	if errf != nil {
		return nil, errf
	}

	regLocs, err := s.Queries.GetLocsListForLake(ctx, sqlc.GetLocsListForLakeParams{
		UserID: userID,
		LakeID: lakeID,
	})
	if err != nil {
		if err.Error() != errs.PGErrNoRowsFound {
			return nil, nil
		}
	}

	// TODO: this is dumb.
	for _, buc := range buckets {
		for _, reg := range regLocs {
			if *buc.Name == reg.BucketName {
				buc.Registered = true
				break
			}
		}
	}

	return buckets, nil
}

func (s *ManagerService) AddLocations(ctx *gin.Context, userID int64, data *dto.AddLocsReq) (*dto.AddLocsResp, *errs.Errorf) {

	// TODO: you need the limits here.

	if data.LakeID == 0 || len(data.LocNames) == 0 {
		return nil, &errs.Errorf{
			Type:      errs.ErrBadRequest,
			Message:   "Lake name and location names are required.",
			ReturnRaw: true,
		}
	}

	lakeData, err := s.Queries.GetLakeData(ctx, data.LakeID)
	if err != nil {
		if err.Error() == errs.PGErrNoRowsFound {
			return nil, &errs.Errorf{
				Type:      errs.ErrBadRequest,
				Message:   "Invalid inputs. Please check and try again.",
				ReturnRaw: true,
			}
		}
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to list lakes for user : " + err.Error(),
		}
	}

	locsList, err := s.Queries.GetLocsListForLake(ctx, sqlc.GetLocsListForLakeParams{
		UserID: userID,
		LakeID: data.LakeID,
	})
	if err != nil {
		if err.Error() != errs.PGErrNoRowsFound {
			return nil, &errs.Errorf{
				Type:    errs.ErrDBQuery,
				Message: "Failed to get list of locations for lake : " + err.Error(),
			}
		}
	}

	toAdd := make([]string, 0)
	// TODO: this is dumb.
	for _, req := range data.LocNames {
		found := false
		for _, loc := range locsList {
			if req == loc.BucketName {
				found = true
				break
			}
		}
		if !found {
			toAdd = append(toAdd, req)
		}
	}

	var client CloudClient

	switch lakeData.Ptype {
	case consts.AWSS3:
		s3Client, err := s.Stash.GetS3Client(ctx, data.LakeID)
		if err != nil {
			return nil, &errs.Errorf{
				Type:    errs.ErrInternalServer,
				Message: "Failed to get S3 client to add locations.",
			}
		}
		client = &S3Client{
			client: s3Client,
		}
	default:
		return nil, &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "No provider type matched for add locations.",
		}
	}

	resp, errf := s.handleAddLocs(ctx, toAdd, client)
	if errf != nil {
		return nil, errf
	}

	for _, added := range resp.Added {
		err = s.Queries.InsertNewLocation(ctx, sqlc.InsertNewLocationParams{
			LakeID:     data.LakeID,
			BucketName: added,
			UserID:     userID,
		})
		if err != nil {
			return nil, &errs.Errorf{
				Type:    errs.ErrDBQuery,
				Message: "Failed to insert a new location : " + err.Error(),
			}
		}
	}

	return resp, nil
}

func (s *ManagerService) DeleteLake(ctx *gin.Context, userID int64, lakeid string) *errs.Errorf {

	lakeID, err := strconv.ParseInt(lakeid, 10, 64)
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrBadForm,
			Message: "Failed to parse lake id as int64 : " + err.Error(),
		}
	}

	_, err = s.Queries.GetLakeDataForUserID(ctx, sqlc.GetLakeDataForUserIDParams{
		UserID: userID,
		LakeID: lakeID,
	})
	if err != nil {
		if err.Error() == errs.PGErrNoRowsFound {
			return &errs.Errorf{
				Type:      errs.ErrNotFound,
				Message:   "Requested lake not found.", // can also be that it does not belong to user.
				ReturnRaw: true,
			}
		}
		return &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get lake data : " + err.Error(),
		}
	}

	// TODO: needs db txn
	// delete creds
	err = s.Queries.DeleteCreds(ctx, lakeID)
	if err != nil {
		if err.Error() != errs.PGErrNoRowsFound {
			return &errs.Errorf{
				Type:    errs.ErrDBQuery,
				Message: "Failed to delete credentials : " + err.Error(),
			}
		}
	}

	// delete lakes, internally deletes locations too.
	err = s.Queries.DeleteLake(ctx, sqlc.DeleteLakeParams{
		UserID: userID,
		LakeID: lakeID,
	})
	if err != nil {
		if err.Error() != errs.PGErrNoRowsFound {
			return &errs.Errorf{
				Type:    errs.ErrDBQuery,
				Message: "Failed to delete lake : " + err.Error(),
			}
		}
	}

	// we are currently relying on the auto delete feature of stash to remove cached data.

	return nil
}

func (s *ManagerService) DeleteLoc(ctx *gin.Context, userID int64, locid string) *errs.Errorf {

	locID, err := strconv.ParseInt(locid, 10, 64)
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrBadForm,
			Message: "Failed to parse lake id as int64 : " + err.Error(),
		}
	}

	// check if user if owner

	locData, err := s.Queries.GetLocationData(ctx, locID)
	if err != nil {
		if err.Error() == errs.PGErrNoRowsFound {
			return &errs.Errorf{
				Type:      errs.ErrNotFound,
				Message:   "Requested resource not found, no such location registered.",
				ReturnRaw: true,
			}
		}
		return &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get location data : " + err.Error(),
		}
	}

	if locData.UserID != userID {
		return &errs.Errorf{
			Type:      errs.ErrUnauthorized,
			Message:   "Requested resource does not belong to you.", // shouldn't say this ??
			ReturnRaw: true,
		}
	}

	// delete loc

	err = s.Queries.DeleteLoc(ctx, sqlc.DeleteLocParams{
		UserID: userID,
		LocID:  locID,
	})
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to delete location : " + err.Error(),
		}
	}

	// we are currently relying on the auto delete feature of stash to remove cached data.

	return nil
}

func (s *ManagerService) GetLakeDetails(ctx *gin.Context, userID int64, lakeid string) (*dto.LakeDetails, *errs.Errorf) {

	lakeID, err := strconv.ParseInt(lakeid, 10, 64)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrBadForm,
			Message: "Failed to parse lake id as int64 : " + err.Error(),
		}
	}

	lakeData, err := s.Queries.GetLakeDataForUserID(ctx, sqlc.GetLakeDataForUserIDParams{
		UserID: userID,
		LakeID: lakeID,
	})
	if err != nil {
		if err.Error() == errs.PGErrNoRowsFound {
			return nil, &errs.Errorf{
				Type:      errs.ErrNotFound,
				Message:   "Requested lake not found.", // can also be that it does not belong to user.
				ReturnRaw: true,
			}
		}
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get lake data : " + err.Error(),
		}
	}

	locs, err := s.Queries.GetLocsListForLake(ctx, sqlc.GetLocsListForLakeParams{
		UserID: userID,
		LakeID: lakeID,
	})
	if err != nil {
		if err.Error() != errs.PGErrNoRowsFound {
			return nil, &errs.Errorf{
				Type:    errs.ErrDBQuery,
				Message: "Failed to get locations list for lake : " + err.Error(),
			}
		}
	}

	locsList := make([]*dto.LocResp, 0)
	for _, loc := range locs {
		locsList = append(locsList, &dto.LocResp{
			LocID:      loc.LocID,
			LakeID:     lakeID,
			BucketName: loc.BucketName,
			CreatedAt:  loc.CreatedAt.Time,
		})
	}

	return &dto.LakeDetails{
		Details: &dto.LakeResp{
			LakeID:    lakeID,
			Name:      lakeData.Name,
			CreatedAt: lakeData.CreatedAt.Time,
			Ptype:     lakeData.Ptype,
			Region:    lakeData.Region,
		},
		Locations: locsList,
	}, nil
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// 0) scan all added locations, for their acl and auths.
// 1) latest history, last changed, etc
// 2) lake size, file type distribution >> this can be debated, maybe just limit it to the added locations and not the entire lake.
// 3) lake size and metrics
// 4) Time-Series Activity Chart
// 5) stats like last scanned,
// 6) other hints and stuff

func (s *ManagerService) GetInitData(ctx *gin.Context, userID int64) *errs.Errorf {

	return nil
}

func (s *ManagerService) GetAllBucsChecks(ctx *gin.Context, userID int64, lakeid string) ([]*dto.LocCheckResp, *errs.Errorf) {

	lakeID, err := strconv.ParseInt(lakeid, 10, 64)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to parse lake id : " + err.Error(),
		}
	}

	locsList, err := s.Queries.GetLocsListForLake(ctx, sqlc.GetLocsListForLakeParams{
		UserID: userID,
		LakeID: lakeID,
	})
	if err != nil {
		if err.Error() == errs.PGErrNoRowsFound {
			return nil, &errs.Errorf{
				Type:      errs.ErrNotFound,
				Message:   "No locations are added for this lake.", // or user not authenticated.
				ReturnRaw: true,
			}
		}
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get locations list for lake : " + err.Error(),
		}
	}

	s3Client, err := s.Stash.GetS3Client(ctx, lakeID)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to get s3 client : " + err.Error(),
		}
	}

	resp := make([]*dto.LocCheckResp, 0)

	for _, loc := range locsList {

		check := new(dto.LocCheckResp)
		check.LocID = loc.LocID
		check.BucketName = loc.BucketName

		oneObj := int32(1)
		_, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:  &loc.BucketName,
			MaxKeys: &oneObj,
		})
		if err != nil {
			var serr *smithy.OperationError
			if errors.As(err, &serr) {
				var httperr *http.ResponseError
				if errors.As(serr.Err, &httperr) && (httperr.HTTPStatusCode() >= 400 && httperr.HTTPStatusCode() < 500) {
					check.ReadCheck = false
				} else {
					return nil, &errs.Errorf{
						Type:    errs.ErrDependencyFailed,
						Message: "Failed to get one object to determine read check : " + err.Error(),
					}
				}
			} else {
				return nil, &errs.Errorf{
					Type:    errs.ErrDependencyFailed,
					Message: "Failed to get one object to determine read check : " + err.Error(),
				}
			}
		} else {
			check.ReadCheck = true
		}

		writeKey := fmt.Sprintf("%s.%d", "temp_obj_lakelens", time.Now().Unix())
		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &loc.BucketName,
			Key:    &writeKey,
		})
		if err != nil {
			var serr *smithy.OperationError
			if errors.As(err, &serr) {
				var httperr *http.ResponseError
				if errors.As(serr.Err, &httperr) && (httperr.HTTPStatusCode() >= 400 && httperr.HTTPStatusCode() < 500) {
					check.WriteCheck = false
				} else {
					return nil, &errs.Errorf{
						Type:    errs.ErrDependencyFailed,
						Message: "Failed to put temp object to determine write check : " + err.Error(),
					}
				}
			} else {
				return nil, &errs.Errorf{
					Type:    errs.ErrDependencyFailed,
					Message: "Failed to put temp object to determine write check : " + err.Error(),
				}
			}
		} else {
			check.WriteCheck = true
		}

		_, err = s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: &loc.BucketName,
		})
		if err != nil {
			var serr *smithy.OperationError
			if errors.As(err, &serr) {
				var httperr *http.ResponseError
				if errors.As(serr.Err, &httperr) && (httperr.HTTPStatusCode() >= 400 && httperr.HTTPStatusCode() < 500) {
					check.AuthCheck = false
				}
			}
			return nil, &errs.Errorf{
				Type:    errs.ErrDependencyFailed,
				Message: "Failed to get head bucket to add location : " + err.Error(),
			}
		} else {
			check.AuthCheck = true
		}

		resp = append(resp, check)
	}

	return resp, nil
}

func (s *ManagerService) GetLakeFileDist(ctx *gin.Context, userID int64, lakeid string) (*dto.LakeFileDist, *errs.Errorf) {

	lakeID, err := strconv.ParseInt(lakeid, 10, 64)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to parse lake id : " + err.Error(),
		}
	}

	locsList, err := s.Queries.GetLocsListForLake(ctx, sqlc.GetLocsListForLakeParams{
		UserID: userID,
		LakeID: lakeID,
	})
	if err != nil {
		if err.Error() == errs.PGErrNoRowsFound {
			return nil, &errs.Errorf{
				Type:      errs.ErrNotFound,
				Message:   "No locations are added for this lake.", // or user not authenticated.
				ReturnRaw: true,
			}
		}
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get locations list for lake : " + err.Error(),
		}
	}

	s3Client, err := s.Stash.GetS3Client(ctx, lakeID)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to get s3 client : " + err.Error(),
		}
	}

	distMp := make(map[string]*dto.LakeFileDistStats, 0)
	var continuationToken *string

	for _, loc := range locsList {
		for {
			objs, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
				Bucket:            &loc.BucketName,
				ContinuationToken: continuationToken,
			})
			if err != nil {
				return nil, &errs.Errorf{
					Type:    errs.ErrDependencyFailed,
					Message: "Failed to list s3 objects : " + err.Error(),
				}
			}

			for _, b := range objs.Contents {
				if ext := path.Ext(*b.Key); ext != "" {
					if distMp[ext] == nil {
						distMp[ext] = &dto.LakeFileDistStats{}
					}

					distMp[ext].TotalSize += *b.Size
					distMp[ext].FileCount += 1
				}
			}

			if !*objs.IsTruncated {
				break
			}
			continuationToken = objs.ContinuationToken
		}
	}

	return &dto.LakeFileDist{
		Dist: distMp,
	}, nil
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

func (s *ManagerService) AnalyzeLake(ctx *gin.Context, userID int64, lakeid string) ([]*dto.BucketData, []*errs.Errorf) {

	lakeID, err := strconv.ParseInt(lakeid, 10, 64)
	if err != nil {
		return nil, []*errs.Errorf{
			{
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

func (s *ManagerService) AnalyzeLoc(ctx *gin.Context, userID int64, locid string) (*dto.BucketData, *errs.Errorf) {

	locID, err := strconv.ParseInt(locid, 10, 64)
	if err != nil {
		return nil, nil
	}

	locData, err := s.Queries.GetLocationData(ctx, locID)
	if err != nil {
		return nil, nil
	}

	if locData.UserID != userID {
		return nil, nil
	}

	lakeData, err := s.Queries.GetLakeData(ctx, locData.LakeID)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	var client CloudClient

	switch lakeData.Ptype {
	case consts.AWSS3:
		s3Client, err := s.Stash.GetS3Client(ctx, locData.LakeID)
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
