package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/MrGunflame/gw2api"
	"github.com/alexlast/bunzap"
	"github.com/uptrace/bun"
	"github.com/vennekilde/gw2verify/v2/internal/config"
	"github.com/vennekilde/gw2verify/v2/internal/orm"
	"github.com/vennekilde/gw2verify/v2/internal/server"
	"github.com/vennekilde/gw2verify/v2/pkg/history"
	"github.com/vennekilde/gw2verify/v2/pkg/sync"
	"github.com/vennekilde/gw2verify/v2/pkg/verify"
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

	// Services initialization
	worldsService := verify.NewWorlds(gw2api.New())
	verificationService := verify.NewVerification(worldsService)
	statisticsService := history.NewStatistics(verificationService)
	eventEmitter := verify.NewEventEmitter(verificationService)
	syncService := sync.NewService(gw2api.New(), eventEmitter)
	banService := verify.NewBanService(eventEmitter)

	// REST endpoints
	verificationEndpoints := server.NewVerificationEndpoint(verificationService, worldsService, statisticsService, eventEmitter, syncService, banService)
	endpoints := server.NewEndpoints(verificationEndpoints)
	// REST server
	restServer := server.NewRESTServer(endpoints)
	go restServer.Start()

	go worldsService.Start()
	syncService.Start()
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
