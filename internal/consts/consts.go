package consts

// General file/table type names.
// Change with care as this is directly linked to external routes.
const (
	ParquetFile  = "parquet"
	IcebergTable = "iceberg"
	DeltaTable   = "delta"
	HudiTable    = "hudi"

	UnknownFile = "unknown"
)

// General storage type names. Do Not Change.
const (
	AWSS3 = "awsS3"
	Azure = "azure"
	MinIO = "minIO"
)

const (
	EPassAuth   = "epass"
	GoogleOAuth = "goauth"
)

var (
	AuthTypeExchange = map[string]string{
		"epass":  "Email Password",
		"goauth": "Google OAuth",
	}
)

// File/Table type extension/folder names to detect them. Do Not Change.
const (
	ParquetFileExt = ".parquet"

	IcebergMetaFolder = "/metadata/"
	IcebergDataFolder = "/data/"

	DeltaLogFolder = "/_delta_log/"

	HudiMetaFolder = "/.hoodie/"
)
