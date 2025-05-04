package formats

type ManifestEntry struct {
	FileSequenceNumber *int64
	SequenceNumber     *int64
	SnapshotID         map[string]any
	Status             int
	DataFile           DataFile
}

type DataFile struct {
	FilePath        string
	FileFormat      string
	RecordCount     int64
	FileSizeInBytes int64
	Partition       map[string]any
}
