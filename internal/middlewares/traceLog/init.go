package tracelog

import (
	sqlc "lakelens/internal/sqlc/generate"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

type TraceLogConfigs struct {
	BufChanCapacity int   // the main chan capacity
	FlushInterval   int64 // the timed trigger interval, in seconds
	BufChanFullLim  int   // the main chan 'full' lim, recommended to be kept <=70% of BufChanCapacity
}

type TraceLogInternals struct {
	BufChan atomic.Pointer[chan *UserAction] // the main chan where UserAction is pushed
	AltChan atomic.Pointer[chan *UserAction] // the stand by chan ready to be swapped in

	TickChan <-chan time.Time // the timed interval flush tick chan

	FlushTrigger chan TriggerType // the flush trigger, flush when anything pushed to it
	LastFlush    int64            // the unix seconds of last flush, to avoid redundant flushes from both timed and capacity full
}

type TraceLog struct {
	Queries     *sqlc.Queries
	RedisClient *redis.Client

	Internals TraceLogInternals
	Configs   TraceLogConfigs
}

func NewTraceLog(queries *sqlc.Queries, redis *redis.Client) *TraceLog {
	ret := &TraceLog{
		Queries:     queries,
		RedisClient: redis,
	}

	ret.Configs.BufChanCapacity = 100
	ret.Configs.BufChanFullLim = 70
	ret.Configs.FlushInterval = 60

	ret.InitSave()

	return ret
}
