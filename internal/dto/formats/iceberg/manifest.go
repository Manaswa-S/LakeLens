package formats

// Main structs used to unmarshal the m0.avro manifest files of iceberg tables.
// These are different from the manifest data found in .metadata.json files.

type ManifestEntry struct {
	FileSequenceNumber *int64
	SequenceNumber     *int64
	SnapshotID         map[string]any
	Status             int
	DataFile           ManifestDataFile
}

type ManifestDataFile struct {
	FilePath        string
	FileFormat      string
	RecordCount     int64
	FileSizeInBytes int64
	Partition       map[string]any
}
