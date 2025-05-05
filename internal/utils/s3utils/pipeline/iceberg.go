package pipeline

import (
	"fmt"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	iceutils "lakelens/internal/utils/iceberg"
	fetcher "lakelens/internal/utils/s3utils/engine/fetcher"
	"path"
	"strings"
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
			Type:    errs.ErrServiceUnavailable,
			Message: "Failed to list objects : " + err.Error(),
		}
	}

	// trial:
	// beyond this point we use traditional caching, lets try a cheeky shortcut here
	// this will need to be adjusted for maxCount > keyCount
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
			Type:      errs.ErrInvalidInput,
			Message:   "No '.metadata.json' metadata files were found.",
			ReturnRaw: true,
		}
	}

	filePath, errf := fetcher.DownIcebergS3(ctx, client, *newBucket.Data.Name, metaPaths[metaLen-1])
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
			Type:      errs.ErrInvalidInput,
			Message:   "No 'snap-*.avro' snapshot files were found.",
			ReturnRaw: true,
		}
	}

	filePath, errf = fetcher.DownAvroS3(ctx, client, *newBucket.Data.Name, snapPaths[snapLen-1])
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
			Type:      errs.ErrInvalidInput,
			Message:   "No '.avro' manifest files were found.",
			ReturnRaw: true,
		}
	}

	var allentries []*dto.IcebergManifest

	for _, pths := range maniPaths {
		filePath, errf = fetcher.DownAvroS3(ctx, client, *newBucket.Data.Name, pths)
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
