package dto


type IsDelta struct {
	Present bool
	URI string
	// MetadataFPaths []string
	// ManifestFPaths []string
	// SnapshotFPaths []string
	// Metadata *IcebergMetadata
	// Snapshot *IcebergSnapshot
	// Manifest []*IcebergManifest
}