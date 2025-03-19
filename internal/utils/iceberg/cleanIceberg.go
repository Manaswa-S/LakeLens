package iceutils

import (
	"main.go/internal/dto"
	"main.go/internal/dto/formats"
)


func CleanIceberg(icebergs []*formats.IcebergData) ([]*dto.IcebergClean, error) {

	cleanIcebergs := make([]*dto.IcebergClean, 0)

	for _, ice := range icebergs {
		cleanIce := dto.IcebergClean{
			Schemas: ice.Schemas,
		}
		cleanIcebergs = append(cleanIcebergs, &cleanIce)
	}

	return cleanIcebergs, nil
}