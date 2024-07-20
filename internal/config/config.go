package config

import (
	"time"

	"github.com/tkanos/gonfig"
	"go.uber.org/zap"
)

type Configuration struct {
	DeleteDataAfterDay            *time.Duration
	ExpirationTime                int
	TemporaryAccessExpirationTime int
	SkipRestrictions              bool
	Debug                         bool
	CollectStatisticsAfter        time.Time
	SyncInterval                  time.Duration

	// DB
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDatabase string
}

var loaded = false
var config = Configuration{
	SyncInterval: time.Second,
}

func Config() Configuration {
	if !loaded {
		err := gonfig.GetConf("", &config)
		if err != nil {
			zap.L().Fatal("could not load config", zap.Error(err))
		}
	}
	return config
}
