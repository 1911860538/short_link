package service

import (
	"context"
	"sync"
	"time"

	"github.com/1911860538/short_link/app/component"
	"github.com/1911860538/short_link/config"
)

type mockCache struct {
	mockLifespan

	mu    sync.Mutex
	cache map[string]*cacheData
}

type cacheData struct {
	value     string
	expiredAt int64
}

func (t *mockCache) Set(ctx context.Context, key string, value string, ttl int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	var expiredAt int64
	if ttl > 0 {
		expiredAt = time.Now().Unix() + int64(ttl)
	} else {
		expiredAt = 0
	}

	t.cache[key] = &cacheData{
		value:     value,
		expiredAt: expiredAt,
	}

	return nil
}

func (t *mockCache) Get(ctx context.Context, key string) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, ok := t.cache[key]
	if !ok {
		return "", nil
	}
	if data.expiredAt != 0 && data.expiredAt < time.Now().Unix() {
		return "", nil
	}

	return data.value, nil
}

var _ component.CacheItf = (*mockCache)(nil)

func getTestCache() *mockCache {
	now := time.Now().Unix()
	cache := &mockCache{
		cache: map[string]*cacheData{
			// 在mock数据库，这个code无过期时间
			"2boMgt": &cacheData{
				value:     "https://www.longurl2.com",
				expiredAt: now + 100,
			},

			// 在mock数据库，这个code有过期时间但未过期
			"fztcmW": &cacheData{
				value:     "https://www.longurl3.com",
				expiredAt: now + 100,
			},

			// 在mock数据库不存在在，缓存中标识code为空
			"empty0": &cacheData{
				value:     config.Conf.Core.CacheNotFoundValue,
				expiredAt: now + 100,
			},
		},
	}

	return cache
}
