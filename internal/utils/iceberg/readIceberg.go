package iceutils

import (
	"encoding/json"
	"fmt"
	"os"

	formats "main.go/internal/dto/formats/iceberg"
)

// ReadIcebergJSON reads and Unmarshals given raw iceberg metadata files to IcebergData.
// Files should strictly follow the format given under ./texts/examples .
func ReadIcebergJSON(filePaths []string) ([]*formats.IcebergMetadataData, []error) {

	icebergs := make([]*formats.IcebergMetadataData, 0)
	errs := make([]error, 0)

	for _, path := range filePaths {

		data, err := os.ReadFile(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read iceberg json file : %v", err))
			continue
		}

		if len(data) == 0 {
			errs = append(errs, fmt.Errorf("empty iceberg json file : %v", err))
			continue
		}

		ice := new(formats.IcebergMetadataData)
		err = json.Unmarshal(data, ice)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to unmarshal iceberg json file : %v", err))
			continue
		}

		icebergs = append(icebergs, ice)
	}

	return icebergs, errs
}


func ReadIcebergAvro(filePaths []string) (error) {


	return nil
}