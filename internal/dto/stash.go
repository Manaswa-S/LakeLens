package dto

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Used with the stash, more it to the stash folder later.

type S3ClientSave struct {
	LastUsed time.Time
	S3Client *s3.Client
}
