package pipeline

import (
	"lakelens/internal/adapters/s3/engine/fetcher"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	iceutils "lakelens/internal/utils/iceberg"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// HandleIceberg handles downloading, reading and extraction of metadata from given bucket containing Iceberg.
func HandleIceberg(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) (bool, *errs.Errorf) {

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

	var wg sync.WaitGroup
	errfCh := make(chan *errs.Errorf, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(task func() *errs.Errorf) {
			defer wg.Done()
			if errf := task(); errf != nil {
				errfCh <- errf
			}
		}(task) // capture the loop var correctly
	}

	wg.Wait()
	close(errfCh)

	var errsCollected []*errs.Errorf
	for errf := range errfCh {
		errsCollected = append(errsCollected, errf)
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

	filePath, errf := fetcher.FetchNdSave(ctx, client, newBucket.Data.Name, newBucket.Iceberg.MetadataFPaths[metaLen-1])
	if errf != nil {
		return errf
	}

	metadata, errf := iceutils.ReadMetadata(filePath)
	if errf != nil {
		return errf
	}

	newBucket.Iceberg.Metadata = iceutils.CleanMetadata(metadata)

	return nil
}

func snapOps(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) *errs.Errorf {

	// listobjectsv2 returns keys in sorted order tho
	slices.Sort(newBucket.Iceberg.SnapshotFPaths)
	snapLen := len(newBucket.Iceberg.SnapshotFPaths)

	if snapLen <= 0 {
		return &errs.Errorf{
			Type:      errs.ErrInvalidInput,
			Message:   "No 'snap-*.avro' snapshot files were found.",
			ReturnRaw: true,
		}
	}

	filePath, errf := fetcher.FetchNdSave(ctx, client, newBucket.Data.Name, newBucket.Iceberg.SnapshotFPaths[snapLen-1])
	if errf != nil {
		return errf
	}

	newBucket.Iceberg.Snapshot, errf = iceutils.ReadSnapshot(filePath)
	if errf != nil {
		return errf
	}

	return nil
}

func maniOps(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket) *errs.Errorf {

	// listobjectsv2 returns keys in sorted order tho
	slices.Sort(newBucket.Iceberg.ManifestFPaths)
	maniLen := len(newBucket.Iceberg.ManifestFPaths)

	if maniLen <= 0 {
		return &errs.Errorf{
			Type:      errs.ErrInvalidInput,
			Message:   "No '.avro' manifest files were found.",
			ReturnRaw: true,
		}
	}

	for _, path := range newBucket.Iceberg.ManifestFPaths {

		filePath, errf := fetcher.FetchNdSave(ctx, client, newBucket.Data.Name, path)
		if errf != nil {
			return errf
		}

		entries, errf := iceutils.ReadManifest(filePath)
		if errf != nil {
			return errf
		}

		newBucket.Iceberg.Manifest = append(newBucket.Iceberg.Manifest, entries)
	}

	return nil
}
