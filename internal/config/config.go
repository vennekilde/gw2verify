package config

import (
	"github.com/golang/glog"
	"github.com/tkanos/gonfig"
	"go.uber.org/zap"
)

type Configration struct {
	RESTAuthToken                 string
	ExpirationTime                int
	TemporaryAccessExpirationTime int
	SkipRestrictions              bool
	Debug                         bool
}

var loaded = false
var config = Configration{}

func Config() Configration {
	if !loaded {
		err := gonfig.GetConf("", &config)
		if err != nil {
			zap.L().Fatal("could not load config", zap.Error(err))
		}
	}
	return config
}
