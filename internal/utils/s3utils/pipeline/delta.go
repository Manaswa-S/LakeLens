package pipeline

import (
	"lakelens/internal/dto"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

func HandleDelta(ctx *gin.Context, client *s3.Client, newBucket *dto.NewBucket)
