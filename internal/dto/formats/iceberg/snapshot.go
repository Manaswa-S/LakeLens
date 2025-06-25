package formats

// Main structs used to unmarshal the snap-*.avro files of iceberg tables.
// These are different from the snapshots found in .metadata.json files, though both contain the same data.

type SnapshotRecord struct {
	SequenceNumber         int64                  `avro:"sequence_number"`
	AddedDataFilesCount    int32                  `avro:"added_data_files_count"`
	AddedRowsCount         int64                  `avro:"added_rows_count"`
	AddedSnapshotID        int64                  `avro:"added_snapshot_id"`
	Content                int32                  `avro:"content"`
	DeletedDataFilesCount  int32                  `avro:"deleted_data_files_count"`
	DeletedRowsCount       int64                  `avro:"deleted_rows_count"`
	ExistingDataFilesCount int32                  `avro:"existing_data_files_count"`
	ExistingRowsCount      int64                  `avro:"existing_rows_count"`
	ManifestLength         int64                  `avro:"manifest_length"`
	ManifestPath           string                 `avro:"manifest_path"`
	MinSequenceNumber      int64                  `avro:"min_sequence_number"`
	Partitions             map[string]interface{} `avro:"partitions"`
	PartitionSpecID        int32                  `avro:"partition_spec_id"`
}
