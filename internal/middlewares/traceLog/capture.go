package tracelog

import (
	"time"

	"github.com/gin-gonic/gin"
)

func (t *TraceLog) Capture() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result := new(UserAction)
		result.Store.Data = make(map[string]string)

		path := ctx.FullPath()
		timeNow := time.Now().Unix()

		info, ok := RoutesList[path]

		ctx.Next()

		if ok {
			result.ActionID = info.ActionID
			result.Store.Action = info.Action
			result.Time = timeNow
			result.UserID = ctx.GetInt64("rid")
			for _, key := range info.Params {
				result.Store.Data[key] = ctx.Param(key)
			}
			for _, key := range info.BodyFields {
				result.Store.Data[key] = ctx.GetString(key)
			}

			bufChanPtr := t.Internals.BufChan.Load()
			if len(*bufChanPtr) >= t.Configs.BufChanFullLim {
				t.TriggerIt(CapacityTrigger)
			}
			*bufChanPtr <- result
		}
	}
}
