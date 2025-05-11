package iceberg

import "github.com/gin-gonic/gin"


func (h *IcebergHandler) RegisterRoutes(routegrp *gin.RouterGroup) {

	routegrp.GET("/alldata/:lakeid/:locid", h.AllData)

	routegrp.GET("/metadata/:lakeid/:locid", h.Metadata)
	routegrp.GET("/snapshot/:lakeid/:locid", h.Snapshot)
	routegrp.GET("/manifest/:lakeid/:locid", h.Manifest)


}
