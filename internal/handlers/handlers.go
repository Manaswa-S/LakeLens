package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"main.go/internal/consts/errs"
	"main.go/internal/dto"
	"main.go/internal/services"
)

type Handlers struct {
	Services *services.Services
}

func NewHandler(services *services.Services) *Handlers {
	return &Handlers{
		Services: services,
	}
}

func (h *Handlers) RegisterRoutes(routegrp *gin.RouterGroup) {

	routegrp.POST("/newuser", h.NewUser)

	routegrp.POST("/newlake", h.RegisterNewLake)
	
	routegrp.GET("/getdata/:lakeid", h.GetLakeMetaData)
	routegrp.GET("/getdata/:lakeid/:locid", h.GetLocMetaData)
}


func (h *Handlers) NewUser(ctx *gin.Context) {
	data := new(dto.NewUser)
	err := ctx.Bind(data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type: errs.ErrBadForm,
			Message: "Missing or Invalid fields in new user form.",
			ReturnRaw: true,
		})
		return
	}

	errf := h.Services.NewUser(ctx, data)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "New user registered successfully.",
	})
}
func (h *Handlers) RegisterNewLake(ctx *gin.Context) {

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




func (h *Handlers) GetLakeMetaData(ctx *gin.Context) {

	lakeid := ctx.Param("lakeid")
	if lakeid == "" {
		return
	}

	response, err := h.Services.GetLakeMetaData(ctx, 1, lakeid)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handlers) GetLocMetaData(ctx *gin.Context) {

	locid := ctx.Param("locid")
	if locid == "" {
		return
	}

	response, err := h.Services.GetLocMetaData(ctx, 1, locid)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}