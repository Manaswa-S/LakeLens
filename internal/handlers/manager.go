package handlers

import (
	"fmt"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	"lakelens/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)


type ManagerHandler struct {
	Manager *services.ManagerService
}

func NewManagerHandler(manager *services.ManagerService) *ManagerHandler {
	return &ManagerHandler{
		Manager: manager,
	}
}

func (h *ManagerHandler) RegisterRoutes(routegrp *gin.RouterGroup) {
	
	// posts a new lake form
	// uses dto.NewLake
	routegrp.POST("/newlake", h.RegisterNewLake)

	routegrp.GET("/getdata/:lakeid", h.GetLakeMetaData)

	routegrp.GET("/getdata/:lakeid/:locid", h.GetLocMetaData)

}



func (h *ManagerHandler) RegisterNewLake(ctx *gin.Context) {

	data := new(dto.NewLake)
	err := ctx.Bind(data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type: errs.ErrBadForm,
			Message: "Missing or invalid form format.",
			ReturnRaw: true,
		})
		return
	}

	// TODO: add JWT and then ctx data extraction

	buckets, errf := h.Services.RegisterNewLake(ctx, 1000000, data)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "New lake registered successfully.",
		"buckets": buckets,
	})
} 


func (h *ManagerHandler) GetLakeMetaData(ctx *gin.Context) {

	lakeid := ctx.Param("lakeid")
	if lakeid == "" {
		return
	}

	response, errfs := h.Services.GetLakeData(ctx, 1, lakeid)
	if len(errfs) != 0 {
		fmt.Println(errfs)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *ManagerHandler) GetLocMetaData(ctx *gin.Context) {

	lakeid := ctx.Param("lakeid")
	if lakeid == "" {
		return
	}

	locid := ctx.Param("locid")
	if locid == "" {
		return
	}

	response, errf := h.Services.GetLocData(ctx, 1, lakeid, locid)
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

