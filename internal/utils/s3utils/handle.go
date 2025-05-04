package s3utils

import (
	"fmt"
	configs "lakelens/internal/config"
	"lakelens/internal/consts"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	iceutils "lakelens/internal/utils/iceberg"
	parqutils "lakelens/internal/utils/parquet"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// HandleIceberg handles downloading, reading and extraction of metadata from given bucket containing Iceberg.
func HandleIceberg(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) (*dto.IsIceberg, bool, *errs.Errorf) {

	resp, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: newBucket.Data.Name,
		Prefix: &newBucket.Iceberg.URI,
	})
	if err != nil {
		return nil, false, &errs.Errorf{
			Type: errs.ErrServiceUnavailable,
			Message: "Failed to list objects : " + err.Error(),
		}
	}
	

	// trial:
	// beyond this point we use traditional caching, lets try a cheeky shortcut here

	if newBucket.Data.KeyCount == 0 {
		newBucket.Data.KeyCount = int64(*resp.KeyCount)
	} else {
		if newBucket.Data.KeyCount == int64(*resp.KeyCount) {
			return nil, true, nil
		}
	}


	fmt.Println(newBucket.Data.KeyCount)



	//



	latestUpdate := time.Time{}

	for _, obj := range resp.Contents {
		
		// trial:

		if obj.LastModified.After(latestUpdate) {
			latestUpdate = *obj.LastModified
		}

		//

		key := *obj.Key
		if strings.HasSuffix(key, ".metadata.json") {
			newBucket.Iceberg.MetadataFPaths = append(newBucket.Iceberg.MetadataFPaths, key)
		} else {
			if path.Ext(key) == ".avro" {
				if _, fname := path.Split(key); strings.HasPrefix(fname, "snap-") {
					newBucket.Iceberg.SnapshotFPaths = append(newBucket.Iceberg.SnapshotFPaths, key)
				} else {
					newBucket.Iceberg.ManifestFPaths = append(newBucket.Iceberg.ManifestFPaths, key)
				}
			}
		}
	}

	// trial:

	if !latestUpdate.After(newBucket.Data.UpdatedAt) && !latestUpdate.IsZero() {
		return nil, true, nil
	}
	newBucket.Data.UpdatedAt = latestUpdate

	//


	// Metadata file Ops:

		metaPaths := newBucket.Iceberg.MetadataFPaths
		// listobjectsv2 returns keys in sorted order tho
		// sort.Strings(jsonPaths)
		metaLen := len(metaPaths)

		if metaLen <= 0 {
			return nil, false, &errs.Errorf{
				Type: errs.ErrInvalidInput,
				Message: "No '.metadata.json' metadata files were found.",
				ReturnRaw: true,
			}
		}

		filePath, errf := DownIcebergS3(ctx, client, *newBucket.Data.Name, metaPaths[metaLen - 1])
		if errf != nil {
			return nil, false, errf
		}

		metadata, errf := iceutils.ReadMetadata(filePath)
		if errf != nil {
			return nil, false, errf
		}

		cleanMetadata := iceutils.CleanMetadata(metadata)

	//



	// Snapshot file Ops:
		snapPaths := newBucket.Iceberg.SnapshotFPaths
		// listobjectsv2 returns keys in sorted order tho
		// sort.Strings(jsonPaths)
		snapLen := len(snapPaths)

		if snapLen <= 0 {
			return nil, false, &errs.Errorf{
				Type: errs.ErrInvalidInput,
				Message: "No 'snap-*.avro' snapshot files were found.",
				ReturnRaw: true,
			}
		}

		filePath, errf = DownAvroS3(ctx, client, *newBucket.Data.Name, snapPaths[snapLen - 1])
		if errf != nil {
			return nil, false, errf
		}

		cleanSnapshots, errf := iceutils.ReadSnapshot(filePath)
		if errf != nil {
			return nil, false, errf
		}

		// cleanIceberg := iceutils.CleanIceberg(metadata)

	//


	// Manifest file Ops:
		maniPaths := newBucket.Iceberg.ManifestFPaths
		// listobjectsv2 returns keys in sorted order tho
		// sort.Strings(jsonPaths)
		maniLen := len(maniPaths)

		if maniLen <= 0 {
			return nil, false, &errs.Errorf{
				Type: errs.ErrInvalidInput,
				Message: "No '.avro' manifest files were found.",
				ReturnRaw: true,
			}
		}

		var allentries []*dto.IcebergManifest

		for _, pths := range maniPaths {
			filePath, errf = DownAvroS3(ctx, client, *newBucket.Data.Name, pths)
			if errf != nil {
				return nil, false, errf
			}

			entries, errf := iceutils.ReadManifest(filePath)
			if errf != nil {
				return nil, false, errf
			}
			allentries = append(allentries, entries)

		}
		// cleanIceberg := iceutils.CleanIceberg(metadata)

	//

	return &dto.IsIceberg{
		Metadata: cleanMetadata,
		Snapshot: cleanSnapshots,
		Manifest: allentries,
	}, false, nil
}

func HandleParquet(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) ([]*dto.ParquetClean, bool, *errs.Errorf) {

	resp, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: newBucket.Data.Name,
	})
	if err != nil {
        return nil, false, &errs.Errorf{
			Type: errs.ErrServiceUnavailable,
			Message: "Failed to list parquet objects : " + err.Error(),
		}
    }

	limit := configs.ParquetFilesLimit
	latestUpdate := time.Time{}

	for _, obj := range resp.Contents {
		if limit <= 0 {
			break
		}

		//
		// fmt.Printf("%s : %s : %s\n", *obj.Key, obj.LastModified, newBucket.Data.UpdatedAt)
		if obj.LastModified.After(latestUpdate) {
			latestUpdate = *obj.LastModified
		}
		//

		if obj.Key != nil {
			key := *obj.Key
			if key[len(key) - 1] != '/' && strings.HasSuffix(key, consts.ParquetFileExt){
				newBucket.Parquet.AllFilePaths = append(newBucket.Parquet.AllFilePaths, key)
				limit--
			}
		}
	}

	if !latestUpdate.After(newBucket.Data.UpdatedAt) && !latestUpdate.IsZero(){
		return nil, true, nil
	}
	newBucket.Data.UpdatedAt = latestUpdate

	var wg sync.WaitGroup
	cleanParquets := make([]*dto.ParquetClean, 0)

	for _, path := range newBucket.Parquet.AllFilePaths {
		wg.Add(1)

		go func(path string) {
			defer wg.Done()

			filePath, errf := DownloadSingleParquetS3(ctx, client, *newBucket.Data.Name, path)
			if errf != nil {
				// TODO: handle error, retry logic
				return
			}

			cleanParquet, errf := parqutils.ReadParquet(filePath)
			if errf != nil {
				fmt.Println(errf.Message)
				return
			}

			cleanParquet.URI = path

			cleanParquets = append(cleanParquets, cleanParquet)
		} (path)
	}

	wg.Wait()
	
	return cleanParquets, false, nil
}