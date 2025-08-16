package tracelog

import (
	"context"
	"fmt"
	"lakelens/internal/consts/errs"
	sqlc "lakelens/internal/sqlc/generate"
	"strconv"
	"strings"
)

// TODO: enrich is in-efficient, also this is not dynamic on the timeline, so any title set today remains so forever.
// we do not have any way to manipulate certain fields, like names that changed, counts that changed, etc.

// Experimental
// A map to note what each actionID does etc
// ActionID :
var ResolveActionIDMap = map[int64]string{
	1001: "Added {{new_lake_name}}",
	1002: "Visited {{locid}}",
	1003: "Analyzed {{locid}}",
}

var ParamsMap = map[string]func(ctx context.Context, q *sqlc.Queries, params ...any) (string, error){
	"locid": func(ctx context.Context, q *sqlc.Queries, params ...any) (string, error) {

		locID, err := strconv.ParseInt(params[0].(string), 10, 64)
		if err != nil {
			return "", err
		}

		locName, err := q.ResolveLocID(ctx, locID)
		if err != nil {
			return "", err
		}

		return locName, nil
	},
	"new_lake_name": func(ctx context.Context, q *sqlc.Queries, params ...any) (string, error) {
		return params[0].(string), nil
	},
}

var DescriptionsMap = map[int64]func(ctx context.Context, q *sqlc.Queries, userID int64, params map[string]string) (string, error){
	1001: func(ctx context.Context, q *sqlc.Queries, userID int64, params map[string]string) (string, error) {

		lakeID, err := q.UnResolveLakeName(ctx, sqlc.UnResolveLakeNameParams{
			UserID: userID,
			Name:   params["new_lake_name"],
		})
		if err != nil {
			return "", err
		}

		locsCnt, err := q.ResolveDescNewLake(ctx, lakeID)
		if err != nil {
			return "", err
		}

		if locsCnt <= 0 {
			return "Start by adding locations from the lake.", nil
		}

		return "", nil
	},
}

func (s *TraceLog) Enrich(ctx context.Context, recents map[int64]map[int64]*UserAction) *errs.Errorf {

	var err error

	for userID, logs := range recents {

		for _, trace := range logs {

			raw := ResolveActionIDMap[trace.ActionID]

			for param, value := range trace.Store.Data {

				replacement, err := ParamsMap[param](ctx, s.Queries, value)
				if err != nil {
					fmt.Println(err)
					continue
				}

				raw = strings.ReplaceAll(raw, `{{`+param+`}}`, replacement)
			}

			descElem, ok := DescriptionsMap[trace.ActionID]
			if ok {
				trace.Description, err = descElem(ctx, s.Queries, userID, trace.Store.Data)
				if err != nil {
					return nil
				}
			}

			trace.Title = raw
		}
	}

	return nil
}
