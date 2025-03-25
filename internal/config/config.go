package configs

var (
	Internal = InitInternalConfig()
)

const (
	// TODO: to be changed to a better/secure/separate location from the actual code center
	IcebergDownloadS3Path = "./lakeDownloads/s3"

	ParquetDownloadS3Path = "./lakeDownloads/s3"
)

const (
	DetermineTableTypeMaxDepth = 10

	ParquetFilesLimit = 12
)