package iceberg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"lakelens/internal/consts"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	formats "lakelens/internal/dto/formats/iceberg"
	sqlc "lakelens/internal/sqlc/generate"
	"lakelens/internal/stash"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type IcebergService struct {
	Queries     *sqlc.Queries
	RedisClient *redis.Client
	DB          *pgxpool.Pool

	Stash *stash.StashService
}

func NewIcebergService(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool, stash *stash.StashService) *IcebergService {
	return &IcebergService{
		Queries:     queries,
		RedisClient: redis,
		DB:          db,

		Stash: stash,
	}
}

func (s *IcebergService) fetchCache(ctx *gin.Context, userID int64, locid string) (*stash.CacheMetadata, *errs.Errorf) {

	locID, err := strconv.ParseInt(locid, 10, 64)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrInvalidInput,
			Message: "Failed to parse location id to int64 : " + err.Error(),
		}
	}

	locData, err := s.Queries.GetLocationData(ctx, locID)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get location data : " + err.Error(),
		}
	}

	if locData.UserID != userID {
		return nil, &errs.Errorf{
			Type:      errs.ErrUnauthorized,
			Message:   "Requested resource does not belong to you.",
			ReturnRaw: true,
		}
	}

	lakeData, err := s.Queries.GetLakeData(ctx, locData.LakeID)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to get lake data : " + err.Error(),
		}
	}

	cache := new(stash.CacheMetadata)
	var exists bool

	switch lakeData.Ptype {
	case consts.AWSS3:
		cache, exists = s.Stash.GetBucketS3(locData.BucketName)

	default:
		// ?
	}

	if !exists {
		return nil, &errs.Errorf{
			Type:      errs.ErrNotFound,
			Message:   "Requested resource not found. Please rescan to fetch data.",
			ReturnRaw: true,
		}
	}

	if cache.Bucket.Data.TableType != consts.IcebergTable {
		return nil, &errs.Errorf{
			Type:      errs.ErrResourceLocked,
			Message:   "Requested resource is not of expected table type (iceberg).",
			ReturnRaw: true,
		}
	}

	return cache, nil
}

func (s *IcebergService) GetOverviewData(ctx *gin.Context, userID int64, locid string) (*dto.OverviewData, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	fileURIs := make(map[string]string, 0)
	fileURIs["Metadata File"] = cache.Bucket.Iceberg.MetadataFPaths[len(cache.Bucket.Iceberg.MetadataFPaths)-1]

	currSnapID := cache.Bucket.Iceberg.Metadata.CurrentSnapshotID
	for _, snap := range cache.Bucket.Iceberg.Metadata.Snapshots {
		if snap.SnapshotID == currSnapID {
			fileURIs["Snapshot File"] = snap.ManifestList
			break
		}
	}

	return &dto.OverviewData{
		FoundAt:     cache.Bucket.Iceberg.URI,
		Location:    cache.Bucket.Iceberg.Metadata.Location,
		TableUUID:   cache.Bucket.Iceberg.Metadata.TableUUID,
		FilesReadMp: map[string]int64{},
		TableType:   cache.Bucket.Data.TableType,
		FileURIs:    fileURIs,
	}, nil
}

func (s *IcebergService) GetOverviewStats(ctx *gin.Context, userID int64, locid string) (*dto.OverviewStats, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	var latestSnap formats.IcebergMetadataSnapshot

	currSnapID := cache.Bucket.Iceberg.Metadata.CurrentSnapshotID
	for _, snap := range cache.Bucket.Iceberg.Metadata.Snapshots {
		if snap.SnapshotID == currSnapID {
			latestSnap = snap
			break
		}
	}

	snapSummary := latestSnap.Summary

	addrecs, _ := strconv.ParseInt(snapSummary.AddedRecords, 10, 64)
	addposdels, _ := strconv.ParseInt(snapSummary.AddedPositionDeletes, 10, 64)
	addeqdels, _ := strconv.ParseInt(snapSummary.AddedEqualityDeletes, 10, 64)

	delta := addrecs - addposdels - addeqdels

	totSize, _ := strconv.ParseInt(snapSummary.TotalFilesSize, 10, 64)
	totDataFiles, _ := strconv.ParseInt(snapSummary.TotalDataFiles, 10, 64)
	avgFileSize := totSize / totDataFiles

	return &dto.OverviewStats{
		Table: dto.OverviewStatsTable{
			TableType:    cache.Bucket.Data.TableType,
			TableVersion: cache.Bucket.Iceberg.Metadata.FormatVersion,
		},
		Rows: dto.OverviewStatsRowCount{
			TotalCount: snapSummary.TotalRecords,
			DeltaCount: delta,
		},
		Version: dto.OverviewStatsVersion{
			CurrentVersion: strconv.FormatInt(cache.Bucket.Iceberg.Metadata.CurrentSnapshotID, 10),
			LastSnapshot:   latestSnap.TimestampMS,
			TotalSnapshots: latestSnap.SequenceNumber,
		},
		Storage: dto.OverviewStatsStorage{
			TotalSize:      totSize,
			TotalDataFiles: totDataFiles,
			AvgFileSize:    avgFileSize,
		},
	}, nil
}

func (s *IcebergService) GetOverviewSchema(ctx *gin.Context, userID int64, locid string) (*dto.OverviewSchema, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	var latestSchema formats.IcebergSchema

	currSchID := cache.Bucket.Iceberg.Metadata.CurrentSchemaID
	for _, schema := range cache.Bucket.Iceberg.Metadata.Schemas {
		if schema.SchemaID == currSchID {
			latestSchema = schema
			break
		}
	}

	fields := make([]*dto.OverviewSchemaField, 0)
	for _, field := range latestSchema.Fields {
		fields = append(fields, &dto.OverviewSchemaField{
			ID:       field.ID,
			Name:     field.Name,
			Type:     field.Type,
			Required: field.Required,
		})
	}

	return &dto.OverviewSchema{
		SchemaID: latestSchema.SchemaID,
		Fields:   fields,
	}, nil
}

func (s *IcebergService) GetOverviewPartition(ctx *gin.Context, userID int64, locid string) (*dto.OverviewPartition, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	var latestSpec formats.IcebergPartitionSpec

	currSpecID := cache.Bucket.Iceberg.Metadata.DefaultSpecID
	for _, spec := range cache.Bucket.Iceberg.Metadata.PartitionSpecs {
		if spec.SpecID == currSpecID {
			latestSpec = spec
			break
		}
	}

	fields := make([]*dto.OverviewPartitionField, 0)
	for _, field := range latestSpec.Fields {
		fields = append(fields, &dto.OverviewPartitionField{
			Name:      field.Name,
			Transform: field.Transform,
			FieldID:   field.FieldID,
			SourceID:  field.SourceID,
		})
	}

	fields = append(fields, &dto.OverviewPartitionField{
		Name:      "demo-partition",
		Transform: "bucket[11]",
		FieldID:   4311,
		SourceID:  5555,
	})

	return &dto.OverviewPartition{
		DefaultSpecID: latestSpec.SpecID,
		Fields:        fields,
	}, nil
}

func (s *IcebergService) GetOverviewSnapshot(ctx *gin.Context, userID int64, locid string) (*dto.OverviewSnapshot, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	var latestSnap formats.IcebergMetadataSnapshot

	currSnapID := cache.Bucket.Iceberg.Metadata.CurrentSnapshotID
	for _, snap := range cache.Bucket.Iceberg.Metadata.Snapshots {
		if snap.SnapshotID == currSnapID {
			latestSnap = snap
			break
		}
	}

	return &dto.OverviewSnapshot{
		SequenceNumber: latestSnap.SequenceNumber,
		LastOperation:  latestSnap.Summary.Operation,
		SchemaID:       latestSnap.SchemaID,

		ManifestList: latestSnap.ManifestList,
	}, nil

}

func (s *IcebergService) GetOverviewGraphs(ctx *gin.Context, userID int64, locid string) ([]*dto.OverviewGraphs, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	snapshots := cache.Bucket.Iceberg.Metadata.Snapshots

	resp := make([]*dto.OverviewGraphs, 0)
	for _, snapshot := range snapshots {
		resp = append(resp, &dto.OverviewGraphs{
			TimeStampMS:      snapshot.TimestampMS,
			TotalRecords:     snapshot.Summary.TotalRecords,
			TotalFileSize:    snapshot.Summary.TotalFilesSize,
			TotalDataFiles:   snapshot.Summary.TotalDataFiles,
			TotalDeleteFiles: snapshot.Summary.TotalDeleteFiles,
		})
	}

	return resp, nil
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

func (s *IcebergService) GetSchemasList(ctx *gin.Context, userID int64, locid string) (*dto.SchemaList, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	abc := make(map[int64]*dto.SchemaListData, 0)

	snaps := cache.Bucket.Iceberg.Metadata.Snapshots

	for _, snap := range snaps {

		if abc[snap.SchemaID] == nil {
			abc[snap.SchemaID] = new(dto.SchemaListData)
			abc[snap.SchemaID].SchemaID = snap.SchemaID
			abc[snap.SchemaID].FromTimeStampMS = snap.TimestampMS
		}
		abc[snap.SchemaID].ValidUptoSnapshotID = strconv.FormatInt(snap.SnapshotID, 10)
	}

	return &dto.SchemaList{
		List: abc,
	}, nil
}

func (s *IcebergService) GetSchema(ctx *gin.Context, userID int64, locid string, schemaid string) (*dto.Schema, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	schemaID, err := strconv.ParseInt(schemaid, 10, 64)
	if err != nil {
		return nil, &errs.Errorf{
			Type:    errs.ErrInvalidInput,
			Message: "Failed to parse schema id to int64 : " + err.Error(),
		}
	}

	schemas := cache.Bucket.Iceberg.Metadata.Schemas

	var schema formats.IcebergSchema
	for _, sch := range schemas {
		if sch.SchemaID == schemaID {
			schema = sch
			break
		}
	}

	fields := make([]*dto.SchemaField, 0)
	for _, field := range schema.Fields {
		fields = append(fields, &dto.SchemaField{
			ID:       field.ID,
			Name:     field.Name,
			Type:     field.Type,
			Required: field.Required,
		})
	}

	// if schemaID == 0 {
	// 	for _, f := range fields {
	// 		f.Required = !f.Required
	// 	}
	// 	fields[0].Type = "dadsda"
	// }

	// if schemaID == 1 {
	// 	fields = slices.Delete(fields, len(fields)-2, len(fields)-1)

	// }

	// manifests := cache.Bucket.Iceberg.Manifest

	// currManifest := manifests[0].Data

	// for _, mani := range currManifest {

	// 	fmt.Println(mani.URI)
	// 	fmt.Println(mani.Metadata.Schema)
	// 	// fmt.Println(mani.Metadata.Content)
	// 	for _, entry := range mani.Entries {

	// 		fmt.Println(entry.DataFile.ColumnSizes)

	// 		// for _, raw := range entry.DataFile.LowerBounds {
	// 		// 	for _, r := range raw.([]any) {
	// 		// 		rmap, ok := r.(map[string]any)
	// 		// 		if !ok {
	// 		// 			fmt.Println("not a map")
	// 		// 			continue
	// 		// 		}

	// 		// 		if rmap["key"].(int32) == 2 {

	// 		// 			fmt.Println("🗝️ Key:", rmap["key"])
	// 		// 			switch val := rmap["value"].(type) {
	// 		// 			case []byte:
	// 		// 				printByteGuess(val)
	// 		// 			case string:
	// 		// 				printByteGuess([]byte(val))
	// 		// 			default:
	// 		// 				fmt.Println("⚠️ Unknown type:", reflect.TypeOf(val))
	// 		// 			}
	// 		// 		}
	// 		// 	}

	// 		// }
	// 	}
	// }

	return &dto.Schema{
		SchemaID: schemaID,
		Fields:   fields,
	}, nil
}

func printByteGuess(b []byte) {
	if len(b) == 4 {
		var v int32
		_ = binary.Read(bytes.NewReader(b), binary.LittleEndian, &v)
		fmt.Println("🧩 Guessed int32:", v)
	}

	if len(b) == 8 {
		var i64 int64
		var f64 float64

		err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &i64)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("🧩 Guessed int64:", i64)
		}

		_ = binary.Read(bytes.NewReader(b), binary.LittleEndian, &f64)
		fmt.Println("🧪 Guessed float64:", f64)
	}

	fmt.Println("📜 As string (maybe):", string(b))
}

func (s *IcebergService) GetSchemaData(ctx *gin.Context, userID int64, locid string, schemaid string) (*dto.SchemaData, *errs.Errorf) {

	// cache, errf := s.fetchCache(ctx, userID, locid)
	// if errf != nil {
	// 	return nil, errf
	// }

	// schemaID, err := strconv.ParseInt(schemaid, 10, 64)
	// if err != nil {
	// 	return nil, &errs.Errorf{
	// 		Type:    errs.ErrInvalidInput,
	// 		Message: "Failed to parse schema id to int64 : " + err.Error(),
	// 	}
	// }

	// schemas := cache.Bucket.Iceberg.Metadata.Schemas

	// var schema formats.IcebergSchema
	// for _, sch := range schemas {
	// 	if sch.SchemaID == schemaID {
	// 		schema = sch
	// 		break
	// 	}
	// }

	return &dto.SchemaData{}, nil

}

func (s *IcebergService) GetSchemaColSizes(ctx *gin.Context, userID int64, locid string, schemaid string) (*dto.SchemaColSizes, *errs.Errorf) {

	cache, errf := s.fetchCache(ctx, userID, locid)
	if errf != nil {
		return nil, errf
	}

	var currSchemaID int64
	var err error

	if schemaid == "latest" {
		currSchemaID = cache.Bucket.Iceberg.Metadata.CurrentSchemaID
	} else {
		currSchemaID, err = strconv.ParseInt(schemaid, 10, 64)
		if err != nil {
			return nil, &errs.Errorf{
				Type:    errs.ErrInvalidInput,
				Message: "Failed to parse schema id to int64 : " + err.Error(),
			}
		}
	}

	schemas := cache.Bucket.Iceberg.Metadata.Schemas

	var schema formats.IcebergSchema
	for _, sch := range schemas {
		if sch.SchemaID == currSchemaID {
			schema = sch
			break
		}
	}

	colSizeMap := make(map[int64]int64)
	nullsCountMap := make(map[int64]int64)
	valsCountMap := make(map[int64]int64)

	mani := cache.Bucket.Iceberg.Manifest[0]
	for _, data := range mani.Data {
		if data.Metadata.Content != "deletes" {
			for _, entry := range data.Entries {
				errf = s.expandDataFile(entry.DataFile.ColumnSizes, colSizeMap)
				if errf != nil {
					return nil, errf
				}

				errf = s.expandDataFile(entry.DataFile.ValueCounts, valsCountMap)
				if errf != nil {
					return nil, errf
				}

				errf = s.expandDataFile(entry.DataFile.NullValueCounts, nullsCountMap)
				if errf != nil {
					return nil, errf
				}
			}
		}
	}

	result := make([]dto.ColSize, 0)

	for key, val := range colSizeMap {
		for _, f := range schema.Fields {
			if f.ID == key {
				result = append(result, dto.ColSize{
					ID:            f.ID,
					Name:          f.Name,
					Size:          val,
					NullCount:     nullsCountMap[key],
					ValueCount:    valsCountMap[key],
					AvgSizePerVal: (float32(val) / float32(valsCountMap[key]-nullsCountMap[key])),
				})
			}
		}
	}

	slices.SortFunc(result, func(a dto.ColSize, b dto.ColSize) int { return int(a.ID) - int(b.ID) })

	return &dto.SchemaColSizes{
		SchemaID: schema.SchemaID,
		ColSizes: result,
	}, nil
}

func (s *IcebergService) expandDataFile(dataMap map[string]any, resultMap map[int64]int64) *errs.Errorf {

	dataArr := dataMap["array"].([]any)
	for _, elem := range dataArr {

		elem, ok := elem.(map[string]any)
		if !ok {
			return &errs.Errorf{
				Type:    errs.ErrInternalServer,
				Message: "The iceberg column_size array element is not of map[string]any type.",
			}
		}

		value, err := strconv.ParseInt(fmt.Sprintf("%v", elem["value"]), 10, 64)
		if err != nil {
			return &errs.Errorf{
				Type:    errs.ErrInternalServer,
				Message: "The iceberg column_size array element's 'value' field is not of int64 type : " + err.Error(),
			}
		}

		key, err := strconv.ParseInt(fmt.Sprintf("%v", elem["key"]), 10, 64)
		if err != nil {
			return &errs.Errorf{
				Type:    errs.ErrInternalServer,
				Message: "The iceberg column_size array element's 'key' field is not of int64 type : " + err.Error(),
			}
		}

		resultMap[key] += value
	}

	return nil
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// func (s *IcebergService) AllData(ctx *gin.Context, userID int64, locid string) (*dto.IsIceberg, *errs.Errorf) {

// 	cache, errf := s.fetchCache(ctx, userID, locid)
// 	if errf != nil {
// 		return nil, errf
// 	}

// 	return &cache.Bucket.Iceberg, nil
// }

// func (s *IcebergService) Metadata(ctx *gin.Context, userID int64, locid string) (*dto.IcebergMetadata, *errs.Errorf) {

// 	cache, errf := s.fetchCache(ctx, userID, locid)
// 	if errf != nil {
// 		return nil, errf
// 	}

// 	return cache.Bucket.Iceberg.Metadata, nil
// }

// func (s *IcebergService) Snapshot(ctx *gin.Context, userID int64, locid string) (*dto.IcebergSnapshot, *errs.Errorf) {

// 	cache, errf := s.fetchCache(ctx, userID, locid)
// 	if errf != nil {
// 		return nil, errf
// 	}

// 	return cache.Bucket.Iceberg.Snapshot, nil
// }

// func (s *IcebergService) Manifest(ctx *gin.Context, userID int64, locid string) ([]*dto.IcebergManifest, *errs.Errorf) {

// 	cache, errf := s.fetchCache(ctx, userID, locid)
// 	if errf != nil {
// 		return nil, errf
// 	}

// 	return cache.Bucket.Iceberg.Manifest, nil
// }
