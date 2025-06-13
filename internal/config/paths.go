package configs

type DownSavePaths struct {
	BaseSavePath string

	S3DownPath string
}

type PathsCfg struct {
	// < HTML Templates
	ResetPassHTMLPath string
	// >

	// < File Paths
	ErrorLoggerFilePath   string
	RequestLoggerFilePath string
	// >

	// < Down Save Paths
	DownSavePaths DownSavePaths
	// >
}

func InitPathsCfg() PathsCfg {
	return PathsCfg{
		ResetPassHTMLPath: "/home/mnswa/zdev/go/projects/LakeLens/templates/emails/reset-pass.html",

		ErrorLoggerFilePath:   "./logs/errors.log",
		RequestLoggerFilePath: "./logs/requests.log",

		DownSavePaths: DownSavePaths{
			BaseSavePath: "./lakeDownloads",
			S3DownPath:   "./lakeDownloads" + "/s3",
		},
	}
}
