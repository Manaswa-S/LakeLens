package dto

import "time"

type RequestLogData struct {
	StartTime     string `json:"starttime"`
	ClientIP      string `json:"clientip"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	StatusCode    int    `json:"statuscode"`
	InternalError string `json:"internalerror"`
	Latency       int64  `json:"latency"` // in milliseconds
}

type LakeResp struct {
	LakeID    int64
	Name      string
	Ptype     string
	CreatedAt time.Time
	Region    string
}

type LocResp struct {
	LocID      int64
	LakeID     int64
	BucketName string
	CreatedAt  time.Time
}

type AccDetailsResp struct {
	Email     string
	CreatedAt time.Time
	Confirmed bool
	UUID      string

	Name    string
	Picture string

	AuthType string
}

type AccBillingResp struct {
	Type       string
	Applicable bool
	NextPay    time.Time
}

type LocsForLake struct {
	Lake LakeResp
	Locs []LocResp
}

type AccProjectsResp struct {
	LocsForLake []*LocsForLake
}

type AccSettingsResp struct {
	// Preferences
	AdvancedMetaData    bool
	CompactView         bool
	AutoRefreshInterval int32
	Notifications       bool
	// UI/UX
	Theme        string
	FontSize     int32
	ShowToolTips bool
	// Usage
	KeyboardShortcuts bool
}



// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>


type SearchChoice struct {
	Label string
	Link string
}