package service

import (
	"context"
	"maps"
	"net/http"
	"strings"
	"testing"

	"github.com/1911860538/short_link/app/server/service"
	"github.com/1911860538/short_link/config"
)

// 测试跳转，code字符非法，直接返回
func TestRedirectInvalidCodeLen(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	cache := getTestCache()
	redirectSvc := &service.RedirectSvc{
		Cache:    cache,
		Database: db,
	}

	// 后续逻辑判断，缓存是否有变化
	cacheSnapshot := make(map[string]*cacheData)
	for k, v := range cache.cache {
		cacheSnapshot[k] = v
	}

	invalidCodes := []string{
		strings.Repeat("a", 10000), // 长度非法
		"你好",                       // 长度合法，但是不是数字字母组合
		"a?X1u2",                   // 长度合法，但是包含非数字字母
		"-?*#~;",                   // 长度合法，但是包含非数字字母
	}

	for _, invalidCode := range invalidCodes {
		redirectRes, err := redirectSvc.Do(context.Background(), invalidCode)
		if err != nil {
			t.Error(err)
			return
		}

		if redirectRes.Redirect {
			t.Errorf("预期是否跳转为falese，实际返回%t", redirectRes.Redirect)
			return
		}

		if redirectRes.StatusCode != http.StatusNotFound {
			t.Errorf("期望返回http状态码404，实际返回%d\n", redirectRes.StatusCode)
			return
		}

		// 判断code长度非法后，直接返回，没有对缓存进行操作
		if !maps.Equal(cache.cache, cacheSnapshot) {
			t.Error("期望不对缓存操作，实际操作了")
			return
		}
	}
}

// 测试跳转，code在缓存中存在
func TestRedirectCodeInCache(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	cache := getTestCache()
	redirectSvc := &service.RedirectSvc{
		Cache:    cache,
		Database: db,
	}

	inCacheCode := "2boMgt"
	redirectRes, err := redirectSvc.Do(context.Background(), inCacheCode)
	if err != nil {
		t.Error(err)
		return
	}

	if redirectRes.StatusCode != config.Conf.Core.RedirectStatusCode {
		t.Errorf("期望返回http状态码%d，实际返回%d\n", config.Conf.Core.RedirectStatusCode, redirectRes.StatusCode)
		return
	}

	if !redirectRes.Redirect {
		t.Errorf("预期是否跳转为true，实际返回%t", redirectRes.Redirect)
		return
	}

	if redirectRes.LongUrl == "" {
		t.Error("预期获得长链接url不为空，实际返回空")
		return
	}
}

// 测试跳转，code在缓存中不存在，但在数据库中有
func TestRedirectCodeInDb(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	cache := getTestCache()
	redirectSvc := &service.RedirectSvc{
		Cache:    cache,
		Database: db,
	}

	// 后续逻辑判断，缓存是否有变化
	cacheSnapshot := make(map[string]*cacheData)
	for k, v := range cache.cache {
		cacheSnapshot[k] = v
	}

	inDbCode := "rCeB3h"
	redirectRes, err := redirectSvc.Do(context.Background(), inDbCode)
	if err != nil {
		t.Error(err)
		return
	}

	if redirectRes.StatusCode != config.Conf.Core.RedirectStatusCode {
		t.Errorf("期望返回http状态码%d，实际返回%d\n", config.Conf.Core.RedirectStatusCode, redirectRes.StatusCode)
		return
	}

	if !redirectRes.Redirect {
		t.Errorf("预期是否跳转为true，实际返回%t", redirectRes.Redirect)
		return
	}

	if redirectRes.LongUrl == "" {
		t.Error("预期获得长链接url不为空，实际返回空")
		return
	}

	// 判断code写入缓存
	if maps.Equal(cache.cache, cacheSnapshot) {
		t.Error("没有写缓存")
		return
	}
	for k, v := range cache.cache {
		isNew := true
		for oldK := range cacheSnapshot {
			if k == oldK {
				isNew = false
				break
			}
		}
		if isNew {
			if v.value != redirectRes.LongUrl {
				t.Errorf("缓存写入long_url错误，期望%s，实际%s", redirectRes.LongUrl, v.value)
				return
			}
		}
	}
}

// 测试跳转，code在缓存中不存在，但是在数据库存在，然而已经过期
func TestRedirectExpired(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	cache := getTestCache()
	redirectSvc := &service.RedirectSvc{
		Cache:    cache,
		Database: db,
	}

	// 后续逻辑判断，缓存是否有变化
	cacheSnapshot := make(map[string]*cacheData)
	for k, v := range cache.cache {
		cacheSnapshot[k] = v
	}

	expiredCode := "BIkBDO"
	redirectRes, err := redirectSvc.Do(context.Background(), expiredCode)
	if err != nil {
		t.Error(err)
		return
	}

	if redirectRes.Redirect {
		t.Errorf("预期是否跳转为falese，实际返回%t", redirectRes.Redirect)
		return
	}

	if redirectRes.StatusCode != http.StatusNotFound {
		t.Errorf("期望返回http状态码404，实际返回%d\n", redirectRes.StatusCode)
		return
	}

	// 判断标识code为空的标识写入缓存中
	if maps.Equal(cache.cache, cacheSnapshot) {
		t.Error("没有写缓存")
		return
	}
	for k, v := range cache.cache {
		isNew := true
		for oldK := range cacheSnapshot {
			if k == oldK {
				isNew = false
				break
			}
		}
		if isNew {
			if v.value != config.Conf.Core.CacheNotFoundValue {
				t.Errorf("缓存写入long_url错误，期望%s，实际%s", config.Conf.Core.CacheNotFoundValue, v.value)
				return
			}
		}
	}
}

// 测试跳转，code在数据库和缓存中都不存在
func TestRedirectNotExist(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	cache := getTestCache()
	redirectSvc := &service.RedirectSvc{
		Cache:    cache,
		Database: db,
	}

	// 后续逻辑判断，缓存是否有变化
	cacheSnapshot := make(map[string]*cacheData)
	for k, v := range cache.cache {
		cacheSnapshot[k] = v
	}

	notExistCode := "OhMyPi"
	redirectRes, err := redirectSvc.Do(context.Background(), notExistCode)
	if err != nil {
		t.Error(err)
		return
	}

	if redirectRes.Redirect {
		t.Errorf("预期是否跳转为falese，实际返回%t", redirectRes.Redirect)
		return
	}

	if redirectRes.StatusCode != http.StatusNotFound {
		t.Errorf("期望返回http状态码404，实际返回%d\n", redirectRes.StatusCode)
		return
	}

	// 判断标识code为空的标识写入缓存中
	if maps.Equal(cache.cache, cacheSnapshot) {
		t.Error("没有写缓存")
		return
	}
	for k, v := range cache.cache {
		isNew := true
		for oldK := range cacheSnapshot {
			if k == oldK {
				isNew = false
				break
			}
		}
		if isNew {
			if v.value != config.Conf.Core.CacheNotFoundValue {
				t.Errorf("缓存写入long_url错误，期望%s，实际%s", redirectRes.LongUrl, v.value)
				return
			}
		}
	}

}

// 测试跳转，code在缓存中标识为不存在
func TestRedirectCachedNotFound(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	cache := getTestCache()
	redirectSvc := &service.RedirectSvc{
		Cache:    cache,
		Database: db,
	}

	// 后续逻辑判断，缓存是否有变化
	cacheSnapshot := make(map[string]*cacheData)
	for k, v := range cache.cache {
		cacheSnapshot[k] = v
	}

	markedNoFoundCode := "empty0"
	redirectRes, err := redirectSvc.Do(context.Background(), markedNoFoundCode)
	if err != nil {
		t.Error(err)
		return
	}

	if redirectRes.Redirect {
		t.Errorf("预期是否跳转为falese，实际返回%t", redirectRes.Redirect)
		return
	}

	if redirectRes.StatusCode != http.StatusNotFound {
		t.Errorf("期望返回http状态码404，实际返回%d\n", redirectRes.StatusCode)
		return
	}

	// 判断标识code在缓存中已存在，不对缓存继续操作
	if !maps.Equal(cache.cache, cacheSnapshot) {
		t.Error("期望不写缓存，实际写了")
		return
	}
	for k, v := range cache.cache {
		isNew := true
		for oldK := range cacheSnapshot {
			if k == oldK {
				isNew = false
				break
			}
		}
		if isNew {
			t.Errorf("缓存写入新值%s，实际不应该有值写入", v.value)
			return

		}
	}
}
