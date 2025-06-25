package iceberg

import (
	"fmt"
	"lakelens/internal/consts/errs"
	"lakelens/internal/services/iceberg"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IcebergHandler struct {
	Iceberg *iceberg.IcebergService
}

func NewIcebergHandler(iceberg *iceberg.IcebergService) *IcebergHandler {
	return &IcebergHandler{
		Iceberg: iceberg,
	}
}

func (h *IcebergHandler) GetOverviewData(ctx *gin.Context) {

	locid := ctx.Param("locid")
	if locid == "" {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:      errs.ErrMissingField,
			Message:   "Missing url params.",
			ReturnRaw: true,
		})
		return
	}

	userID, errf := h.getUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	response, errf := h.Iceberg.GetOverviewData(ctx, userID, locid)
	if errf != nil {
		fmt.Println(errf.Message)
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *IcebergHandler) GetOverviewStats(ctx *gin.Context) {

	locid := ctx.Param("locid")
	if locid == "" {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:      errs.ErrMissingField,
			Message:   "Missing url params.",
			ReturnRaw: true,
		})
		return
	}

	userID, errf := h.getUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	response, errf := h.Iceberg.GetOverviewStats(ctx, userID, locid)
	if errf != nil {
		fmt.Println(errf.Message)
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// func (h *IcebergHandler) AllData(ctx *gin.Context) {

// 	locid := ctx.Param("locid")
// 	if locid == "" {
// 		ctx.JSON(http.StatusBadRequest, errs.Errorf{
// 			Type:      errs.ErrMissingField,
// 			Message:   "Missing url params.",
// 			ReturnRaw: true,
// 		})
// 		return
// 	}

// 	userID, errf := h.getUserID(ctx)
// 	if errf != nil {
// 		ctx.JSON(http.StatusBadRequest, errf)
// 		return
// 	}

// 	response, errf := h.Iceberg.AllData(ctx, userID, locid)
// 	if errf != nil {
// 		fmt.Println(errf.Message)
// 		if errf.ReturnRaw {
// 			ctx.JSON(http.StatusBadRequest, errf)
// 		} else {
// 			ctx.Set("error", errf.Message)
// 			ctx.Status(http.StatusInternalServerError)
// 		}
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, response)
// }

// func (h *IcebergHandler) Metadata(ctx *gin.Context) {

// 	lakeid := ctx.Param("lakeid")
// 	if lakeid == "" {
// 		return
// 	}

// 	locid := ctx.Param("locid")
// 	if locid == "" {
// 		return
// 	}

// 	// userID, errf := h.extractUserID(ctx)
// 	// if errf != nil {
// 	// 	ctx.JSON(http.StatusBadRequest, errf)
// 	// 	return
// 	// }

// 	response, errf := h.Iceberg.Metadata(ctx, 1000000, locid)
// 	if errf != nil {
// 		if errf.ReturnRaw {
// 			ctx.JSON(http.StatusBadRequest, errf)
// 		} else {
// 			fmt.Println(errf.Message)
// 		}
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, response)

// }

// func (h *IcebergHandler) Snapshot(ctx *gin.Context) {

// 	lakeid := ctx.Param("lakeid")
// 	if lakeid == "" {
// 		return
// 	}

// 	locid := ctx.Param("locid")
// 	if locid == "" {
// 		return
// 	}

// 	// userID, errf := h.extractUserID(ctx)
// 	// if errf != nil {
// 	// 	ctx.JSON(http.StatusBadRequest, errf)
// 	// 	return
// 	// }

// 	response, errf := h.Iceberg.Snapshot(ctx, 1000000, locid)
// 	if errf != nil {
// 		if errf.ReturnRaw {
// 			ctx.JSON(http.StatusBadRequest, errf)
// 		} else {
// 			fmt.Println(errf.Message)
// 		}
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, response)

// }

// func (h *IcebergHandler) Manifest(ctx *gin.Context) {

// 	lakeid := ctx.Param("lakeid")
// 	if lakeid == "" {
// 		return
// 	}

// 	locid := ctx.Param("locid")
// 	if locid == "" {
// 		return
// 	}

// 	// userID, errf := h.extractUserID(ctx)
// 	// if errf != nil {
// 	// 	ctx.JSON(http.StatusBadRequest, errf)
// 	// 	return
// 	// }

// 	response, errf := h.Iceberg.Manifest(ctx, 1000000, locid)
// 	if errf != nil {
// 		if errf.ReturnRaw {
// 			ctx.JSON(http.StatusBadRequest, errf)
// 		} else {
// 			fmt.Println(errf.Message)
// 		}
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, response)
// }
