package formats

// Main structs used to unmarshal the m0.avro manifest files of iceberg tables.
// These are different from the manifest data found in .metadata.json files.

type IcebergManifest struct {
	Data []*ManifestData
}

type ManifestData struct {
	URI      string           `json:"uri"`
	Metadata ManifestMetadata `json:"metadata"`
	Entries  []ManifestEntry  `json:"entries"`
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

type ManifestEntry struct {
	FileSequenceNumber *int64
	SequenceNumber     *int64
	SnapshotID         map[string]any
	Status             int
	DataFile           ManifestDataFile
}

type ManifestDataFile struct {
	Content         any `json:"content"`
	FilePath        string
	FileFormat      string
	RecordCount     int64
	FileSizeInBytes int64
	ColumnSizes     map[string]any `json:"column_sizes"`
	ValueCounts     map[string]any `json:"value_counts"`
	NullValueCounts map[string]any `json:"null_value_counts"`
	NANValueCounts  map[string]any `json:"nan_value_counts"`
	// TODO: these fields are intensive to decode just for a few use cases. let's wait and see if there arise more uses and then we can implement it all.
	// Use Cases:
	// 1) For lower/upper bounds in schema tables.
	LowerBounds  map[string]any `json:"lower_bounds"`
	UpperBounds  map[string]any `json:"upper_bounds"`
	KeyMetadata  string         `json:"key_metadata"`
	SplitOffsets map[string]any `json:"split_offsets"`
	EqualityIDs  map[string]any `json:"equality_ids"`
	SortOrderID  string         `json:"sort_order_id"`

	Partition map[string]any
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

type ManifestMetadata struct {
	AvroCodec       string                        `json:"avro.codec"`
	FormatVersion   string                        `json:"format-version"`
	PartitionSpecID string                        `json:"partition-spec-id"`
	IcebergSchema   ManifestMetadataIcebergSchema `json:"iceberg.schema"`
	PartitionSpec   []IcebergPartitionSpecField   `json:"partition-spec"`
	Content         string                        `json:"content"`
	Schema          IcebergSchema                 `json:"schema"`
	AvroSchema      map[string]any                `json:"avro.schema"`
}

type ManifestMetadataIcebergSchema struct {
	Type     string `json:"type"`
	SchemaID int64  `json:"schema-id"`
	Fields   []ManifestMetadataIcebergSchemaField
}

type ManifestMetadataIcebergSchemaField struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Type     any    `json:"type"`
	Doc      string `json:"doc"`
}
