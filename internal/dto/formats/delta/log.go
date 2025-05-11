package formats


type DeltaLog struct {
	CommitInfo DeltaCommitInfo `json:"commitInfo"`
	Protocol DeltaProtocol `json:"protocol"`
	Metadata DeltaMetadata `json:"metaData"`
	Add []DeltaAdd `json:"add"`
	Remove []DeltaRemove `json:"remove"`
	Txn DeltaTxn `json:"txn"`
}

type DeltaLogSingle struct {
	CommitInfo *DeltaCommitInfo `json:"commitInfo"`
	Protocol *DeltaProtocol `json:"protocol"`
	Metadata *DeltaMetadata `json:"metaData"`
	Add *DeltaAdd `json:"add"`
	Remove *DeltaRemove `json:"remove"`
	Txn *DeltaTxn `json:"txn"`
}


type DeltaCommitInfo struct {
	Timestamp int64 `json:"timestamp"`
	UserID string `json:"userId"`
	UserName string `json:"userName"`
	Operation string `json:"operation"`
	OperationParams DeltaOperationParams `json:"operationParameters"`
	Notebook DeltaNotebook `json:"notebook"`
	ClusterID string `json:"clusterId"`
	IsolationLvl string `json:"isolationLevel"`
	IsBlindAppend bool `json:"isBlindAppend"`
	OperationMetrics DeltaOperationMetrics `json:"operationMetrics"`
	EngineInfo string `json:"engineInfo"`
	TxnID string `json:"txnId"`
}
type DeltaOperationParams struct {
	Mode string `json:"mode"`
	PartitionBy string `json:"partitionBy"`
}
type DeltaNotebook struct {
	NotebookID string `json:"notebookId"`
	NotebookPath string `json:"notebookPath"`
	ClusterID string `json:"clusterId"`
}
type DeltaOperationMetrics struct {
	NumFiles string `json:"numFiles"`
	NumOutputRows string `json:"numOutputRows"`
	NumOutputBytes string `json:"numOutputBytes"`
}


type DeltaTxn struct {
	AppID string `json:"appId"`
	Version int64 `json:"version"`
}


type DeltaProtocol struct {
	MinReaderVersion int64 `json:"minReaderVersion"`
	MinWriterVersion int64 `json:"minWriterVersion"`
	ReaderVersion int64 `json:"readerVersion"`
    WriterVersion int64 `json:"writerVersion"`
    ReaderFeatures []string `json:"readerFeatures"`
    WriterFeatures []string `json:"writerFeatures"`
}


type DeltaMetadata struct {
	ID string `json:"id"`
	Format DeltaFormat `json:"format"`
	SchemaString string `json:"schemaString"`
	Schema DeltaSchema `json:"schema"`
	PartitionColumns []any `json:"partitionColumns"`
	Configuration DeltaConfiguration `json:"configuration"`
	CreatedTime int64 `json:"createdTime"`
}
type DeltaFormat struct {
	Provider string `json:"provider"`
	Options any `json:"options"`
}
type DeltaSchema struct {
	Type string `json:"type"`
	Fields []DeltaSchemaField `json:"fields"`
}
type DeltaSchemaField struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Nullable bool `json:"nullable"`
	Metadata any `json:"metadata"`
}

type DeltaConfiguration struct {
	AppendOnly bool `json:"appendOnly"`
}


type DeltaAdd struct {
	Path string `json:"path"`
	PartitionValues any `json:"partitionValues"`
	Size int64 `json:"size"`
	ModificationTime int64 `json:"modificationTime"`
	DataChange bool `json:"dataChange"`
	BaseRowID int64 `json:"baseRowId"`
	DefaultRowCommitVersion int64 `json:"defaultRowCommitVersion"`
	ClusteringProvider string `json:"clusteringProvider"`
	Stats string `json:"stats"`
	Tags DeltaTags `json:"tags"`
}
type DeltaTags struct {
	InsertionTime string `json:"INSERTION_TIME"`
	MinInsertionTime string `json:"MIN_INSERTION_TIME"`
	MaxInsertionTime string `json:"MAX_INSERTION_TIME"`
	OptimizeTargetSize string `json:"OPTIMIZE_TARGET_SIZE"`
}


type DeltaRemove struct {
	Path string `json:"path"`
	DeletionTimestamp int64 `json:"deletionTimestamp"`
	BaseRowID int64 `json:"baseRowId"`
	DefaultRowCommitVersion int64 `json:"defaultRowCommitVersion"`
	DataChange bool `json:"dataChange"`
}