package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2apidb/pkg/orm"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/pkg/history"
	"github.com/vennekilde/gw2verify/pkg/sync"
	"github.com/vennekilde/gw2verify/pkg/verify"
	"go.uber.org/zap"
)

func init() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	_ = zap.ReplaceGlobals(logger)
	zap.L().Info("replaced zap's global loggers")
}

func main() {
	flag.Set("stderrthreshold", "INFO")
	flag.Parse()
	defer glog.Flush()

	if config.Config().Debug {
		orm.DB().Debug()
		orm.DB().LogMode(true)
	}

	orm.DB().AutoMigrate(&gw2api.Account{}, &gw2api.TokenInfo{})
	orm.DB().AutoMigrate(&history.History{})
	orm.DB().AutoMigrate(&verify.ServiceLink{}, &verify.TemporaryAccess{})

	/*go func() {
		statistics.Collect()
		for range time.Tick(time.Minute * 5) {
			statistics.Collect()
		}
	}()*/

	go api.StartServer()

	gw2api := gw2api.NewGW2Api()
	go verify.BeginWorldLinksSyncLoop(gw2api)
	sync.StartAPISynchronizer(gw2api)
}
