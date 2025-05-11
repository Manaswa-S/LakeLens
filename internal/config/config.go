package configs

var (
	Internal = InitInternalConfig()
)

const (
	BaseSavePath = "./lakeDownloads"
)

const (
	// TODO: to be changed to a better/secure/separate location from the actual code center
	S3DownPath = BaseSavePath + "/s3"
)

const (
	DetermineTableTypeMaxDepth = 10

	ParquetFilesLimit = 12
)