package public

import (
	"fmt"
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

func (h *PublicHandler) EPassAuth(ctx *gin.Context) {

	data := new(dto.EPassAuth)
	err := ctx.Bind(data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:      errs.ErrBadForm,
			Message:   "Missing or Invalid fields in user auth form.",
			ReturnRaw: true,
		})
		return
	}

	tks, errf := h.Public.AccAuth(ctx, &dto.UserCreds{
		EPass: data,
	})
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusCreated, tks)
}

func (h *PublicHandler) NewOAuth(ctx *gin.Context) {

	resp, errf := h.Public.NewOAuth(ctx)
	if errf != nil {
		fmt.Println(errf.Message)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

func (h *PublicHandler) OAuthCallback(ctx *gin.Context) {

	data := new(dto.GoogleOAuthCallback)
	err := ctx.Bind(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	tks, errf := h.Public.OAuthCallback(ctx, data)
	if errf != nil {
		fmt.Println(errf.Message)
		return
	}

	ctx.JSON(http.StatusCreated, tks)
}

func (h *PublicHandler) AuthRefresh(ctx *gin.Context) {

	t_ref, err := ctx.Cookie("t_ref")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:    errs.ErrBadRequest,
			Message: "t_ref not found.",
		})
		return
	}

	tks, errf := h.Public.AuthRefresh(ctx, t_ref)
	if errf != nil {
		if errf.ReturnRaw {
			fmt.Println(errf.Message)

			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			fmt.Println(errf.Message)
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, tks)
}

func (h *PublicHandler) AuthCheck(ctx *gin.Context) {

	t_acc, err := ctx.Cookie("t_acc")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:    errs.ErrBadRequest,
			Message: "t_ref not found.",
		})
		return
	}

	resp, errf := h.Public.AuthCheck(ctx, t_acc)
	if errf != nil {
		fmt.Println(errf.Message)
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *PublicHandler) ForgotPass(ctx *gin.Context) {

	data := new(dto.ForgotPassReq)
	err := ctx.Bind(data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:      errs.ErrMissingField,
			Message:   "Invalid or missing fields in form.",
			ReturnRaw: true,
		})
	}

	errf := h.Public.ForgotPass(ctx, data)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			fmt.Println(errf.Message)
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *PublicHandler) VerifyResetPassToken(ctx *gin.Context) {

	token := ctx.Param("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:      errs.ErrMissingField,
			Message:   "Missing token param in url.",
			ReturnRaw: true,
		})
		return
	}

	validity, errf := h.Public.VerifyResetPassToken(ctx, token)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			fmt.Println(errf.Message)
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.JSON(http.StatusOK, validity)
}

func (h *PublicHandler) ResetPass(ctx *gin.Context) {

	data := new(dto.ResetPassReq)
	err := ctx.Bind(data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Errorf{
			Type:      errs.ErrMissingField,
			Message:   "Invalid or missing fields.",
			ReturnRaw: true,
		})
		return
	}

	errf := h.Public.ResetPass(ctx, data)
	if errf != nil {
		if errf.ReturnRaw {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			fmt.Println(errf.Message)
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.Status(http.StatusOK)
}
