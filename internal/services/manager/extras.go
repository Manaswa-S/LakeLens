package manager

import (
	"encoding/json"
	"fmt"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	sqlc "lakelens/internal/sqlc/generate"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *ManagerService) GetTip(ctx *gin.Context, userID int64, tipid string) (*dto.TipResp, *errs.Errorf) {

	tipID, err := strconv.ParseInt(tipid, 10, 64)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrInvalidInput,
			Message: "Failed to parse location id to int64 : " + err.Error(),
		}
	}

	tipData, err := s.Queries.GetTipForID(ctx, tipID)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	hrefs := make(map[string]dto.TipRespHRef, 0)
	err = json.Unmarshal(tipData.Hrefs, &hrefs)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	return &dto.TipResp{
		Tip:   tipData.Tip,
		HRefs: hrefs,
	}, nil
}

func (s *ManagerService) GetFeaturesTour(ctx *gin.Context, userID int64) (*dto.FeatureTour, *errs.Errorf) {

	tour, err := s.Queries.GetTour(ctx, userID)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get tour data : " + err.Error(),
		}
	}

	return &dto.FeatureTour{
		LastTour: tour.Version.Int32,
	}, nil
}

func (s *ManagerService) UpdateFeaturesTour(ctx *gin.Context, userID int64, versionStr string) *errs.Errorf {

	version, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrBadForm,
			Message: "Failed to parse version str to int64 : " + err.Error(),
		}
	}

	err = s.Queries.UpdateTourStatus(ctx, sqlc.UpdateTourStatusParams{
		UserID:  userID,
		Version: pgtype.Int4{Int32: int32(version), Valid: true},
		ShownAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to update tour status : " + err.Error(),
		}
	}

	return nil
}

func (s *ManagerService) GetRecentActivity(ctx *gin.Context, userID int64, offset string) ([]*dto.RecentsResp, *errs.Errorf) {

	// offSet, err := strconv.ParseInt(offset, 10, 64)
	// if err != nil {
	// 	return nil, &errs.Errorf{
	// 		Type:    errs.ErrInvalidInput,
	// 		Message: "Failed to parse offset to int64 : " + err.Error(),
	// 	}
	// }

	// recents, err := s.Queries.GetRecents(ctx, sqlc.GetRecentsParams{
	// 	UserID: userID,
	// 	Limit:  20,
	// 	Offset: int32(offSet),
	// })
	// if err != nil {
	// 	// TODO: if no rows found
	// 	return nil, &errs.Errorf{
	// 		Type:    errs.ErrDBQuery,
	// 		Message: "Failed to get recents from db : " + err.Error(),
	// 	}
	// }

	return nil, nil
}
