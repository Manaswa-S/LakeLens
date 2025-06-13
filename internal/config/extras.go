package configs

type ExtraCfg struct {
	DetermineTableTypeMaxDepth int32

	ParquetFilesLimit int32
}

func InitExtraCfg() ExtraCfg {
	return ExtraCfg{
		DetermineTableTypeMaxDepth: 10,
		ParquetFilesLimit: 12,
	}
}