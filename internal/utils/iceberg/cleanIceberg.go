package iceutils

import (
	"main.go/internal/dto"
	formats "main.go/internal/dto/formats/iceberg"
)


func CleanIceberg(iceberg *formats.IcebergMetadataData) (*dto.IcebergClean) {

	cleanIceberg := new(dto.IcebergClean)

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