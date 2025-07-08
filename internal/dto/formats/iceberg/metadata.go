package formats

// Main structs used to unmarshal the .metadata.json files of iceberg tables.

type IcebergMetadata struct {
	FormatVersion       int64                        `json:"format-version"`
	TableUUID           string                       `json:"table-uuid"`
	Location            string                       `json:"location"`
	LastSequenceNumber  int64                        `json:"last-sequence-number"`
	LastUpdatedMS       int64                        `json:"last-updated-ms"`
	LastColumnID        int64                        `json:"last-column-id"`
	CurrentSchemaID     int64                        `json:"current-schema-id"`
	Schemas             []IcebergSchema              `json:"schemas"`
	DefaultSpecID       int64                        `json:"default-spec-id"`
	PartitionSpecs      []IcebergPartitionSpec       `json:"partition-specs"`
	LastPartitionID     int64                        `json:"last-partition-id"`
	DefaultSortOrderID  int64                        `json:"default-sort-order-id"`
	SortOrders          []IcebergSortOrder           `json:"sort-orders"`
	Properties          IcebergProperties            `json:"properties"`
	CurrentSnapshotID   int64                        `json:"current-snapshot-id"`
	Refs                IcebergRefs                  `json:"refs"`
	Snapshots           []IcebergMetadataSnapshot    `json:"snapshots"`
	Statistics          []IcebergStatistics          `json:"statistics"`
	SnapshotLog         []IcebergSnapshotLog         `json:"snapshot-log"`
	MetadataLog         []IcebergMetadataLog         `json:"metadata-log"`
	PartitionStatistics []IcebergPartitionStatistics `json:"partition-statistics"`
}

type IcebergSchema struct {
	Type     string               `json:"type"`
	SchemaID int64                `json:"schema-id"`
	Fields   []IcebergSchemaField `json:"fields"`
}
type IcebergSchemaField struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Type     string `json:"type"`
}

type IcebergPartitionSpec struct {
	SpecID int64                       `json:"spec-id"`
	Fields []IcebergPartitionSpecField `json:"fields"`
}
type IcebergPartitionSpecField struct {
	Name      string `json:"name"`
	Transform string `json:"transform"`
	SourceID  int64  `json:"source-id"`
	FieldID   int64  `json:"field-id"`
}

type IcebergSortOrder struct {
	OrderID int64                   `json:"order-id"`
	Fields  []IcebergSortOrderField `json:"fields"`
}
type IcebergSortOrderField struct {
}

type IcebergProperties struct {
	Write_ObjectStorage_Enabled    string `json:"write.object-storage.enabled"`
	Write_ObjectStorage_Path       string `json:"write.object-storage.path"`
	Write_Parquet_CompressionCodec string `json:"write.parquet.compression-codec"`
}

type IcebergRefs struct {
	Main IcebergRefsMain `json:"main"`
}
type IcebergRefsMain struct {
	SnapshotID int64  `json:"snapshot-id"`
	Type       string `json:"type"`
}

type IcebergMetadataSnapshot struct {
	SequenceNumber int64                  `json:"sequence-number"`
	SnapshotID     int64                  `json:"snapshot-id"`
	TimestampMS    int64                  `json:"timestamp-ms"`
	Summary        IcebergSnapshotSummary `json:"summary"`
	ManifestList   string                 `json:"manifest-list"`
	SchemaID       int64                  `json:"schema-id"`
}
type IcebergSnapshotSummary struct {
	Operation                  string `json:"operation"`
	TrinoQueryID               string `json:"trino_query_id"`
	AddedDataFiles             string `json:"added-data-files"`
	DeletedDataFiles           string `json:"deleted-data-files"`
	AddedDeleteFiles           string `json:"added-delete-files"`
	AddedEqualityDeleteFiles   string `json:"added-equality-delete-files"`
	RemovedEqualityDeleteFiles string `json:"removed-equality-delete-files"`
	AddedPositionDeleteFiles   string `json:"added-position-delete-files"`
	AddedDvs                   string `json:"added-dvs"`
	RemovedDvs                 string `json:"removed-dvs"`
	RemovedDeleteFiles         string `json:"removed-delete-files"`
	AddedPositionDeletes       string `json:"added-position-deletes"`
	RemovedPositionDeletes     string `json:"removed-position-deletes"`
	AddedEqualityDeletes       string `json:"added-equality-deletes"`
	RemovedEqualityDeletes     string `json:"removed-equality-deletes"`
	DeletedDuplicateFiles      string `json:"deleted-duplicate-files"`
	WapID                      string `json:"wap.id"`
	SourceSnapshotID           string `json:"source-snapshot-id"`
	EngineName                 string `json:"engine-name"`
	EngineVersion              string `json:"engine-version"`
	DeletedRecords             string `json:"deleted-records"`
	AddedRecords               string `json:"added-records"`
	AddedFilesSize             string `json:"added-files-size"`
	RemovedFilesSize           string `json:"removed-files-size"`
	ChangedPartitionCount      string `json:"changed-partition-count"`
	TotalRecords               string `json:"total-records"`
	TotalFilesSize             string `json:"total-files-size"`
	TotalDataFiles             string `json:"total-data-files"`
	TotalDeleteFiles           string `json:"total-delete-files"`
	TotalPositionDeletes       string `json:"total-position-deletes"`
	TotalEqualityDeletes       string `json:"total-equality-deletes"`
}

type IcebergStatistics struct {
	SnapshotID            string                          `json:"snapshot-id"`
	StatisticsPath        string                          `json:"statistics-path"`
	FileSizeInBytes       int64                           `json:"file-size-in-bytes"`
	FileFooterSizeInBytes int64                           `json:"file-footer-size-in-bytes"`
	BlobMetadata          []IcebergStatisticsBlobMetadata `json:"blob-metadata"`
}
type IcebergStatisticsBlobMetadata struct {
	Type           string                                  `json:"type"`
	SnapshotID     int64                                   `json:"snapshot-id"`
	SequenceNumber int64                                   `json:"sequence-number"`
	Fields         []int64                                 `json:"fields"`
	Properties     IcebergStatisticsBlobMetadataProperties `json:"properties"`
}
type IcebergStatisticsBlobMetadataProperties struct {
	StatisticType string `json:"statistic-type"`
}

type IcebergSnapshotLog struct {
	TimestampMS int64 `json:"timestampms"`
	SnapshotID  int64 `json:"snapshot-id"`
}

type IcebergMetadataLog struct {
	TimestampMS  int64  `json:"timestamp-ms"`
	MetadataFile string `json:"metadata-file"`
}

type IcebergPartitionStatistics struct {
	SnapshotID      int64  `json:"snapshot-id"`
	StatisticsPath  string `json:"statistics-path"`
	FileSizeInBytes int64  `json:"file-size-in-bytes"`
}
