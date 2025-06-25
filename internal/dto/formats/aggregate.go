package formats

import (
	deltaformats "lakelens/internal/dto/formats/delta"
	icebergformats "lakelens/internal/dto/formats/iceberg"
	parquetformats "lakelens/internal/dto/formats/parquet"
)

// These structs are used to aggregate other smaller structs.

type IsIceberg struct {
	Present        bool
	URI            string
	MetadataFPaths []string
	ManifestFPaths []string
	SnapshotFPaths []string
	Metadata       *icebergformats.IcebergMetadata
	Snapshot       []*icebergformats.SnapshotRecord
	Manifest       [][]*icebergformats.ManifestEntry
}

type IsParquet struct {
	Present      bool
	AllFilePaths []string
	Metadata     []*parquetformats.ParquetClean
}

type IsHudi struct {
	Present bool
}

type IsDelta struct {
	Present   bool
	URI       string
	LogFPaths []string
	CRCFPaths []string
	Log       []*deltaformats.DeltaLog
}
