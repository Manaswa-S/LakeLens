package dto

import formats "lakelens/internal/dto/formats/delta"


type DeltaMetaData struct {
	CommitInfo formats.DeltaCommitInfo
	Protocol formats.DeltaProtocol
	Metadata formats.DeltaMetadata
	Transaction formats.DeltaTxn
}

type IsDelta struct {
	Present bool
	URI string
	LogFPaths []string
	CRCFPaths []string
	Metadata []*DeltaMetaData
}