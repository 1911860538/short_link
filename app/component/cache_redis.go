package component

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"

	"github.com/1911860538/short_link/config"
)

type Redis struct {
	client *redis.Client
}

var _ CacheItf = (*Redis)(nil)

var DefaultRedis = &Redis{}

func (r *Redis) Startup() error {
	addr := fmt.Sprintf("%s:%d", config.Conf.Redis.Host, config.Conf.Redis.Port)
	dailTimeout := time.Duration(config.Conf.Redis.ConnTimeout) * time.Second
	opt := &redis.Options{
		Network:     "tcp",
		Addr:        addr,
		Password:    config.Conf.Redis.Password,
		DB:          config.Conf.Redis.Db,
		DialTimeout: dailTimeout,
	}
	r.client = redis.NewClient(opt)

	if err := r.client.Ping().Err(); err != nil {
		return err
	}
	log.Printf("成功连接redis")

	return nil
}

func (r *Redis) Shutdown() error {
	if err := r.client.Close(); err != nil {
		return err
	}
	log.Printf("关闭redis连接")
	return nil
}

func (r *Redis) Set(ctx context.Context, key string, value string, ttl int) error {
	expiration := time.Duration(ttl) * time.Second
	setRes := r.client.WithContext(ctx).Set(key, value, expiration)
	return setRes.Err()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	getRes := r.client.WithContext(ctx).Get(key)
	err := getRes.Err()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}

	return getRes.Val(), nil
}
