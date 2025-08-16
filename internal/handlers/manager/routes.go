package manager

import (
	"lakelens/internal/consts/errs"
	"lakelens/internal/services/manager"

	"github.com/gin-gonic/gin"
)

type ManagerHandler struct {
	Manager *manager.ManagerService
}

func NewManagerHandler(manager *manager.ManagerService) *ManagerHandler {
	return &ManagerHandler{
		Manager: manager,
	}
}

func (h *ManagerHandler) RegisterRoutes(routegrp *gin.RouterGroup) {

	// sends the user data for dashboard, etc.
	routegrp.GET("/account/details", h.AccDetails)
	routegrp.GET("/account/billing", h.AccBilling)
	routegrp.GET("/account/projects", h.AccProjects)
	routegrp.GET("/account/settings", h.AccSettings)
	routegrp.PATCH("/account/settings/update", h.AccSettingsUpdate)

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	routegrp.GET("/search/choices", h.GetSearchChoices)
	routegrp.GET("/features/tour", h.GetFeaturesTour)
	routegrp.PATCH("/features/tour/:version", h.UpdateFeaturesTour)
	routegrp.GET("/recent/activity/:offset", h.GetRecentActivity)

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	// posts a new lake form
	routegrp.POST("/newlake", h.RegisterNewLake)
	// gets the locations for user given the lake id, if any.
	routegrp.GET("/getlocs/:lakeid", h.GetLocations)
	// posts the new locations to be added to an existing lake
	routegrp.POST("/addlocs", h.AddLocations)

	// deletes a lake, give the lakeid.
	routegrp.DELETE("/lake/:lakeid", h.DeleteLake)
	// deletes a location, give the location id.
	routegrp.DELETE("/loc/:locid", h.DeleteLoc)

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	routegrp.GET("/lake/details/:lakeid", h.GetLakeDetails)

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	routegrp.GET("/meta/file-dist/:lakeid", h.GetLakeFileDist)
	routegrp.GET("/meta/bucs-check/:lakeid", h.GetAllBucsChecks)

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	routegrp.GET("/tip/:tipid", h.GetTip)

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	// starts analyzing requested lake, lake should obv be already registered
	routegrp.GET("/analyze/:lakeid", h.AnalyzeLake)
	// starts analyzing requested lake, lake should obv be already registered
	routegrp.GET("/analyze/loc/:locid", h.AnalyzeLoc)
	// returns the entire report of a location
	routegrp.GET("/fetch/:lakeid/:locid", h.FetchLocation)
}

// extractUserID extracts the user ID and other required parameters from the context with explicit type assertion.
// any returned error is directly included in the response as returned
func (h *ManagerHandler) getUserID(ctx *gin.Context) (int64, *errs.Errorf) {

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
