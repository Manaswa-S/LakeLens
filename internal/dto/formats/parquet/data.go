package formats

import "github.com/xitongsys/parquet-go/parquet"

type ParquetClean struct {
	URI            string
	Schema         []*parquet.SchemaElement
	CreatedBy      *string
	Version        int32
	NumRows        int64
	EncryptionAlgo *parquet.EncryptionAlgorithm
}
