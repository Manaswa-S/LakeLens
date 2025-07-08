package iceutils

import (
	"bufio"
	"encoding/json"
	"lakelens/internal/consts/errs"
	formats "lakelens/internal/dto/formats/iceberg"
	"os"

	"github.com/linkedin/goavro/v2"
)

// ReadMetadata reads and Unmarshals given raw iceberg metadata file.
// Files should strictly follow the format given under ./texts/examples .
func ReadMetadata(filePath string) (*formats.IcebergMetadata, *errs.Errorf) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrStorageFailed,
			Message: "Failed to read iceberg metadata file : " + err.Error(),
		}
	}

	if len(data) == 0 {
		return nil, &errs.Errorf{
			Type:    errs.ErrStorageFailed,
			Message: "Empty iceberg metadata file : filepath = " + filePath,
		}
	}

	iceberg := new(formats.IcebergMetadata)
	err = json.Unmarshal(data, iceberg)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to un marshal to json : " + err.Error(),
		}
	}

	return iceberg, nil
}

func ReadManifest(filePath string) (*formats.ManifestData, *errs.Errorf) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrStorageFailed,
			Message: "Failed to open iceberg manifest file : " + err.Error(),
		}
	}
	defer file.Close()

	bfile := bufio.NewReader(file)
	ocfr, err := goavro.NewOCFReader(bfile)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to return reader for avro ocf : " + err.Error(),
		}
	}

	entriesMap := make([]map[string]any, 0)

	for ocfr.Scan() {

		datum, err := ocfr.Read()
		if err != nil {
			return nil, &errs.Errorf{
				Type:    errs.ErrDependencyFailed,
				Message: "Failed to read from avro ocf : " + err.Error(),
			}
		}

		recordMap, ok := datum.(map[string]any)
		if !ok {
			return nil, &errs.Errorf{
				Type:    errs.ErrDependencyFailed,
				Message: "Avro ocf datum isn't of expected 'map[string]any' type.",
			}
		}
		entriesMap = append(entriesMap, recordMap)
	}

	entries := CleanManifestEntry(entriesMap)

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	ocfrMeta := ocfr.MetaData()

	metadata := CleanManifestMetadata(ocfrMeta)

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	return &formats.ManifestData{
		Metadata: metadata,
		Entries:  entries,
	}, nil
}

func ReadSnapshot(filePath string) (*formats.IcebergSnapshot, *errs.Errorf) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrStorageFailed,
			Message: "Failed to open iceberg snapshot file : " + err.Error(),
		}
	}
	defer file.Close()

	bfile := bufio.NewReader(file)
	ocfr, err := goavro.NewOCFReader(bfile)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDependencyFailed,
			Message: "Failed to return reader for avro ocf : " + err.Error(),
		}
	}

	recordMaps := make([]map[string]any, 0)

	for ocfr.Scan() {

		datum, err := ocfr.Read()
		if err != nil {
			return nil, &errs.Errorf{
				Type:    errs.ErrDependencyFailed,
				Message: "Failed to read from avro ocf : " + err.Error(),
			}
		}

		recordMap, ok := datum.(map[string]any)
		if !ok {
			return nil, &errs.Errorf{
				Type:    errs.ErrDependencyFailed,
				Message: "Avro ocf datum isn't of expected 'map[string]any' type.",
			}
		}
		recordMaps = append(recordMaps, recordMap)
	}

	records := CleanSnapshot(recordMaps)

	return &formats.IcebergSnapshot{
		Records: records,
	}, nil
}
