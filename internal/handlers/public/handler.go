package public

import (
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	"lakelens/internal/services/public"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PublicHandler struct {
	Public *public.PublicService
}

func NewPublicHandler(public *public.PublicService) *PublicHandler {
	return &PublicHandler{
		Public: public,
	}
}

func (h *PublicHandler) NewUser(ctx *gin.Context) {
	data := new(dto.NewUser)
	err := ctx.Bind(data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:      errs.ErrBadForm,
			Message:   "Missing or Invalid fields in new user form.",
			ReturnRaw: true,
		})
		return
	}

	errf := h.Public.NewUser(ctx, data)
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
