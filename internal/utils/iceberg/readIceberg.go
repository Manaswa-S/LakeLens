package iceutils

import (
	"encoding/json"
	"os"

	"main.go/internal/consts/errs"
	formats "main.go/internal/dto/formats/iceberg"
)

// ReadIcebergJSON reads and Unmarshals given raw iceberg metadata file.
// Files should strictly follow the format given under ./texts/examples .
func ReadIcebergJSON(filePath string) (*formats.IcebergMetadataData, *errs.Errorf) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, &errs.Errorf{
			Type: errs.ErrStorageFailed,
			Message: "Failed to read iceberg metadata file : " + err.Error(),
		}
	}

	if len(data) == 0 {
		return nil, &errs.Errorf{
			Type: errs.ErrStorageFailed,
			Message: "Empty iceberg metadata file : filepath = " + filePath,
		}
	}

	iceberg := new(formats.IcebergMetadataData)
	err = json.Unmarshal(data, iceberg)
	if err != nil {
		return nil, &errs.Errorf{
			Type: errs.ErrInternalServer,
			Message: "Failed to un marshal to json : " + err.Error(),
		}
	}

	return iceberg, nil
}


func ReadIcebergAvro(filePaths []string) (error) {


	return nil
}