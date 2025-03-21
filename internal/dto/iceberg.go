package dto

import formats "main.go/internal/dto/formats/iceberg"


type IcebergClean struct {
	TableUUID string
	Location string
	LastUpdatedMS int64

	CurrentSchemaID int64
	Schemas []formats.IcebergSchema

	CurrentSnapshotID int64
	Snapshots []formats.IcebergSnapshot
	SnapshotLog []formats.IcebergSnapshotLog

	MetadataLog []formats.IcebergMetadataLog

	DefaultSpecID int64
	PartitionSpecs []formats.IcebergPartitionSpec
	LastPartitonID int64
}

// type IcebergCleanSnapshot struct {
// 	SnapshotID int64
// 	TimestampMS int64
// 	Operation string
	
// 	AddedFiles int64
// 	AddedRecords int64
// 	AddedFilesSize int64
// 	TotalRecords int64
// 	TotalFilesSize int64
// }

type IsIceberg struct {
	Present bool
	JSONFilePaths []string
	AvroFilePaths []string
}