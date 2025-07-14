package tracelog

import (
	"context"
	"encoding/json"
	"fmt"
	sqlc "lakelens/internal/sqlc/generate"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func (t *TraceLog) InitSave() {

	bufChan := make(chan *UserAction, t.Configs.BufChanCapacity)
	t.Internals.BufChan.Store(&bufChan)
	altChan := make(chan *UserAction, t.Configs.BufChanCapacity)
	t.Internals.AltChan.Store(&altChan)

	t.Internals.FlushTrigger = make(chan TriggerType, 1)
	t.Internals.TickChan = time.Tick(time.Duration(t.Configs.FlushInterval) * time.Second)
	t.Internals.LastFlush = time.Now().Unix()

	go t.initTimedWait()
	go t.waitForTrigger()
}

func (t *TraceLog) initTimedWait() {
	for range t.Internals.TickChan {
		fmt.Println("its time for a mandatory flush")
		t.TriggerIt(TimedTrigger)
	}
}

func (t *TraceLog) TriggerIt(triggerType TriggerType) {
	select {
	case t.Internals.FlushTrigger <- triggerType:
	default:
		fmt.Println("FlushTrigger chan is already full, another flush is in progress, skipping")
	}
}

func (t *TraceLog) waitForTrigger() {
	for trigger := range t.Internals.FlushTrigger {

		if trigger == TimedTrigger {
			if (time.Now().Unix() - t.Internals.LastFlush) < t.Configs.FlushInterval {
				fmt.Println("last flush wasn't interval seconds ago, redundant flush, skipping")
				continue
			}
		}

		if trigger == CapacityTrigger {
			bufChanPtr := t.Internals.BufChan.Load()
			if len(*bufChanPtr) < t.Configs.BufChanFullLim {
				fmt.Println("capacity is not upto limit, redundant flush, skipping")
				continue
			}
		}

		t.FlushChans()
	}
}

func (t *TraceLog) FlushChans() {
	fmt.Println("Processing recents logs batch")

	t.Internals.LastFlush = time.Now().Unix()

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	altChanPtr := t.Internals.AltChan.Load()

	if len(*altChanPtr) > 0 { // this can backfire but lets keep it for now
		fmt.Println("the alternate channel still has elems, risk of loss of logs, skipping for now")
		return
	}

	bufChanPtr := t.Internals.BufChan.Swap(altChanPtr)
	_ = t.Internals.AltChan.Swap(bufChanPtr)

	if len(*bufChanPtr) <= 0 {
		fmt.Println("the channel is still empty, redundant flush, skipping")
		return
	}

	t.drainChan(ctx, bufChanPtr)
}

func (t *TraceLog) drainChan(ctx context.Context, toDrainChan *chan *UserAction) {

	// a map of userID to map of actionID to actionData
	uniqueMap := make(map[int64]map[int64]*UserAction, 0)
	// can replace with a direct range, but this just guarantees that the chan is drained completely.
	for len(*toDrainChan) > 0 {
		temp := <-*toDrainChan

		if uniqueMap[temp.UserID] == nil {
			uniqueMap[temp.UserID] = make(map[int64]*UserAction)
		}

		// here we are assuming that the order or events is preserved in the channel and that they are in
		// order of latest last. FIFO.
		uniqueMap[temp.UserID][temp.ActionID] = temp
	}

	err := t.Enrich(ctx, uniqueMap)
	if err != nil {
		fmt.Println(err)
		return
	}

	forRedis := make(map[int64][]any)
	forDB := new(SaveToDBData)

	for userID, actionMap := range uniqueMap {
		redisList := make([]any, 0, len(actionMap))

		for actionID, action := range actionMap {

			actionByte, err := json.Marshal(action)
			if err != nil {
				fmt.Println(err)
				return
			}

			storeByte, err := json.Marshal(action.Store)
			if err != nil {
				fmt.Println(err)
				return
			}

			redisList = append(redisList, actionByte)

			forDB.UserIDs = append(forDB.UserIDs, userID)
			forDB.Times = append(forDB.Times, pgtype.Timestamptz{Time: time.Unix(action.Time, 0), Valid: true})
			forDB.ActionIDs = append(forDB.ActionIDs, actionID)
			forDB.Actions = append(forDB.Actions, storeByte)
			forDB.Titles = append(forDB.Titles, action.Title)
			forDB.Descriptions = append(forDB.Descriptions, action.Description)
		}

		forRedis[userID] = redisList
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func(ctx context.Context, actionMap map[int64][]any) {
		defer wg.Done()
		t.pushToRedis(ctx, forRedis)
	}(ctx, forRedis)

	go func(ctx context.Context, data *SaveToDBData) {
		defer wg.Done()
		t.pushToDB(ctx, forDB)
	}(ctx, forDB)

	wg.Wait()
}

func (t *TraceLog) pushToRedis(ctx context.Context, actionMap map[int64][]any) {

	for userID, list := range actionMap {

		err := t.RedisClient.LPush(ctx, fmt.Sprintf("userid:%d:recents", userID), list...).Err()
		if err != nil {
			fmt.Println(err)
			return
		}

		err = t.RedisClient.LTrim(ctx, fmt.Sprintf("userid:%d:recents", userID), 0, 49).Err()
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func (t *TraceLog) pushToDB(ctx context.Context, data *SaveToDBData) {

	if (len(data.UserIDs) != len(data.Times)) ||
		(len(data.ActionIDs) != len(data.Actions)) ||
		(len(data.Times) != len(data.ActionIDs)) ||
		(len(data.Titles) != len(data.Descriptions)) ||
		(len(data.Titles) != len(data.UserIDs)) {
		return
	}

	err := t.Queries.InsertRecentsBulk(ctx, sqlc.InsertRecentsBulkParams{
		Column1: data.UserIDs,
		Column2: data.ActionIDs,
		Column3: data.Times,
		Column4: data.Actions,
		Column5: data.Titles,
		Column6: data.Descriptions,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

}
