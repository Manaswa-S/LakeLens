package middlewares

import (
	"encoding/json"
	"fmt"
	"lakelens/cmd/errpipe"
	"lakelens/internal/dto"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(requestsLoggerFile *os.File) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		startTime := time.Now()

		ctx.Next()

		logData := dto.RequestLogData{
			StartTime:     startTime.Format("2006-01-02 15:04:05"),
			ClientIP:      ctx.ClientIP(),
			Method:        ctx.Request.Method,
			Path:          ctx.Request.URL.Path,
			StatusCode:    ctx.Writer.Status(),
			InternalError: ctx.Errors.String(),
			Latency:       time.Since(startTime).Milliseconds(),
		}

		errorCheck(ctx, &logData)

		jsonLog, err := json.Marshal(logData)
		if err != nil {
			errpipe.NewError(&errpipe.ErrorData{
				Error:   "Failed to marshal error struct to json : " + err.Error(),
				LogData: &logData,
			})
			return
		}
		// TODO: use a buffered writer instead, too much i/o ops involved here
		_, err = requestsLoggerFile.WriteString(string(jsonLog) + "\n")
		if err != nil {
			errpipe.NewError(&errpipe.ErrorData{
				Error:   "Failed to write request log in logger file : " + err.Error(),
				LogData: &logData,
			})
			return
		}
	}
}

func errorCheck(ctx *gin.Context, logData *dto.RequestLogData) {

	errData := new(errpipe.ErrorData)
	var hasErr bool

	if debugErr, exists := ctx.Get("debug"); exists {
		errData.Debug = fmt.Sprintf("%s", debugErr)
		hasErr = true
	} else if infoErr, exists := ctx.Get("info"); exists {
		errData.Info = fmt.Sprintf("%s", infoErr)
		hasErr = true
	} else if warnErr, exists := ctx.Get("warn"); exists {
		errData.Warn = fmt.Sprintf("%s", warnErr)
		hasErr = true
	} else if errorErr, exists := ctx.Get("error"); exists {
		errData.Error = fmt.Sprintf("%s", errorErr)
		hasErr = true
	} else if criticalErr, exists := ctx.Get("critical"); exists {
		errData.Critical = fmt.Sprintf("%s", criticalErr)
		hasErr = true
	} else if fatalErr, exists := ctx.Get("fatal"); exists {
		errData.Fatal = fmt.Sprintf("%s", fatalErr)
		hasErr = true
	}
	if hasErr {
		errData.LogData = logData
		errpipe.NewError(errData)
	}
}
