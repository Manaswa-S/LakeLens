package iceutils

import (
	formats "lakelens/internal/dto/formats/iceberg"
)

func CleanManifest(recordMaps []map[string]any) []*formats.ManifestEntry {

	records := make([]*formats.ManifestEntry, 0)

	for _, recordMap := range recordMaps {

		dataFileMap, ok := recordMap["data_file"].(map[string]any)
		if !ok {
			continue
		}
		manifestEntry := formats.ManifestEntry{
			FileSequenceNumber: toNullableInt64(recordMap["file_sequence_number"]),
			SequenceNumber:     toNullableInt64(recordMap["sequence_number"]),
			SnapshotID:         recordMap["snapshot_id"].(map[string]any),
			Status:             int(recordMap["status"].(int32)),
			DataFile: formats.ManifestDataFile{
				FilePath:        dataFileMap["file_path"].(string),
				FileFormat:      dataFileMap["file_format"].(string),
				RecordCount:     dataFileMap["record_count"].(int64),
				FileSizeInBytes: dataFileMap["file_size_in_bytes"].(int64),
				Partition:       dataFileMap["partition"].(map[string]any),
				// Add other fields like column stats, etc. if you want
			},
		}

		records = append(records, &manifestEntry)
	}

	return records
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
