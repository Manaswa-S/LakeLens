package auth

import (
	"context"
	"fmt"
	"lakelens/internal/consts/errs"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ATJWT struct {
	JWTID      string `json:"jid"`
	Issuer     string `json:"iss"`
	Subject    string `json:"sub"`
	IssuedAt   int64  `json:"iat"`
	ExpiresAt  int64  `json:"exp"`
	Role       string `json:"rol"`
	UserID     int64  `json:"rid"`
	UUID       string `json:"uid"`
	AuthMethod string `json:"amt"`
}

type RTJWT struct {
	JWTID      string `json:"jid"`
	Issuer     string `json:"iss"`
	Subject    string `json:"sub"`
	IssuedAt   int64  `json:"iat"`
	ExpiresAt  int64  `json:"exp"`
	Role       string `json:"rol"`
	UserID     int64  `json:"rid"`
	UUID       string `json:"uid"`
	AuthMethod string `json:"amt"`
	Version    int64  `json:"ver"`
	Rotation   int64  `json:"rot"`
}

type ATJWTParams struct {
	Issuer     string `json:"iss"`
	Subject    string `json:"sub"`
	Role       string `json:"rol"`
	UserID     int64  `json:"rid"`
	UUID       string `json:"uid"`
	AuthMethod string `json:"amt"`
}

type RTJWTParams struct {
	Issuer     string `json:"iss"`
	Subject    string `json:"sub"`
	Role       string `json:"rol"`
	UserID     int64  `json:"rid"`
	UUID       string `json:"uid"`
	AuthMethod string `json:"amt"`
}

type AuthServCreds struct {
	SigningKey string
	ATTTL      int64
	RTTTL      int64

	RefreshATIssuer string
	RefrestATSub    string

	AccAuthIssuer string
	AccAuthSub    string
}

type AuthService struct {
	Creds AuthServCreds

	RedisClient *redis.Client
	ctx         context.Context
}

// NewAuthService returns a AuthService instance.
// key is the secret jwt signing string.
// atttl and rtttl are TTL for AT and RT respectively, in seconds.
func NewAuthService(creds *AuthServCreds,
	redis *redis.Client) *AuthService {
	return &AuthService{
		Creds: *creds,

		RedisClient: redis,
		ctx:         context.Background(),
	}
}

// TODO: we are not refreshing RTs' anywhere ????

func (s *AuthService) RefreshAT(rtjwt *RTJWT) (string, *errs.Errorf) {

	key := fmt.Sprintf("signrt_ver_%d", rtjwt.UserID)
	currVer, err := s.RedisClient.Get(s.ctx, key).Result()
	if err != nil {
		return "", &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to get ver from redis : " + err.Error(),
		}
	}

	ver, err := strconv.ParseInt(currVer, 10, 64)
	if err != nil {
		return "", &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to parse ver to int64 : " + err.Error(),
		}
	}

	if rtjwt.Version != ver {
		return "", &errs.Errorf{
			Type:      errs.ErrForbidden,
			Message:   "Expired token. Version didn't match",
			ReturnRaw: true,
		}
	}

	// key = fmt.Sprintf("signrt_rot_%d", rtjwt.UserID)
	// currRot, err := s.RedisClient.Get(s.ctx, key).Result()
	// if err != nil {
	// 	return "", &errs.Errorf{
	// 		Type:    errs.ErrDependencyFailed,
	// 		Message: "Failed to get rot from redis : " + err.Error(),
	// 	}
	// }

	// rot, err := strconv.ParseInt(currRot, 10, 64)
	// if err != nil {
	// 	return "", &errs.Errorf{
	// 		Type:    errs.ErrInternalServer,
	// 		Message: "Failed to parse ver to int64 : " + err.Error(),
	// 	}
	// }

	// if rtjwt.Rotation != rot {
	// 	return "", &errs.Errorf{
	// 		Type:      errs.ErrForbidden,
	// 		Message:   "Expired token. Rotation didn't match",
	// 		ReturnRaw: true,
	// 	}
	// }

	issuer := "lakelens-refreshat"
	sub := "service:auth-account-refresh"

	t_acc, errf := s.SignAT(&ATJWTParams{
		Issuer:     issuer,
		Subject:    sub,
		Role:       rtjwt.Role,
		UserID:     rtjwt.UserID,
		UUID:       rtjwt.UUID,
		AuthMethod: rtjwt.AuthMethod,
	})
	if errf != nil {
		return "", errf
	}

	return t_acc, nil
}

func (s *AuthService) SignAT(atParams *ATJWTParams) (string, *errs.Errorf) {

	jti, err := s.getJWTId()
	if err != nil {
		return "", &errs.Errorf{}
	}
	issAt := time.Now().Unix()
	expAt := issAt + s.Creds.ATTTL

	key := fmt.Sprintf("signrt_rot_%d", atParams.UserID)
	_, err = s.RedisClient.Incr(s.ctx, key).Result()
	if err != nil {
		return "", &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to incr rotation : " + err.Error(),
		}
	}

	atData := ATJWT{
		JWTID:      jti,
		Issuer:     atParams.Issuer,
		Subject:    atParams.Subject,
		IssuedAt:   issAt,
		ExpiresAt:  expAt,
		Role:       atParams.Role,
		UserID:     atParams.UserID,
		UUID:       atParams.UUID,
		AuthMethod: atParams.AuthMethod,
	}

	// sign and return thats it.
	at_unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jid": atData.JWTID,
		"iss": atData.Issuer,
		"sub": atData.Subject,
		"iat": atData.IssuedAt,
		"exp": atData.ExpiresAt,
		"rol": atData.Role,
		"rid": atData.UserID,
		"uid": atData.UUID,
		"amt": atData.AuthMethod,
	})

	atoken, err := at_unsigned.SignedString([]byte(s.Creds.SigningKey))
	if err != nil {
		return "", &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to sign new access token : " + err.Error(),
		}
	}

	return atoken, nil
}

func (s *AuthService) SignRT(rtParams *RTJWTParams) (string, *errs.Errorf) {

	jti, err := s.getJWTId()
	if err != nil {
		return "", &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to get jwt id : " + err.Error(),
		}
	}
	issAt := time.Now().Unix()
	expAt := issAt + s.Creds.RTTTL

	// TODO: theres a massive bug here, suppose the redis db is flushed, and then server is restarted.
	// When a user visits the home page, the '/check/auth' or '/refresh/auth' fails here as the
	// 'signrt_ver_*' is not available. It is assumed to be set by the IncrBy command here, but in such a scenario it
	// is accessed before it was set.
	// It does not fail here but on redis.Get().
	// No brain has been stormed here.
	key := fmt.Sprintf("signrt_ver_%d", rtParams.UserID)
	ver, err := s.RedisClient.IncrBy(s.ctx, key, 1).Result()
	if err != nil {
		return "", &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to incr version : " + err.Error(),
		}
	}

	key = fmt.Sprintf("signrt_rot_%d", rtParams.UserID)
	_, err = s.RedisClient.Set(s.ctx, key, 1, 0).Result()
	if err != nil {
		return "", &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to set rotation : " + err.Error(),
		}
	}

	rtData := RTJWT{
		JWTID:      jti,
		Issuer:     rtParams.Issuer,
		Subject:    rtParams.Subject,
		IssuedAt:   issAt,
		ExpiresAt:  expAt,
		Role:       rtParams.Role,
		UserID:     rtParams.UserID,
		UUID:       rtParams.UUID,
		AuthMethod: rtParams.AuthMethod,
		Version:    ver,
		Rotation:   1,
	}

	// sign and return thats it.
	rt_unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jid": rtData.JWTID,
		"iss": rtData.Issuer,
		"sub": rtData.Subject,
		"iat": rtData.IssuedAt,
		"exp": rtData.ExpiresAt,
		"rol": rtData.Role,
		"rid": rtData.UserID,
		"uid": rtData.UUID,
		"amt": rtData.AuthMethod,
		"ver": rtData.Version,
		"rot": rtData.Rotation,
	})

	rtoken, err := rt_unsigned.SignedString([]byte(s.Creds.SigningKey))
	if err != nil {
		return "", &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to sign new refresh token : " + err.Error(),
		}
	}

	return rtoken, nil
}

func (s *AuthService) VerifyJWT(tokenStr string) (jwt.MapClaims, *errs.Errorf) {

	token, err := jwt.Parse(tokenStr,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method : %v", t.Method.Alg())
			}
			return []byte(s.Creds.SigningKey), nil
		},
	)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to parse jwt with claims : " + err.Error(),
		}
	}

	if !token.Valid {
		return nil, &errs.Errorf{
			Type:      errs.ErrBadRequest,
			Message:   "Invalid token.",
			ReturnRaw: true,
		}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, &errs.Errorf{
			Type:    errs.ErrBadForm,
			Message: "Bad format of claims in token.",
		}
	}

	return claims, nil
}

func (s *AuthService) ParseAT(claims jwt.MapClaims) (*ATJWT, *errs.Errorf) {

	atjwt := new(ATJWT)

	if jid, ok := claims["jid"]; ok {
		if jwtid, ok := jid.(string); ok {
			atjwt.JWTID = jwtid
		}
	}
	if iss, ok := claims["iss"]; ok {
		if issuer, ok := iss.(string); ok {
			atjwt.Issuer = issuer
		}
	}
	if sub, ok := claims["sub"]; ok {
		if subject, ok := sub.(string); ok {
			atjwt.Subject = subject
		}
	}
	if iat, ok := claims["iat"]; ok {
		if issuedat, ok := iat.(float64); ok {
			atjwt.IssuedAt = int64(issuedat)
		}
	}

	if exp, ok := claims["exp"]; ok {
		if expiresAt, ok := exp.(float64); ok {
			atjwt.ExpiresAt = int64(expiresAt)
		}
	}

	if rol, ok := claims["rol"]; ok {
		if role, ok := rol.(string); ok {
			atjwt.Role = role
		}
	}
	if rid, ok := claims["rid"]; ok {
		if userID, ok := rid.(float64); ok {
			atjwt.UserID = int64(userID)
		}
	}
	if uid, ok := claims["uid"]; ok {
		if uuid, ok := uid.(string); ok {
			atjwt.UUID = uuid
		}
	}
	if amt, ok := claims["amt"]; ok {
		if authMethod, ok := amt.(string); ok {
			atjwt.AuthMethod = authMethod
		}
	}

	return atjwt, nil
}

func (s *AuthService) ParseRT(claims jwt.MapClaims) (*RTJWT, *errs.Errorf) {

	rtjwt := new(RTJWT)

	if jid, ok := claims["jid"]; ok {
		if jwtid, ok := jid.(string); ok {
			rtjwt.JWTID = jwtid
		}
	}
	if iss, ok := claims["iss"]; ok {
		if issuer, ok := iss.(string); ok {
			rtjwt.Issuer = issuer
		}
	}
	if sub, ok := claims["sub"]; ok {
		if subject, ok := sub.(string); ok {
			rtjwt.Subject = subject
		}
	}
	if iat, ok := claims["iat"]; ok {
		if issuedat, ok := iat.(float64); ok {
			rtjwt.IssuedAt = int64(issuedat)
		}
	}

	if exp, ok := claims["exp"]; ok {
		if expiresAt, ok := exp.(float64); ok {
			rtjwt.ExpiresAt = int64(expiresAt)
		}
	}

	if rol, ok := claims["rol"]; ok {
		if role, ok := rol.(string); ok {
			rtjwt.Role = role
		}
	}
	if rid, ok := claims["rid"]; ok {
		if userID, ok := rid.(float64); ok {
			rtjwt.UserID = int64(userID)
		}
	}
	if uid, ok := claims["uid"]; ok {
		if uuid, ok := uid.(string); ok {
			rtjwt.UUID = uuid
		}
	}
	if amt, ok := claims["amt"]; ok {
		if authMethod, ok := amt.(string); ok {
			rtjwt.AuthMethod = authMethod
		}
	}
	if ver, ok := claims["ver"]; ok {
		if version, ok := ver.(float64); ok {
			rtjwt.Version = int64(version)
		}
	}
	if rot, ok := claims["rot"]; ok {
		if rotation, ok := rot.(float64); ok {
			rtjwt.Rotation = int64(rotation)
		}
	}

	return rtjwt, nil
}

func (s *AuthService) getJWTId() (string, error) {
	uid, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}
