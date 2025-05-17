package public

import "github.com/gin-gonic/gin"

func (h *PublicHandler) RegisterRoutes(routegrp *gin.RouterGroup) {

	// post the epass auth form
	routegrp.POST("/acc/auth", h.EPassAuth)
	// get the oauth redirect url, primarily only google for now
	routegrp.GET("/acc/oauth", h.NewOAuth)
	// post the oauth callback details
	routegrp.POST("/acc/oauth/callback", h.OAuthCallback)
	// post the t_ref for t_acc refresh
	routegrp.POST("/acc/auth/refresh", h.AuthRefresh)
	// get t_acc checked and some user details are returned
	routegrp.GET("/acc/auth/check", h.AuthCheck)

	// generate the token and send the email
	routegrp.POST("/acc/auth/forgot-pass", h.ForgotPass)
	// get the validity of the token
	routegrp.GET("/acc/auth/reset-pass/verify/:token", h.VerifyResetPassToken)
	// post the form, {new-pass and token}
	routegrp.POST("/acc/auth/reset-pass", h.ResetPass)
}
