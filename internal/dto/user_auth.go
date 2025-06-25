package dto

type NewOAuth struct {
	URL      *string `json:"url"`
	State    *string `json:"state"`
	StateTTL int64   `json:"stttl"`
}

type GoogleOAuth struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Id            string `json:"id"`
}

type EPassAuth struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type UserCreds struct {
	EPass  *EPassAuth   `json:"epass"`
	GOAuth *GoogleOAuth `json:"goauth"`
}

type AuthCheckRet struct {
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	Confirmed bool   `json:"confirmed"`
}

type GoogleOAuthCallback struct {
	State    string `json:"state"`
	Code     string `json:"code"`
	Scope    string `json:"scope"`
	AuthUser string `json:"authuser"`
	Prompt   string `json:"prompt"`

	CookieState string `json:"cookie_state"`
}

type ReqData struct {
	UserID int64
	UUID   string
}

type AuthRet struct {
	ATStr *string
	RTStr *string

	ATTTL int64
	RTTTL int64
}

type ForgotPassReq struct {
	Email string `json:"email"`
}

type ResetPassVerify struct {
	Valid bool  `json:"valid"`
	TTL   int32 `json:"ttl"`
}

type ResetPassReq struct {
	Token    string `json:"token"`
	NewPass  string `json:"newpass"`
	ConfPass string `json:"confpass"`
}
