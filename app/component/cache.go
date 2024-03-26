package component

import (
	"context"
	"log"

	"github.com/1911860538/short_link/config"
)

type CacheItf interface {
	Lifespan

	Set(ctx context.Context, key string, value string, ttl int) error
	Get(ctx context.Context, key string) (string, error)
}

var Cache CacheItf

func init() {
	switch cacheType := config.Conf.Server.CacheType; cacheType {
	case "redis":
		Cache = DefaultRedis
	default:
		log.Fatalf("不支持的缓存组件：%s\n", cacheType)
	}
}
