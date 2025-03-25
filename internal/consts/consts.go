package consts

// General file/table type names.
const (
	ParquetFile = "parquet"
	IcebergTable = "iceberg"
	DeltaTable = "delta"
	HudiTable = "hudi"

	UnknownFile = "unknown"
)

// General storage type names.
const (
	AWSS3 = "awsS3"
)

// File/Table type extension/folder names to detect them. Do Not Change.
const (
	ParquetFileExt = ".parquet"

	IcebergMetaFolder = "/metadata/"
	IcebergDataFolder = "/data/"

	DeltaLogFolder = "/_delta_log/"

	HudiMetaFolder = "/.hoodie/"
)