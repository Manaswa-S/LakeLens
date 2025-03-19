package formats

type IcebergData struct {
	FormatVersion int64 `json:"format-version"`
	TableUUID string `json:"table-uuid"`
	Location string `json:"location"`
	LastSequenceNumber int64 `json:"last-sequence-number"`
	LastUpdatedMS int64 `json:"last-updated-ms"`
	LastColumnID int64 `json:"last-column-id"`
	CurrentSchemaID int64 `json:"current-schema-id"`
	Schemas []IcebergSchema `json:"schemas"`
	DefaultSpecID int64 `json:"default-spec-id"`
	PartitionSpecs []IcebergPartitionSpec `json:"partition-specs"`
	LastPartitionID int64 `json:"last-partition-id"`
	DefaultSortOrderID int64 `json:"default-sort-order-id"`
	SortOrders []IcebergSortOrder `json:"sort-orders"`
	Properties IcebergProperties `json:"properties"`
	CurrentSnapshotID int64 `json:"current-snapshot-id"`
	Refs IcebergRefs `json:"refs"`
	Snapshots []IcebergSnapshot `json:"snapshots"`
	Statistics []IcebergStatistics `json:"statistics"`
	SnapshotLog []IcebergSnapshotLog `json:"snapshot-log"`
	MetadataLog []IcebergMetadataLog `json:"metadata-log"`
}

type IcebergSchema struct {
	Type string `json:"type"`
	SchemaID int64 `json:"schema-id"`
	Fields []IcebergSchemaField `json:"fields"`
}
type IcebergSchemaField struct {
	ID int64 `json:"id"`
	Name string `json:"name"`
	Required bool `json:"required"`
	Type string `json:"type"`
}

type IcebergPartitionSpec struct {
	SpecID int64 `json:"spec-id"`
	Fields []IcebergPartitionSpecField `json:"fields"`
}
type IcebergPartitionSpecField struct {
	Name string `json:"name"`
	Transform string `json:"transform"`
	SourceID int64 `json:"source-id"`
	FieldID int64 `json:"field-id"`
}

type IcebergSortOrder struct {
	OrderID int64 `json:"order-id"`
	Fields []IcebergSortOrderField `json:"fields"`
}
type IcebergSortOrderField struct {

}

type IcebergProperties struct {
	Write_ObjectStorage_Enabled string `json:"write.object-storage.enabled"`
	Write_ObjectStorage_Path string `json:"write.object-storage.path"`
	Write_Parquet_CompressionCodec string `json:"write.parquet.compression-codec"`
}

type IcebergRefs struct {
	Main IcebergRefsMain `json:"main"`
}
type IcebergRefsMain struct {
	SnapshotID int64 `json:"snapshot-id"`
	Type string `json:"type"`
}

type IcebergSnapshot struct {
	SequenceNumber int64 `json:"sequence-number"`
	SnapshotID int64 `json:"snapshot-id"`
	TimestampMS int64 `json:"timestamp-ms"`
	Summary IcebergSnapshotSummary `json:"summary"`
	ManifestList string `json:"manifest-list"`
	SchemaID int64 `json:"schema-id"`
}
type IcebergSnapshotSummary struct {
	Operation string `json:"operation"`
	TrinoQueryID string `json:"trino_query_id"`
	AddedDataFiles string `json:"added-data-files"`
	AddedRecords string `json:"added-records"`
	AddedFilesSize string `json:"added-files-size"`
	ChangedPartitionCount string `json:"changed-partition-count"`
	TotalRecords string `json:"total-records"`
	TotalFilesSize string `json:"total-files-size"`
	TotalDataFiles string `json:"total-data-files"`
	TotalDeleteFiles string `json:"total-delete-files"`
	TotalPositionDeletes string `json:"total-position-deletes"`
	TotalEqualityDeletes string `json:"total-equality-deletes"`
}

type IcebergStatistics struct {

}

type IcebergSnapshotLog struct {
	TimestampMS int64 `json:"timestampms"`
	SnapshotID int64 `json:"snapshot-id"`
}

type IcebergMetadataLog struct {
	TimestampMS int64 `json:"timestamp-ms"`
	MetadataFile string `json:"metadata-file"`
}

