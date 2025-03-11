package cutils

// func ReadParquetFileMetadata() error {

// 	for _, fKey := range fileKeys {
// 		fr, err := local.NewLocalFileReader(fKey)
// 		if err != nil {
// 			return err
// 		}
// 		defer fr.Close()

// 		pr, err := reader.NewParquetReader(fr, nil, 4)
// 		if err != nil {
// 			return err
// 		}

// 		filePR = append(filePR, pr)
// 	}

// 	return nil
// }