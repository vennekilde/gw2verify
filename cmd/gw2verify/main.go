package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2apidb/pkg/orm"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

func main() {
	flag.Set("stderrthreshold", "INFO")
	flag.Parse()
	defer glog.Flush()

	if config.Config().Debug {
		orm.DB().Debug()
		orm.DB().LogMode(true)
	}

	orm.DB().AutoMigrate(&gw2api.Account{}, &gw2api.TokenInfo{})
	//orm.DB().AutoMigrate(verify.ServiceLink{}, verify.TemporaryAccess{})
	orm.DB().AutoMigrate(&verify.ServiceLink{}, &verify.TemporaryAccess{})

	go api.StartServer()

	gw2api := gw2api.NewGW2Api()
	verify.StartAPISynchronizer(gw2api)
}
