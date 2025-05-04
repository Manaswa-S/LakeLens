package dto

import formats "lakelens/internal/dto/formats/iceberg"



type IcebergMetadata struct {
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

type IcebergSnapshot struct {
	SnapshotRecords []formats.SnapshotRecord
}

type IcebergManifest struct {
	ManifestEntries []formats.ManifestEntry
}
	


type IsIceberg struct {
	Present bool
	URI string
	MetadataFPaths []string
	ManifestFPaths []string
	SnapshotFPaths []string
	Metadata *IcebergMetadata
	Snapshot *IcebergSnapshot
	Manifest []*IcebergManifest
}