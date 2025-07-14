package pipeline

import (
	"fmt"
	"lakelens/internal/adapters/s3/engine/fetcher"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	formats "lakelens/internal/dto/formats/iceberg"
	iceutils "lakelens/internal/utils/iceberg"
	"path"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// HandleIceberg handles downloading, reading and extraction of metadata from given bucket containing Iceberg.
func HandleIceberg(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) (bool, *errs.Errorf) {

	// TODO: paginate this
	resp, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &newBucket.Data.Name,
		Prefix: &newBucket.Iceberg.URI,
	})
	if err != nil {
		return false, &errs.Errorf{
			Type:    errs.ErrServiceUnavailable,
			Message: "Failed to list objects : " + err.Error(),
		}
	}

	for _, obj := range resp.Contents {

		key := *obj.Key
		if strings.HasSuffix(key, ".metadata.json") {
			newBucket.Iceberg.MetadataFPaths = append(newBucket.Iceberg.MetadataFPaths, key)
		} else if path.Ext(key) == ".avro" {
			if _, fname := path.Split(key); strings.HasPrefix(fname, "snap-") {
				newBucket.Iceberg.SnapshotFPaths = append(newBucket.Iceberg.SnapshotFPaths, key)
			} else {
				newBucket.Iceberg.ManifestFPaths = append(newBucket.Iceberg.ManifestFPaths, key)
			}
		}
	}

	newBucket.Errors = runOps([]func() *errs.Errorf{
		func() *errs.Errorf { return metaOps(ctx, client, newBucket) },
		func() *errs.Errorf { return snapOps(ctx, client, newBucket) },
		func() *errs.Errorf { return maniOps(ctx, client, newBucket) },
	})

	return false, nil
}

func runOps(tasks []func() *errs.Errorf) []*errs.Errorf {

	var errsCollected []*errs.Errorf

	for _, task := range tasks {
		if errf := task(); errf != nil {
			errsCollected = append(errsCollected, errf)
		}
	}

	return errsCollected
}

func metaOps(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) *errs.Errorf {

	// listobjectsv2 returns keys in sorted order tho
	slices.Sort(newBucket.Iceberg.MetadataFPaths)
	metaLen := len(newBucket.Iceberg.MetadataFPaths)

	if metaLen <= 0 {
		return &errs.Errorf{
			Type:      errs.ErrInvalidInput,
			Message:   "No '.metadata.json' metadata files were found.",
			ReturnRaw: true,
		}
	}

	filePath, errf := fetcher.FetchNdSave(ctx, client, newBucket.Data.Name, newBucket.Iceberg.MetadataFPaths[metaLen-1], "")
	if errf != nil {
		return errf
	}

	metadata, errf := iceutils.ReadMetadata(filePath)
	if errf != nil {
		return errf
	}

	newBucket.Iceberg.Metadata = metadata

	return nil
}

func snapOps(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) *errs.Errorf {

	var snapPath string
	snaps := newBucket.Iceberg.Metadata.Snapshots

	if len(snaps) <= 0 {
		return &errs.Errorf{
			Type:      errs.ErrInvalidInput,
			Message:   "No 'snap-*.avro' snapshot files were found.",
			ReturnRaw: true,
		}
	}

	currSnapID := newBucket.Iceberg.Metadata.CurrentSnapshotID
	for _, snap := range snaps {
		if snap.SnapshotID == currSnapID {
			snapPath = snap.ManifestList
			break
		}
	}

	filePath, errf := fetcher.FetchNdSave(ctx, client, newBucket.Data.Name, "", snapPath)
	if errf != nil {
		fmt.Println(*errf)
		return errf
	}

	snap, errf := iceutils.ReadSnapshot(filePath)
	if errf != nil {
		return errf
	}

	newBucket.Iceberg.Snapshot = append(newBucket.Iceberg.Snapshot, snap)

	return nil
}

func maniOps(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) *errs.Errorf {

	snaps := newBucket.Iceberg.Snapshot
	if len(snaps) <= 0 {
		return &errs.Errorf{
			Type:      errs.ErrInvalidInput,
			Message:   "No '.avro' manifest files were found.",
			ReturnRaw: true,
		}
	}

	snapRecords := snaps[0].Records
	if len(snapRecords) <= 0 {
		return &errs.Errorf{
			Type:      errs.ErrInvalidInput,
			Message:   "No '.avro' manifest files were found.",
			ReturnRaw: true,
		}
	}

	data := make([]*formats.ManifestData, 0)

	for _, record := range snapRecords {

		// TODO: ok so theres a problem here,
		// suppose theres a bucket with a table in it, the table files were copied from another bucket that has the og table.
		// so now the files all have their names according to the og bucket but the new bucket has a different name,
		// and hence paths/locations don't match then.
		// So, maybe we can get the og name from the 'location' field or something.
		filePath, errf := fetcher.FetchNdSave(ctx, client, newBucket.Data.Name, "", record.ManifestPath)
		if errf != nil {
			return errf
		}

		entries, errf := iceutils.ReadManifest(filePath)
		if errf != nil {
			return errf
		}
		entries.URI = record.ManifestPath

		data = append(data, entries)
	}

	newBucket.Iceberg.Manifest = append(newBucket.Iceberg.Manifest, &formats.IcebergManifest{Data: data})

	return nil
}
