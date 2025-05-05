package public

import "github.com/gin-gonic/gin"


func (h *PublicHandler) RegisterRoutes(routegrp *gin.RouterGroup) {
	routegrp.POST("/newuser", h.NewUser)
}
