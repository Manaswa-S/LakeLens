package dto

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
