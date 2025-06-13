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

	userID, errf := h.getUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	buckets, errf := h.Manager.RegisterNewLake(ctx, userID, data)
	if errf != nil {
		fmt.Println(errf.Message)
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusCreated, buckets)
}

func (h *ManagerHandler) AddLocations(ctx *gin.Context) {

	data := new(dto.AddLocsReq)
	err := ctx.Bind(data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:      errs.ErrBadForm,
			Message:   "Missing or invalid form format.",
			ReturnRaw: true,
		})
		return
	}

	userID, errf := h.getUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.Manager.AddLocations(ctx, userID, data)
	if errf != nil {
		fmt.Println(errf.Message)
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusCreated, nil)
}

func (h *ManagerHandler) AccDetails(ctx *gin.Context) {

	userID, errf := h.getUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	accDetails, errf := h.Manager.AccDetails(ctx, userID)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			fmt.Println(errf.Message)
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.JSON(http.StatusOK, accDetails)
}

func (h *ManagerHandler) AccBilling(ctx *gin.Context) {

	userID, errf := h.getUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	accBilling, errf := h.Manager.AccBilling(ctx, userID)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			fmt.Println(errf.Message)
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.JSON(http.StatusOK, accBilling)
}

func (h *ManagerHandler) AccProjects(ctx *gin.Context) {

	userID, errf := h.getUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	accProjects, errf := h.Manager.AccProjects(ctx, userID)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			fmt.Println(errf.Message)
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.JSON(http.StatusOK, accProjects)
}

func (h *ManagerHandler) AccSettings(ctx *gin.Context) {

	userID, errf := h.getUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	accSettings, errf := h.Manager.AccSettings(ctx, userID)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			fmt.Println(errf.Message)
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.JSON(http.StatusOK, accSettings)
}

func (h *ManagerHandler) AccSettingsUpdate(ctx *gin.Context) {

	data := new(dto.AccSettingsResp)
	err := ctx.Bind(data)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:      errs.ErrInvalidFormat,
			Message:   "Missing or invalid form format.",
			ReturnRaw: true,
		})
		return
	}

	userID, errf := h.getUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.Manager.AccSettingsUpdate(ctx, data, userID)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			fmt.Println(errf.Message)
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Message": "Settings updated.",
	})
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

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

	userID, errf := h.getUserID(ctx)
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
