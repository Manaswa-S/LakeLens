package s3utils

import (
	"path"
	"strings"

	"main.go/internal/consts"
	"main.go/internal/dto"
)

func GetAllFilePaths(tree dto.AllFilesMp, currentPath string) []string {
	var filePaths []string

	for name, subTree := range tree {
		fullPath := path.Join(currentPath, name)

		if nested, ok := subTree.(dto.AllFilesMp); ok {
			filePaths = append(filePaths, GetAllFilePaths(nested, fullPath)...)
		} else {
			filePaths = append(filePaths, fullPath)
		}
	}

	return filePaths
}


func GetIcebergFilePaths(tree dto.AllFilesMp, currentPath string) ([]string, []string) {
	var jsonPaths []string
	var avroPaths []string

	for name, subTree := range tree {
		fullPath := path.Join(currentPath, name)

		if nested, ok := subTree.(dto.AllFilesMp); ok {
			jps, aps := GetIcebergFilePaths(nested, fullPath)
			jsonPaths = append(jsonPaths, jps...)
			avroPaths = append(avroPaths, aps...)
		} else {
			ext := path.Ext(fullPath)
			switch ext {
			case ".json":
				jsonPaths = append(jsonPaths, fullPath)
			case ".avro":
				avroPaths = append(avroPaths, fullPath)
			}
		}
	}

	return jsonPaths, avroPaths
}

func DetermineType(tree dto.AllFilesMp) (string) {

	// identify Iceberg files
	if _, exists := tree["data"]; exists {
		if _, exists := tree["metadata"]; exists {
			return consts.IcebergFile
		}
	}

	// jsonData, _ := json.MarshalIndent(tree, "", "  ")
	// fmt.Println(string(jsonData))

	for objName, subt := range tree {

		if subt, ok := subt.(dto.AllFilesMp); ok {
			result := DetermineType(subt)
			if result == consts.IcebergFile {
				return consts.IcebergFile
			} else if result == consts.ParquetFile {
				return consts.ParquetFile
			} 
		} else {
			if path.Ext(objName) == consts.ParquetFile {
				return consts.ParquetFile
			} else {
				return consts.UnknownFile
			}
		}
	}

	return consts.UnknownFile
}

func InsertIntoTree(tree dto.AllFilesMp, path string) {
	parts := strings.Split(path, "/")
	current := tree

	l := len(parts) - 1

	for i, part := range parts {

		if part == "" {
			continue
		}

		if i == l {
			current[part] = nil
		} else {
			if _, exists := current[part]; !exists {
				current[part] = dto.AllFilesMp{}
			}
			current = current[part].(dto.AllFilesMp)
		}
	}
}