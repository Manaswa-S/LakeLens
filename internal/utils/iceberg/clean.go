package iceutils

import (
	"lakelens/internal/dto"
	formats "lakelens/internal/dto/formats/iceberg"
)

// TODO: clean and validate first
func CleanMetadata(iceberg *formats.IcebergMetadataData) (*dto.IcebergMetadata) {

	cleanIceberg := new(dto.IcebergMetadata)

	cleanIceberg.TableUUID = iceberg.TableUUID
	cleanIceberg.Location = iceberg.Location
	cleanIceberg.LastUpdatedMS = iceberg.LastUpdatedMS

	cleanIceberg.CurrentSchemaID = iceberg.CurrentSchemaID
	cleanIceberg.Schemas = iceberg.Schemas

	cleanIceberg.CurrentSnapshotID = iceberg.CurrentSnapshotID
	cleanIceberg.Snapshots = iceberg.Snapshots
	cleanIceberg.SnapshotLog = iceberg.SnapshotLog

	cleanIceberg.MetadataLog = iceberg.MetadataLog

	cleanIceberg.DefaultSpecID = iceberg.DefaultSpecID
	cleanIceberg.PartitionSpecs = iceberg.PartitionSpecs
	cleanIceberg.LastPartitonID = iceberg.LastPartitionID

	return cleanIceberg
}

func CleanManifest(recordMaps []map[string]any) (*dto.IcebergManifest) {

	records := new(dto.IcebergManifest)

	for _, recordMap := range recordMaps {

		dataFileMap, ok := recordMap["data_file"].(map[string]any)
		if !ok {
			continue
			// return nil, &errs.Errorf{
			// 	Type: errs.ErrDependencyFailed,
			// 	Message: "Avro ocf datum either doesn't have 'data_file' field or it isn't of expected 'map[string]any' type.",
			// }
		}
		manifestEntry := formats.ManifestEntry{
			FileSequenceNumber: toNullableInt64(recordMap["file_sequence_number"]),
			SequenceNumber:     toNullableInt64(recordMap["sequence_number"]),
			SnapshotID:         recordMap["snapshot_id"].(map[string]any),
			Status:             int(recordMap["status"].(int32)),
			DataFile: formats.DataFile{
				FilePath:         dataFileMap["file_path"].(string),
				FileFormat:       dataFileMap["file_format"].(string),
				RecordCount:      dataFileMap["record_count"].(int64),
				FileSizeInBytes:  dataFileMap["file_size_in_bytes"].(int64),
				Partition:        dataFileMap["partition"].(map[string]any),
				// Add other fields like column stats, etc. if you want
			},
		}

		records.ManifestEntries = append(records.ManifestEntries, manifestEntry)
	}

	return records
}

func CleanSnapshot(recordMaps []map[string]any) (*dto.IcebergSnapshot) {
	
	records := new(dto.IcebergSnapshot)
	
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

		records.SnapshotRecords = append(records.SnapshotRecords, snap)
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
