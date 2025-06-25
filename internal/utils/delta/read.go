package deltautils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"lakelens/internal/consts/errs"
	formats "lakelens/internal/dto/formats/delta"
	"os"
)

// ReadMetadata reads and Unmarshals given raw delta log file.
// Files should strictly follow the format given under ./texts/examples .
func ReadMetadata(filePath string) (*formats.DeltaLog, *errs.Errorf) {

	log := new(formats.DeltaLog)

	data, err := os.Open(filePath)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrStorageFailed,
			Message: "Failed to read iceberg metadata file : " + err.Error(),
		}
	}

	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		var entry formats.DeltaLogSingle
		err := json.Unmarshal(scanner.Bytes(), &entry)
		if err != nil {
			fmt.Println(err)
			continue
		}

		switch {
			case entry.CommitInfo != nil:
				log.CommitInfo = *entry.CommitInfo
			case entry.Metadata != nil:
				log.Metadata = *entry.Metadata

				log.Metadata.Schema = *unmarshalSchema(entry.Metadata.SchemaString)

			case entry.Protocol != nil:
				log.Protocol = *entry.Protocol
			case entry.Transaction != nil:
				log.Transaction = *entry.Transaction
			case entry.Add != nil:
				log.Add = append(log.Add, *entry.Add)
			case entry.Remove != nil:
				log.Remove = append(log.Remove, *entry.Remove)
			default:
				fmt.Println("none")
		}
	}

	return log, nil
}

func unmarshalSchema(schemaStr string) (*formats.DeltaSchema) {

	var schema formats.DeltaSchema

	err := json.Unmarshal([]byte(schemaStr), &schema)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &schema
}