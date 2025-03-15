package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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

	routegrp.POST("/newlake", h.RegisterNewLake)
	
	routegrp.GET("/getdata", h.GetMetaData)
}

func (h *Handlers) RegisterNewLake(ctx *gin.Context) {

	data := new(dto.NewLake)
	err := ctx.Bind(data)
	if err != nil {
		return
	}

	err = h.Services.RegisterNewLake(ctx, data)
	if err != nil {
		return
	}

	ctx.Status(http.StatusOK)
} 



func (h *Handlers) GetMetaData(ctx *gin.Context) {

	// TODO: get userid from token
	// TODO: get internal lakeid from token

	lakeid := ctx.Query("lakeid")
	if lakeid == "" {
		return
	}

	response, err := h.Services.GetMetaData(ctx, 1, lakeid)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}