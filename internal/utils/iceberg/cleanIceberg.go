package iceutils

import (
	"main.go/internal/dto"
	formats "main.go/internal/dto/formats/iceberg"
)


func CleanIceberg(icebergs []*formats.IcebergMetadataData) ([]*dto.IcebergClean, error) {

	cleanIcebergs := make([]*dto.IcebergClean, 0)

	for _, ice := range icebergs {

		cleanIce := dto.IcebergClean{
			TableUUID: ice.TableUUID,
			Location: ice.Location,
			LastUpdatedMS: ice.LastUpdatedMS,

			CurrentSchemaID: ice.CurrentSchemaID,
			Schemas: ice.Schemas,

			CurrentSnapshotID: ice.CurrentSnapshotID,
			Snapshots: ice.Snapshots,
			SnapshotLog: ice.SnapshotLog,

			MetadataLog: ice.MetadataLog,

			DefaultSpecID: ice.DefaultSpecID,
			PartitionSpecs: ice.PartitionSpecs,
			LastPartitonID: ice.LastPartitionID,
		}

		cleanIcebergs = append(cleanIcebergs, &cleanIce)
	}

	return cleanIcebergs, nil
}