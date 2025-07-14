package manager

import (
	"fmt"
	"lakelens/internal/consts"
	"lakelens/internal/consts/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *ManagerHandler) GetSearchChoices(ctx *gin.Context) {

	// TODO: these should also include other user related choices, his history, most recent, etc.

	ctx.JSON(http.StatusOK, consts.SearchChoices)
}

func (h *ManagerHandler) GetTip(ctx *gin.Context) {

	tipid := ctx.Param("tipid")
	if tipid == "" {
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

	response, errf := h.Manager.GetTip(ctx, userID, tipid)
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

func (h *ManagerHandler) GetRecentActivity(ctx *gin.Context) {

	offset := ctx.Param("offset")
	if offset == "" {
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

	response, errf := h.Manager.GetRecentActivity(ctx, userID, offset)
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
