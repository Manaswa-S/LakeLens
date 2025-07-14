package tracelog

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type RouteInfo struct {
	ActionID   int64    // unique id assigned to the action, helps identify and process.
	Action     string   // the '_' separate action name, should be constant across states.
	Params     []string // the params to pull from the url.
	BodyFields []string // the fields to pull from the request body.
}

// Only routes mentioned here will be traced and logged.
// Request.URL.Path : RouteInfo{}
var RoutesList = map[string]RouteInfo{
	"/lens/manager/newlake": {
		ActionID:   1001,
		Action:     "new_lake",
		BodyFields: []string{"new_lake_name"},
	},
	"/lens/iceberg/overview/data/:locid": {
		ActionID: 1002,
		Action:   "overview_data",
		Params:   []string{"locid"},
	},
	"/lens/manager/analyze/loc/:locid": {
		ActionID: 1003,
		Action:   "analyze_loc",
		Params:   []string{"locid"},
	},
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

type ActionData struct {
	Action string            // the action performed
	Data   map[string]string // the key value map for any required data.
}

type UserAction struct {
	UserID      int64  // the id of the user
	ActionID    int64  // the action id assigned
	Time        int64  // the unix seconds at the time of the action.
	Title       string // the main title shown to the user
	Description string // the main description shown to the user
	Store       ActionData
}

type TriggerType int

const (
	TimedTrigger TriggerType = iota
	CapacityTrigger
)

type SaveToDBData struct {
	UserIDs      []int64
	Times        []pgtype.Timestamptz
	ActionIDs    []int64
	Actions      [][]byte
	Titles       []string
	Descriptions []string
}
