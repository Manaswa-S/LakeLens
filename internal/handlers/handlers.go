package handlers

import (
	"github.com/gin-gonic/gin"
	"main.go/internal/services"
	"main.go/internal/utils/s3utils"
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
	routegrp.GET("/test", h.GetList)
}


func (h *Handlers) GetList(ctx *gin.Context) {
	s3utils.ListFromS3()
}