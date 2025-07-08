package iceberg

import (
	"lakelens/internal/consts/errs"

	"github.com/gin-gonic/gin"
)

func (h *IcebergHandler) RegisterRoutes(routegrp *gin.RouterGroup) {

	routegrp.GET("/overview/data/:locid", h.GetOverviewData)
	routegrp.GET("/overview/stats/:locid", h.GetOverviewStats)
	routegrp.GET("/overview/schema/:locid", h.GetOverviewSchema)
	routegrp.GET("/overview/partition/:locid", h.GetOverviewPartition)
	routegrp.GET("/overview/snapshot/:locid", h.GetOverviewSnapshot)
	routegrp.GET("/overview/graphs/:locid", h.GetOverviewGraphs)

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	routegrp.GET("/schema/compare/list/:locid", h.GetSchemasList)
	routegrp.GET("/schema/compare/getschema/:locid/:schemaid", h.GetSchema)
	routegrp.GET("/schema/data/:locid/:schemaid", h.GetSchemaData)
	routegrp.GET("/schema/colsizes/:locid/:schemaid", h.GetSchemaColSizes)

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	// routegrp.GET("/alldata/:lakeid/:locid", h.AllData)

	// routegrp.GET("/metadata/:lakeid/:locid", h.Metadata)
	// routegrp.GET("/snapshot/:lakeid/:locid", h.Snapshot)
	// routegrp.GET("/manifest/:lakeid/:locid", h.Manifest)

}

// extractUserID extracts the user ID and other required parameters from the context with explicit type assertion.
// any returned error is directly included in the response as returned
func (h *IcebergHandler) getUserID(ctx *gin.Context) (int64, *errs.Errorf) {

	userid, exists := ctx.Get("rid")
	if !exists {
		return 0, &errs.Errorf{
			Type:      errs.ErrInvalidCredentials,
			Message:   "Missing user ID in request.",
			ReturnRaw: true,
		}
	}

	userID, ok := userid.(int64)
	if !ok {
		return 0, &errs.Errorf{
			Type:      errs.ErrInvalidFormat,
			Message:   "User ID of improper format.",
			ReturnRaw: true,
		}
	}

	return userID, nil
}
