package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"github.com/spf13/viper"
	"github.com/windrivder/gopkg/errorx"
	"github.com/windrivder/gopkg/logx"
)

// Options
type Options struct {
	Addr        string        `json:"Addr"`
	Password    string        `json:"Password"`
	DB          int           `json:"DB"`
	PoolSize    int           `json:"PoolSize"`
	IdleTimeout time.Duration `json:"IdleTimeout"`
}

// NewOptions
func NewOptions(v *viper.Viper) (o Options, err error) {
	if err = v.UnmarshalKey("Redis", &o); err != nil {
		return o, errorx.Wrap(err, "unmarshal redis server option error")
	}

	return o, nil
}

// New
func New(o Options) (client *redis.Client, cleanFunc func(), err error) {
	client = redis.NewClient(&redis.Options{
		Addr:        o.Addr,
		Password:    o.Password,
		DB:          o.DB,
		PoolSize:    o.PoolSize,
		IdleTimeout: o.IdleTimeout * time.Second,
	})
	cleanFunc = func() {
		err := client.Close()
		if err != nil {
			logx.Fatal().Msgf("redis close error: %v", err)
		}
	}

	ping, err := client.Ping(context.Background()).Result()
	if err != nil {
		logx.Fatal().Msg("redis close errors")
	}

	logx.Info().
		Str("ping", ping).
		Str("addr", o.Addr).
		Int("db", o.DB).
		Msg("redis connected")

	return client, cleanFunc, nil
}

var ProviderSet = wire.NewSet(New, NewOptions)
