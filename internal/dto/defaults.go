package dto

import (
	"time"
)

// These are structs that are intended to be put out to the user.

type RequestLogData struct {
	StartTime     string `json:"starttime"`
	ClientIP      string `json:"clientip"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	StatusCode    int    `json:"statuscode"`
	InternalError string `json:"internalerror"`
	Latency       int64  `json:"latency"` // in milliseconds
}

type TipRespHRef struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

type TipResp struct {
	Tip   string
	HRefs map[string]TipRespHRef
}

type RecentsResp struct {
	Action      string
	Description string
	Time        time.Time
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
	Link  string
}

type FeatureTour struct {
	LastTour int32
}

type LakeDetails struct {
	Details   *LakeResp
	Locations []*LocResp
}

type LakeFileDistStats struct {
	TotalSize int64
	FileCount int64
}

type LakeFileDist struct {
	Dist map[string]*LakeFileDistStats
}

type LocCheckResp struct {
	LocID       int64
	BucketName  string
	AuthCheck   bool
	PolicyCheck bool
	ReadCheck   bool
	WriteCheck  bool
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
	LakeID    int64
	Locations []Locations
}

type AddLocsReq struct {
	LakeID   int64
	LocNames []string
}

type AddLocsResp struct {
	Failed []string
	Added  []string
}

type Locations struct { // use of this is discouraged. use LocResp instead.
	Name         *string
	CreationDate *time.Time
	Region       *string
	Registered   bool // is set to true if location is already registered.
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

type OverviewData struct {
	FoundAt     string // the uri where the table was found.
	Location    string
	TableUUID   string
	FilesReadMp map[string]int64  // the kind and number of files read.
	TableType   string            // the table type that was detected.
	FileURIs    map[string]string // the uris' of things like metadata files, snapshots, etc.
}

type OverviewStatsTable struct {
	TableType    string // the table type, iceberg/delta/etc
	TableVersion any    // the version of the table
	TableSpecs   string // the specs of the table
}

type OverviewStatsVersion struct {
	CurrentVersion string
	LastSnapshot   int64
	TotalSnapshots int64
}

type OverviewStatsRowCount struct {
	TotalCount string
	DeltaCount int64
}

type OverviewStatsStorage struct {
	TotalSize      int64
	TotalDataFiles int64
	AvgFileSize    int64
}

type OverviewStats struct {
	Table   OverviewStatsTable
	Version OverviewStatsVersion
	Rows    OverviewStatsRowCount
	Storage OverviewStatsStorage
}

type OverviewSchemaField struct {
	ID       int64
	Name     string
	Required bool
	Type     string
}

type OverviewSchema struct {
	SchemaID int64
	Fields   []*OverviewSchemaField
}

type OverviewPartitionField struct {
	Name      string
	Transform string
	SourceID  int64
	FieldID   int64
}

type OverviewPartition struct {
	DefaultSpecID int64
	Fields        []*OverviewPartitionField
}

type OverviewSnapshot struct {
	SequenceNumber int64
	LastOperation  string
	SchemaID       int64
	ManifestList   string
}

type OverviewGraphs struct {
	TimeStampMS      int64
	TotalRecords     string
	TotalFileSize    string
	TotalDataFiles   string
	TotalDeleteFiles string
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

type SchemaListData struct {
	SchemaID            int64
	FromTimeStampMS     int64  // timestamp from when the schema was applied
	ValidUptoSnapshotID string // snapshot id until when valid
}

type SchemaList struct {
	List map[int64]*SchemaListData
}

type SchemaField struct {
	ID       int64
	Name     string
	Required bool
	Type     string
}

type Schema struct {
	SchemaID int64
	Fields   []*SchemaField
}

type SchemaData struct {
	LastUpdatedMS    int64
	LastColumnID     int64
	CurrentSchemaID  int64
	SchemaType       string
	SchemaID         int64
	RelatedSnapshots []int64
	ColumnSizes      map[string]int64 // column name to size in bytes
}

type ColSize struct {
	ID            int64
	Name          string
	Size          int64
	NullCount     int64
	ValueCount    int64
	AvgSizePerVal float32
}

type SchemaColSizes struct {
	SchemaID int64
	ColSizes []ColSize
}
