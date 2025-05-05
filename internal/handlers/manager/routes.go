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

	// posts a new lake form
	// uses dto.NewLake
	routegrp.POST("/newlake", h.RegisterNewLake)

	// starts analyzing requested lake, lake should obv be already registered
	routegrp.GET("/analyze/:lakeid", h.AnalyzeLake)

	// starts analyzing requested lake, lake should obv be already registered
	routegrp.GET("/analyze/:lakeid/:locid", h.AnalyzeLoc)

}

// extractUserID extracts the user ID and other required parameters from the context with explicit type assertion.
// any returned error is directly included in the response as returned
func (h *ManagerHandler) extractUserID(ctx *gin.Context) (int64, *errs.Errorf) {

	userid, exists := ctx.Get("ID")
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
