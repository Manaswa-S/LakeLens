package manager

import (
	"encoding/json"
	"fmt"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	"strconv"

	"github.com/gin-gonic/gin"
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
