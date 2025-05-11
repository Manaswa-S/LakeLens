package deltautils

import (
	"lakelens/internal/dto"
	formats "lakelens/internal/dto/formats/delta"
)

func CleanMetadata(log *formats.DeltaLog) *dto.DeltaMetaData {

	delta := new(dto.DeltaMetaData)

	delta.CommitInfo = log.CommitInfo
	
	delta.Metadata = formats.DeltaMetadata{
		ID: log.Metadata.ID,
		Format: log.Metadata.Format,
		Schema: log.Metadata.Schema,
		PartitionColumns: log.Metadata.PartitionColumns,
		Configuration: log.Metadata.Configuration,
		CreatedTime: log.Metadata.CreatedTime,
	}

	delta.Protocol = log.Protocol
	delta.Transaction = log.Txn

	return delta
}
