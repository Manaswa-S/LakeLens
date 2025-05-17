package manager

import (
	"fmt"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *ManagerHandler) RegisterNewLake(ctx *gin.Context) {

	data := new(dto.NewLake)
	err := ctx.Bind(data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:      errs.ErrBadForm,
			Message:   "Missing or invalid form format.",
			ReturnRaw: true,
		})
		return
	}

	// TODO: add JWT and then ctx data extraction

	buckets, errf := h.Manager.RegisterNewLake(ctx, 1000000, data)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "New lake registered successfully.",
		"buckets": buckets,
	})
}

func (h *ManagerHandler) AnalyzeLake(ctx *gin.Context) {

	lakeid := ctx.Param("lakeid")
	if lakeid == "" {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:      errs.ErrMissingField,
			Message:   "Missing lakeid param in url.",
			ReturnRaw: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	response, errfs := h.Manager.AnalyzeLake(ctx, userID, lakeid)
	if len(errfs) != 0 {
		errResp := make([]*errs.Errorf, 0)
		for _, errf := range errfs {
			if errf.ReturnRaw {
				errResp = append(errResp, errf)
			} else {
				// TODO: handle, probably send over to the error channel
				fmt.Println(errf.Message)
			}
		}
		ctx.JSON(http.StatusBadRequest, errResp)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *ManagerHandler) AnalyzeLoc(ctx *gin.Context) {

	lakeid := ctx.Param("lakeid")
	if lakeid == "" {
		return
	}

	locid := ctx.Param("locid")
	if locid == "" {
		return
	}

	// userID, errf := h.extractUserID(ctx)
	// if errf != nil {
	// 	ctx.JSON(http.StatusBadRequest, errf)
	// 	return
	// }

	response, errf := h.Manager.AnalyzeLoc(ctx, 1, lakeid, locid)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			fmt.Println(errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *ManagerHandler) FetchLocation(ctx *gin.Context) {

	lakeid := ctx.Param("lakeid")
	if lakeid == "" {
		return
	}

	locid := ctx.Param("locid")
	if locid == "" {
		return
	}

	// userID, errf := h.extractUserID(ctx)
	// if errf != nil {
	// 	ctx.JSON(http.StatusBadRequest, errf)
	// 	return
	// }

	response, errf := h.Manager.FetchLocation(ctx, 1000000, locid)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			fmt.Println(errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, response)
}
