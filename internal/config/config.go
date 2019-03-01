package config

import (
	"github.com/golang/glog"
	"github.com/tkanos/gonfig"
)

type Configration struct {
	RESTAuthToken                 string
	APIKeyPrefix                  string
	ExpirationTime                int
	TemporaryAccessExpirationTime int
	HomeWorld                     int
	Debug                         bool
}

var loaded = false
var config = Configration{}

func Config() Configration {
	if loaded == false {
		err := gonfig.GetConf("", &config)
		if err != nil {
			glog.Fatalf("Could not load config. Error: %s", err.Error())
		}
	}
	return config
}
