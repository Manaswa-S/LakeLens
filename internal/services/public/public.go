package public

import (
	"encoding/json"
	"errors"
	"fmt"
	"lakelens/internal/auth"
	configs "lakelens/internal/config"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	sqlc "lakelens/internal/sqlc/generate"
	utils "lakelens/internal/utils/common"
	"lakelens/internal/utils/common/emails"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

type PublicService struct {
	Queries     *sqlc.Queries
	RedisClient *redis.Client
	DB          *pgxpool.Pool

	GOAuth     *oauth2.Config
	AuthClient *auth.AuthService
}

func NewPublicService(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool, goauth *oauth2.Config, auth *auth.AuthService) *PublicService {
	return &PublicService{
		Queries:     queries,
		RedisClient: redis,
		DB:          db,
		GOAuth:      goauth,
		AuthClient:  auth,
	}
}

func (s *PublicService) NewOAuth(ctx *gin.Context) (*dto.NewOAuth, *errs.Errorf) {

	state, err := s.genuuidv7()
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrEnvNotFound,
			Message: "Failed to generate oauth state : " + err.Error(),
		}
	}
	url := s.GOAuth.AuthCodeURL(state)

	return &dto.NewOAuth{
		URL:      &url,
		State:    &state,
		StateTTL: 300,
	}, nil
}
func (s *PublicService) genuuidv7() (string, error) {
	uid, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}
func (s *PublicService) timeSinceuuidv7(uidStr string) (int64, error) {
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return 0, err
	}
	if uid.Version() != 7 {
		return 0, errors.New("requested uuid is not of version 7")
	}
	secs, _ := uid.Time().UnixTime()
	return secs, nil
}

func (s *PublicService) OAuthCallback(ctx *gin.Context, authData *dto.GoogleOAuthCallback) (*dto.AuthRet, *errs.Errorf) {

	if authData.State != authData.CookieState {
		return nil, &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "States dont match",
		}
	}

	stateSecs, err := s.timeSinceuuidv7(authData.State)
	if err != nil {
		return nil, &errs.Errorf{
			Type:      errs.ErrTokenInvalid,
			Message:   "State is not of expected format.",
			ReturnRaw: true,
		}
	}
	if stateSecs >= 300 {
		return nil, &errs.Errorf{
			Type:      errs.ErrTokenExpired,
			Message:   "State has expired. Try again.",
			ReturnRaw: true,
		}
	}

	token, err := s.GOAuth.Exchange(ctx, authData.Code)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to exchange : " + err.Error(),
		}
	}

	// what to do after this ?
	client := s.GOAuth.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to get user info: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	userInfo := new(dto.GoogleOAuth)
	if err := json.NewDecoder(resp.Body).Decode(userInfo); err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to decode user info: " + err.Error(),
		}
	}

	// generate tokens and return them

	authRet, errf := s.AccAuth(ctx, &dto.UserCreds{
		GOAuth: userInfo,
	})
	if errf != nil {
		return nil, errf
	}

	return authRet, nil
}

// TODO: the email verified check hasnt been added
func (s *PublicService) AccAuth(ctx *gin.Context, creds *dto.UserCreds) (*dto.AuthRet, *errs.Errorf) {

	userID := int64(0)
	errf := new(errs.Errorf)
	amt := ""

	switch {

	case creds.EPass != nil:
		amt = configs.EPassAuth
		userID, errf = s.epassAuth(ctx, creds.EPass)

	case creds.GOAuth != nil:
		amt = configs.GoogleOAuth
		userID, errf = s.googleOAuth(ctx, creds.GOAuth)

	default:
		return nil, &errs.Errorf{
			Type:      errs.ErrOutOfRange,
			Message:   "Missing or unrecognized authentication method.",
			ReturnRaw: true,
		}
	}
	if errf != nil {
		return nil, errf
	}

	userData, err := s.Queries.GetUserData(ctx, userID)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get user data for auth.",
		}
	}

	role := "client1"
	uuid := userData.UserUuid.String()

	at, errf := s.AuthClient.SignAT(&auth.ATJWTParams{
		Issuer:     s.AuthClient.Creds.AccAuthIssuer,
		Subject:    s.AuthClient.Creds.AccAuthSub,
		Role:       role,
		UserID:     userID,
		UUID:       uuid,
		AuthMethod: amt,
	})
	if errf != nil {
		return nil, errf
	}

	rt, errf := s.AuthClient.SignRT(&auth.RTJWTParams{
		Issuer:     s.AuthClient.Creds.AccAuthIssuer,
		Subject:    s.AuthClient.Creds.AccAuthSub,
		Role:       role,
		UserID:     userID,
		UUID:       uuid,
		AuthMethod: amt,
	})
	if errf != nil {
		return nil, errf
	}

	return &dto.AuthRet{
		ATStr: &at,
		RTStr: &rt,

		ATTTL: s.AuthClient.Creds.ATTTL,
		RTTTL: s.AuthClient.Creds.RTTTL,
	}, nil
}

func (s *PublicService) epassAuth(ctx *gin.Context, epass *dto.EPassAuth) (int64, *errs.Errorf) {

	newUser := false
	userData, err := s.Queries.GetUserFromEmail(ctx, epass.Email)
	if err != nil {
		if err.Error() == errs.PGErrNoRowsFound {
			newUser = true
		} else {
			return 0, &errs.Errorf{
				Type:    errs.ErrDBQuery,
				Message: "Failed to get user from email : " + err.Error(),
			}
		}
	}

	if newUser {

		passHash, err := bcrypt.GenerateFromPassword([]byte(epass.Password), configs.Internal.BcryptPasswordCost)
		if err != nil {
			return 0, &errs.Errorf{
				Type:    errs.ErrInternalServer,
				Message: "Failed to encrypt password : " + err.Error(),
			}
		}

		userID, err := s.Queries.InsertNewUser(ctx, sqlc.InsertNewUserParams{
			Email:    epass.Email,
			Password: string(passHash),
		})
		if err != nil {
			return 0, &errs.Errorf{
				Type:    errs.ErrDBQuery,
				Message: "Failed to insert new user into db : " + err.Error(),
			}
		}

		return userID, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(epass.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return 0, &errs.Errorf{
				Type:      errs.ErrInvalidCredentials,
				Message:   "Credentials did not match. Try again.",
				ReturnRaw: true,
			}
		}
		return 0, &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to compare passwords.",
		}
	}

	return userData.UserID, nil
}

func (s *PublicService) googleOAuth(ctx *gin.Context, goauth *dto.GoogleOAuth) (int64, *errs.Errorf) {

	newUser := false
	userData, err := s.Queries.GetUserFromEmail(ctx, goauth.Email)
	if err != nil {
		if err.Error() == errs.PGErrNoRowsFound {
			newUser = true
		} else {
			return 0, &errs.Errorf{
				Type:    errs.ErrDBQuery,
				Message: "Failed to get user from email : " + err.Error(),
			}
		}
	}

	if newUser {

		userID, err := s.Queries.InsertNewUser(ctx, sqlc.InsertNewUserParams{
			Email:    goauth.Email,
			Password: string(""),
		})
		if err != nil {
			return 0, &errs.Errorf{
				Type:    errs.ErrDBQuery,
				Message: "Failed to insert new user into db : " + err.Error(),
			}
		}

		err = s.Queries.InsertNewGOAuth(ctx, sqlc.InsertNewGOAuthParams{
			UserID:  userID,
			Email:   goauth.Email,
			Name:    pgtype.Text{String: goauth.Name, Valid: true},
			Picture: pgtype.Text{String: goauth.Picture, Valid: true},
			ID:      pgtype.Text{String: goauth.Id, Valid: true},
		})
		if err != nil {
			return 0, &errs.Errorf{
				Type:    errs.ErrDBQuery,
				Message: "Failed to insert new user into goauth : " + err.Error(),
			}
		}

		return userID, nil
	}

	gData, err := s.Queries.GetGoogleID(ctx, userData.UserID)
	if err != nil {
		return 0, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get google auth data from db : " + err.Error(),
		}
	}

	if gData.Email != goauth.Email || gData.ID.String != goauth.Id {
		return 0, &errs.Errorf{
			Type:      errs.ErrConflict,
			Message:   "Credential fail to match.",
			ReturnRaw: true,
		}
	}

	return userData.UserID, nil
}

func (s *PublicService) AuthRefresh(ctx *gin.Context, t_ref string) (*dto.AuthRet, *errs.Errorf) {

	claims, errf := s.AuthClient.VerifyJWT(t_ref)
	if errf != nil {
		return nil, errf
	}

	rtjwt, errf := s.AuthClient.ParseRT(claims)
	if errf != nil {
		return nil, errf
	}

	t_acc, errf := s.AuthClient.RefreshAT(rtjwt)
	if errf != nil {
		return nil, errf
	}

	return &dto.AuthRet{
		ATStr: &t_acc,
		ATTTL: s.AuthClient.Creds.ATTTL,
	}, nil
}

func (s *PublicService) AuthCheck(ctx *gin.Context, t_acc string) (*dto.AuthCheckRet, *errs.Errorf) {

	claims, errf := s.AuthClient.VerifyJWT(t_acc)
	if errf != nil {
		return nil, errf
	}

	atjwt, errf := s.AuthClient.ParseAT(claims)
	if errf != nil {
		return nil, errf
	}

	// get essential data like username, etc.
	userData, err := s.Queries.GetUserData(ctx, atjwt.UserID)
	if err != nil {
		if err.Error() == errs.PGErrNoRowsFound {
			return nil, &errs.Errorf{
				Type:      errs.ErrUnauthorized,
				Message:   "User not found.",
				ReturnRaw: true,
			}
		}
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get user data from db : " + err.Error(),
		}
	}

	return &dto.AuthCheckRet{
		Name:      "UName",
		Picture:   "https://lh3.googleusercontent.com/a/ACg8ocIEGbv1m0cOz2_T9vIUSoK3Fpkoo5n-KSZnR89B1_TdNJ_j8g=s96-c",
		Confirmed: userData.Confirmed,
	}, nil

}

func (s *PublicService) ForgotPass(ctx *gin.Context, data *dto.ForgotPassReq) *errs.Errorf {

	if _, ok := utils.ValidateEmail(data.Email); !ok {
		return &errs.Errorf{
			Type:      errs.ErrBadRequest,
			Message:   "Email has invalid structure.",
			ReturnRaw: true,
		}
	}

	userID, err := s.Queries.CheckIfEPassOnly(ctx, data.Email)
	if err != nil {
		// if not return silently, dont confirm to user
		if err.Error() == errs.PGErrNoRowsFound {
			return nil
		}
		return &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to check if user email exists : " + err.Error(),
		}
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to generate uuid for forgot-pass token : " + err.Error(),
		}
	}
	token := uid.String()

	key := fmt.Sprintf("forgotpass:%s", token)
	err = s.RedisClient.Set(ctx, key, userID, 900*time.Second).Err()
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to set token in redis : " + err.Error(),
		}
	}

	fendBase, exists := os.LookupEnv("FRONTEND_URL")
	if !exists {
		return &errs.Errorf{
			Type:    errs.ErrEnvNotFound,
			Message: "Frontend base url not found in env.",
		}
	}
	fpath := "/account/reset-pass?token="

	link := fmt.Sprintf("%s%s%s", fendBase, fpath, token)
	fmt.Println(link)
	// works async, with inbuilt retry mechanism
	emails.ResetPassEmail(link, data.Email)

	return nil
}

func (s *PublicService) VerifyResetPassToken(ctx *gin.Context, token string) (*dto.ResetPassVerify, *errs.Errorf) {

	key := fmt.Sprintf("forgotpass:%s", token)
	ttl, err := s.RedisClient.TTL(ctx, key).Result()
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to get token ttl from redis : " + err.Error(),
		}
	}

	if ttl <= 0 {
		return nil, &errs.Errorf{
			Type:      errs.ErrTokenExpired,
			Message:   "Token has expired.",
			ReturnRaw: true,
		}
	}

	return &dto.ResetPassVerify{
		Valid: true,
		TTL:   int32(ttl.Seconds()),
	}, nil
}

// TODO: this does not invalidate already valid tokens, etc.
// It will be implemented with the actual {Change_Password/Invoke_Tokens} method.
func (s *PublicService) ResetPass(ctx *gin.Context, data *dto.ResetPassReq) *errs.Errorf {

	validity, errf := s.VerifyResetPassToken(ctx, data.Token)
	if errf != nil {
		return errf
	}
	if validity.TTL < 3 {
		return &errs.Errorf{
			Type:      errs.ErrTokenExpired,
			Message:   "Token has expired.",
			ReturnRaw: true,
		}
	}

	if data.NewPass != data.ConfPass {
		return &errs.Errorf{
			Type:      errs.ErrBadForm,
			Message:   "New password and confirmed password do not match.",
			ReturnRaw: true,
		}
	}
	instr, ok := utils.ValidatePassword(data.ConfPass)
	if !ok {
		return &errs.Errorf{
			Type:      errs.ErrBadForm,
			Message:   instr,
			ReturnRaw: true,
		}
	}

	key := fmt.Sprintf("forgotpass:%s", data.Token)
	userId, err := s.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to get from redis : " + err.Error(),
		}
	}
	userID, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to parse userid to int64 : " + err.Error(),
		}
	}

	currPass, err := s.Queries.GetPassForUserID(ctx, userID)
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to get user data from db : " + err.Error(),
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(currPass), []byte(data.ConfPass))
	if err == nil {
		return &errs.Errorf{
			Type:      errs.ErrConflict,
			Message:   "You cannot reuse a previous password.",
			ReturnRaw: true,
		}
	}
	if err != bcrypt.ErrMismatchedHashAndPassword {
		return &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to compare hash and pass : " + err.Error(),
		}
	}

	err = s.RedisClient.Del(ctx, key).Err()
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to del from redis : " + err.Error(),
		}
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(data.ConfPass), configs.Internal.BcryptPasswordCost)
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to generate hash for pass : " + err.Error(),
		}
	}

	err = s.Queries.UpdatePass(ctx, sqlc.UpdatePassParams{
		UserID:   userID,
		Password: string(pass),
	})
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to update pass in db : " + err.Error(),
		}
	}

	return nil
}
