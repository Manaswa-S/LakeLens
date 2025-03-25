package s3utils

// import (
// 	"path"
// 	"strings"

// 	"github.com/aws/aws-sdk-go-v2/service/s3/types"
// 	"main.go/internal/consts"
// 	"main.go/internal/dto"
// )

// func GetAllFilePaths(tree dto.FileTreeMap, currentPath string) []string {
// 	var filePaths []string

// 	for name, subTree := range tree {
// 		fullPath := path.Join(currentPath, name)

// 		if nested, ok := subTree.(dto.FileTreeMap); ok {
// 			filePaths = append(filePaths, GetAllFilePaths(nested, fullPath)...)
// 		} else {
// 			filePaths = append(filePaths, fullPath)
// 		}
// 	}

// 	return filePaths
// }


// func GetIcebergFilePaths(tree dto.FileTreeMap, currentPath string) ([]string, []string) {
// 	var jsonPaths []string
// 	var avroPaths []string

// 	for name, subTree := range tree {
// 		fullPath := path.Join(currentPath, name)

// 		if nested, ok := subTree.(dto.FileTreeMap); ok {
// 			jps, aps := GetIcebergFilePaths(nested, fullPath)
// 			jsonPaths = append(jsonPaths, jps...)
// 			avroPaths = append(avroPaths, aps...)
// 		} else {
// 			ext := path.Ext(fullPath)
// 			switch ext {
// 			case ".json":
// 				jsonPaths = append(jsonPaths, fullPath)
// 			case ".avro":
// 				avroPaths = append(avroPaths, fullPath)
// 			}
// 		}
// 	}

// 	return jsonPaths, avroPaths
// }

// func DetermineType(tree dto.FileTreeMap) (string) {

// 	// identify Iceberg files
// 	if _, exists := tree["data"]; exists {
// 		if _, exists := tree["metadata"]; exists {
// 			return consts.IcebergFile
// 		}
// 	}

// 	// jsonData, _ := json.MarshalIndent(tree, "", "  ")
// 	// fmt.Println(string(jsonData))

// 	for objName, subt := range tree {

// 		if subt, ok := subt.(dto.FileTreeMap); ok {
// 			result := DetermineType(subt)
// 			if result == consts.IcebergFile {
// 				return consts.IcebergFile
// 			} else if result == consts.ParquetFile {
// 				return consts.ParquetFile
// 			} 
// 		} else {
// 			if path.Ext(objName) == consts.ParquetFile {
// 				return consts.ParquetFile
// 			} else {
// 				return consts.UnknownFile
// 			}
// 		}
// 	}

// 	return consts.UnknownFile
// }

// func InsertIntoTree(tree dto.FileTreeMap, contents *[]types.Object) {
// 	var parts []string
// 	var lenParts int

// 	for _, obj := range *contents {

// 		parts = strings.Split(*obj.Key, "/")
// 		current := tree
// 		lenParts = len(parts) - 1

// 		for i, part := range parts {
// 			if part == "" {continue}

// 			if i == lenParts {
// 				current[part] = nil
// 			} else {
// 				if _, exists := current[part]; !exists {
// 					current[part] = dto.FileTreeMap{}
// 				}
// 				current = current[part].(dto.FileTreeMap)
// 			}
// 		}
// 	}
// }