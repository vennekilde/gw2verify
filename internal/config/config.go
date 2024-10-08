package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Configuration struct {
	DeleteDataAfter               *time.Duration `mapstructure:"DELETE_DATA_AFTER"`
	ExpirationTime                int            `mapstructure:"EXPIRATION_TIME"`
	TemporaryAccessExpirationTime int            `mapstructure:"TEMPORARY_ACCESS_EXPIRATION_TIME"`
	SkipRestrictions              bool           `mapstructure:"SKIP_RESTRICTIONS"`
	Debug                         bool           `mapstructure:"DEBUG"`
	CollectStatisticsAfter        time.Time      `mapstructure:"COLLECT_STATISTICS_AFTER"`
	SyncInterval                  time.Duration  `mapstructure:"SYNC_INTERVAL"`
	MaxConcurrentSyncs            int32          `mapstructure:"MAX_CONCURRENT_SYNCS"`

	// DB
	PostgresHost     string `mapstructure:"POSTGRES_HOST"`
	PostgresPort     int    `mapstructure:"POSTGRES_PORT"`
	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDatabase string `mapstructure:"POSTGRES_DATABASE"`
}

var loaded = false
var config *Configuration

func Config() *Configuration {
	if !loaded {
		v := viper.NewWithOptions(viper.ExperimentalBindStruct())
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		conf := Configuration{
			MaxConcurrentSyncs: 5,
			SyncInterval:       time.Second,
		}

		err := v.Unmarshal(&conf)
		if err != nil {
			zap.L().Panic("could not load config", zap.Error(err))
		} else {
			config = &conf
			loaded = true
		}
	}
	return config
}
