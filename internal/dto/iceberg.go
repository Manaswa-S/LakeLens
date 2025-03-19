package dto

import "main.go/internal/dto/formats"


type IcebergClean struct {
	Schemas []formats.IcebergSchema
}

type IsIceberg struct {
	Present bool
	JSONFilePaths []string
	AvroFilePaths []string
}