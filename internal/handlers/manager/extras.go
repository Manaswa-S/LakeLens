package manager

import (
	"lakelens/internal/consts"
	"net/http"

	"github.com/gin-gonic/gin"
)


func (h *ManagerHandler) GetSearchChoices(ctx *gin.Context) {

	// TODO: these should also include other user related choices, his history, most recent, etc.

	ctx.JSON(http.StatusOK, consts.SearchChoices)
}