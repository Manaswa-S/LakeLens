package iceberg

import "lakelens/internal/services/iceberg"


type IcebergHandler struct {
	Iceberg *iceberg.IcebergService
}

func NewIcebergHandler(iceberg *iceberg.IcebergService) *IcebergHandler {
	return &IcebergHandler{
		Iceberg: iceberg,
	}
}


