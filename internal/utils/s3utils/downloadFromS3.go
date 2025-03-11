package s3utils

// func dwnFiles(client *s3.Client) error {

// 	rangeHeader := "bytes=-8192"

// 	// client.

// 	for _, fKey := range fileKeys {
// 		obj, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
// 			Bucket: &bucket,
// 			Key: &fKey,
// 			Range: &rangeHeader,
// 		})
// 		if err != nil {
// 			return err
// 		}
// 		defer obj.Body.Close()

// 		outfile, err := os.Create(fKey)
// 		if err != nil {
// 			return err
// 		}

// 		_, err = outfile.ReadFrom(obj.Body)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	fmt.Println("DOWNLOAD SUCCESSFUL")

// 	return nil
// }
 