package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/alexlast/bunzap"
	"github.com/uptrace/bun"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/internal/orm"
	"github.com/vennekilde/gw2verify/internal/server"
	"github.com/vennekilde/gw2verify/pkg/sync"
	"github.com/vennekilde/gw2verify/pkg/verify"
	"gitlab.com/MrGunflame/gw2api"
	"go.uber.org/zap"
)

func init() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	_ = zap.ReplaceGlobals(logger)
	zap.L().Info("replaced zap's global loggers")
}

func main() {
	bunzapHook := bunzap.QueryHookOptions{
		Logger:       zap.L(),
		SlowDuration: 200 * time.Millisecond, // Omit to log all operations as debug
	}
	if config.Config().Debug {
		// Print all sql queries
		bunzapHook.SlowDuration = 0
	}
	hook := QueryHookMiddleware{
		Next: bunzap.NewQueryHook(bunzapHook),
	}
	orm.DB().AddQueryHook(hook)

	/*go func() {
		statistics.Collect()
		for range time.Tick(time.Minute * 5) {
			statistics.Collect()
		}
	}()*/

	restServer := server.NewRESTServer()
	go restServer.Start()

	go verify.BeginWorldLinksSyncLoop(gw2api.New())
	sync.StartAPISynchronizer(gw2api.New())
}

type QueryHookMiddleware struct {
	Next bun.QueryHook
}

func (qh QueryHookMiddleware) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if event.Err == sql.ErrNoRows {
		// Suppress no rows errors
		event.Err = nil
		qh.Next.AfterQuery(ctx, event)
		// Re-set sql.ErrNoRows error
		event.Err = sql.ErrNoRows
	} else {
		qh.Next.AfterQuery(ctx, event)
	}
}

func (qh QueryHookMiddleware) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return qh.Next.BeforeQuery(ctx, event)
}
