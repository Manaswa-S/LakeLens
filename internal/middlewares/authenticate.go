package middlewares

import (
	"errors"
	"fmt"
	"lakelens/internal/auth"
	"lakelens/internal/consts/errs"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Authenticator struct {
	AllowedIssuers map[string]bool
	BannedJWTIDs   map[string]int64

	AuthClient  *auth.AuthService
	RedisClient *redis.Client
}

func NewAuthMiddleware(allowedIssuers map[string]bool, redis *redis.Client, auth *auth.AuthService) *Authenticator {
	return &Authenticator{
		AllowedIssuers: allowedIssuers,
		BannedJWTIDs:   make(map[string]int64),

		RedisClient: redis,
		AuthClient:  auth,
	}
}

func (a *Authenticator) Authenticator() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		atoken, err := ctx.Cookie("t_acc")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				ctx.JSON(http.StatusUnauthorized, errs.Errorf{
					Type:      errs.ErrBadRequest,
					Message:   "t_acc not found.",
					ReturnRaw: true,
				})
				ctx.Abort()
				return
			}
			// unknown err
			ctx.JSON(http.StatusUnauthorized, errs.Errorf{
				Type:      errs.ErrBadRequest,
				Message:   "Failed to get t_acc from cookies.",
				ReturnRaw: true,
			})
			ctx.Abort()
			return
		}

		// verify jwt
		aclaims, errf := a.AuthClient.VerifyJWT(atoken)
		if errf != nil {
			if errf.ReturnRaw {
				ctx.JSON(http.StatusUnauthorized, errf)
			} else {
				fmt.Println(errf.Message)
				ctx.Status(http.StatusUnauthorized)
			}
			ctx.Abort()
			return
		}

		atjwt, errf := a.AuthClient.ParseAT(aclaims)
		if errf != nil {
			if errf.ReturnRaw {
				ctx.JSON(http.StatusUnauthorized, errf)
			} else {
				fmt.Println(errf.Message)
			}
			ctx.Abort()
			return
		}

		// check for validity

		if atjwt.ExpiresAt <= time.Now().Unix() {
			ctx.JSON(http.StatusUnauthorized, errs.Errorf{
				Type:      errs.ErrBadRequest,
				Message:   "t_acc token expired.",
				ReturnRaw: true,
			})
			ctx.Abort()
			return
		}
		_, ok := a.AllowedIssuers[atjwt.Issuer]
		if !ok {
			ctx.JSON(http.StatusUnauthorized, errs.Errorf{
				Type:      errs.ErrBadRequest,
				Message:   "unknown issuer.",
				ReturnRaw: true,
			})
			ctx.Abort()
			return
		}
		_, ok = a.BannedJWTIDs[atjwt.JWTID]
		if ok {
			ctx.JSON(http.StatusUnauthorized, errs.Errorf{
				Type:      errs.ErrBadRequest,
				Message:   "t_acc has been banned.",
				ReturnRaw: true,
			})
			ctx.Abort()
			return
		}

		// set data on ctx line

		ctx.Set("rid", atjwt.UserID)

		// proceed

		ctx.Next()
	}
}

// TODO:
func (a *Authenticator) BanJWTID(jwtID string) bool {

	return false
}
