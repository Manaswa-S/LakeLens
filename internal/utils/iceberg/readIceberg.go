package iceutils

import (
	"encoding/json"
	"fmt"
	"os"

	"main.go/internal/dto/formats"
)

// ReadIcebergJSON reads and Unmarshals given raw iceberg metadata files to IcebergData.
// Files should strictly follow the format given under ./texts/examples .
func ReadIcebergJSON(filePaths []string) ([]*formats.IcebergData, []error) {

	icebergs := make([]*formats.IcebergData, 0)
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

		ice := new(formats.IcebergData)
		err = json.Unmarshal(data, ice)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to unmarshal iceberg json file : %v", err))
			continue
		}

		icebergs = append(icebergs, ice)
	}

	return icebergs, errs
}