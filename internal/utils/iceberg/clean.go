package iceutils

import (
	"encoding/json"
	"fmt"
	formats "lakelens/internal/dto/formats/iceberg"
)

func CleanManifestEntry(entriesMap []map[string]any) []formats.ManifestEntry {

	records := make([]formats.ManifestEntry, 0)

	for _, entryMap := range entriesMap {

		dataFileMap, ok := entryMap["data_file"].(map[string]any)
		if !ok {
			continue
		}
		manifestEntry := formats.ManifestEntry{
			FileSequenceNumber: toNullableInt64(entryMap["file_sequence_number"]),
			SequenceNumber:     toNullableInt64(entryMap["sequence_number"]),
			SnapshotID:         entryMap["snapshot_id"].(map[string]any),
			Status:             int(entryMap["status"].(int32)),
			DataFile: formats.ManifestDataFile{
				FilePath:        dataFileMap["file_path"].(string),
				FileFormat:      dataFileMap["file_format"].(string),
				RecordCount:     dataFileMap["record_count"].(int64),
				FileSizeInBytes: dataFileMap["file_size_in_bytes"].(int64),
				Partition:       dataFileMap["partition"].(map[string]any),
				ColumnSizes:     dataFileMap["column_sizes"].(map[string]any),
				Content:         dataFileMap["content"],
				UpperBounds:     dataFileMap["upper_bounds"].(map[string]any),
				LowerBounds:     dataFileMap["lower_bounds"].(map[string]any),
				ValueCounts:     dataFileMap["value_counts"].(map[string]any),
				NullValueCounts: dataFileMap["null_value_counts"].(map[string]any),
				NANValueCounts:  dataFileMap["nan_value_counts"].(map[string]any),
				// Add other fields like column stats, etc. if you want
			},
		}

		records = append(records, manifestEntry)
	}

	return records
}

func CleanManifestMetadata(entriesMap map[string][]byte) formats.ManifestMetadata {

	var metadata formats.ManifestMetadata

	metadata.AvroCodec = string(entriesMap["avro.codec"])
	metadata.FormatVersion = string(entriesMap["format-version"])
	metadata.PartitionSpecID = string(entriesMap["partition-spec-id"])
	metadata.Content = string(entriesMap["content"])

	var partitionSpec []formats.IcebergPartitionSpecField
	err := json.Unmarshal(entriesMap["partition-spec"], &partitionSpec)
	if err != nil {
		fmt.Println(err)
	}
	metadata.PartitionSpec = partitionSpec

	var icebergSchema formats.ManifestMetadataIcebergSchema
	err = json.Unmarshal(entriesMap["iceberg.schema"], &icebergSchema)
	if err != nil {
		fmt.Println(err)
	}
	metadata.IcebergSchema = icebergSchema

	var schema formats.IcebergSchema
	err = json.Unmarshal(entriesMap["schema"], &schema)
	if err != nil {
		fmt.Println(err)
	}
	metadata.Schema = schema

	var avroSchema map[string]any
	err = json.Unmarshal(entriesMap["avro.schema"], &avroSchema)
	if err != nil {
		fmt.Println(err)
	}
	metadata.AvroSchema = avroSchema

	return metadata
}
func CleanSnapshot(recordMaps []map[string]any) []*formats.SnapshotRecord {

	records := make([]*formats.SnapshotRecord, 0)

	for _, recordMap := range recordMaps {

		snap := formats.SnapshotRecord{
			AddedDataFilesCount:    recordMap["added_data_files_count"].(int32),
			AddedRowsCount:         recordMap["added_rows_count"].(int64),
			AddedSnapshotID:        recordMap["added_snapshot_id"].(int64),
			Content:                recordMap["content"].(int32),
			DeletedDataFilesCount:  recordMap["deleted_data_files_count"].(int32),
			DeletedRowsCount:       recordMap["deleted_rows_count"].(int64),
			ExistingDataFilesCount: recordMap["existing_data_files_count"].(int32),
			ExistingRowsCount:      recordMap["existing_rows_count"].(int64),
			ManifestLength:         recordMap["manifest_length"].(int64),
			ManifestPath:           recordMap["manifest_path"].(string),
			MinSequenceNumber:      recordMap["min_sequence_number"].(int64),
			Partitions:             recordMap["partitions"].(map[string]any),
			PartitionSpecID:        recordMap["partition_spec_id"].(int32),
			SequenceNumber:         recordMap["sequence_number"].(int64),
		}

		records = append(records, &snap)
	}

	return records
}

func toNullableInt64(v any) *int64 {
	if v == nil {
		return nil
	}
	num := v.(int64)
	return &num
}
