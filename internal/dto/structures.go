package dto

import (
	"lakelens/internal/consts/errs"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Internal

type S3ClientSave struct {
	LastUsed time.Time
	S3Client *s3.Client
}

type NewOAuth struct {
	URL      *string `json:"url"`
	State    *string `json:"state"`
	StateTTL int64   `json:"stttl"`
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

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// User Requests
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

type NewLake struct {
	Name string // the lake project name, whatever the user wants.

	// only one is valid, others remain nil.
	S3    *NewLakeS3
	Azure *NewLakeAzure
	GCP   *NewLakeGCP
}

type NewLakeS3 struct {
	AccessID   string
	AccessKey  string
	LakeRegion string
}

type NewLakeAzure struct {
	// TODO:
}

type NewLakeGCP struct {
	// TODO:
}


type NewLakeResp struct {
	Name *string
	CreationDate *time.Time
	Region *string
}


type AddLocsReq struct {
	LakeName string
	LocNames []string
}



// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// User Responses

type BucketData struct {
	Name         string
	StorageType  string
	Region       *string
	CreationDate *time.Time
	TableType    string
	UpdatedAt    time.Time
	//
	KeyCount int64
	//
	LocationID int64
}

type NewBucket struct {
	Data    BucketData
	Parquet IsParquet
	Iceberg IsIceberg
	Delta   IsDelta
	Hudi    IsHudi
	Errors  []*errs.Errorf
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Internal Operations

type Latency struct {
	Start              int64
	ListBuckets        int64
	DetermineTableType int64
	Handle             int64
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

type AuthCheckRet struct {
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	Confirmed bool   `json:"confirmed"`
}
