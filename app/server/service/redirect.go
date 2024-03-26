package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/1911860538/short_link/app/component"
	"github.com/1911860538/short_link/config"
)

type RedirectSvc struct {
	Cache    component.CacheItf
	Database component.DatabaseItf
}

type RedirectRes struct {
	StatusCode int
	Msg        string

	Redirect bool
	LongUrl  string
}

var (
	confRedirectStatusCode = config.Conf.Core.RedirectStatusCode
	confCodeTtl            = config.Conf.Core.CodeTtl
	confCodeLen            = config.Conf.Core.CodeLen
	confCacheNotFoundValue = config.Conf.Core.CacheNotFoundValue

	sfGroup singleflight.Group
)

func (s *RedirectSvc) Do(ctx context.Context, code string) (RedirectRes, error) {
	if len(code) != confCodeLen {
		return s.notFound(code)
	}

	longUrl, err := s.Cache.Get(ctx, code)
	if err != nil {
		return s.internalErr(err)
	}

	// confCacheNotFoundValue，用来在缓存标识某个code不存在，防止缓存穿透
	// 防止当某个code在数据库和缓存都不存在，大量无用请求反复读取缓存和数据库
	if longUrl == confCacheNotFoundValue {
		return s.notFound(code)
	}

	if longUrl != "" {
		return s.redirect(longUrl)
	}

	// 使用singleflight防止缓存击穿
	// 防止某个code缓存过期，大量该code请求过来，造成全部请求去数据读值
	result, err, _ := sfGroup.Do(code, func() (any, error) {
		link, err := s.getLinkSetCache(ctx, code)
		if err != nil {
			return nil, err
		}
		return link, nil
	})

	if err != nil {
		return s.internalErr(err)
	}
	if result == nil {
		return s.notFound(code)
	}
	link, ok := result.(*component.Link)
	if !ok {
		err := fmt.Errorf("singleflight group.Do返回值%v，类型错误，非*component.Link", result)
		return s.internalErr(err)
	}
	if link == nil {
		return s.notFound(code)
	}

	return s.redirect(link.LongUrl)
}

func (s *RedirectSvc) getLinkSetCache(ctx context.Context, code string) (*component.Link, error) {
	filter := map[string]any{
		"code": code,
	}
	link, err := s.Database.Get(ctx, filter)
	if err != nil {
		return nil, err
	}

	if link == nil || link.LongUrl == "" || link.Expired() {
		if err := s.Cache.Set(ctx, code, confCacheNotFoundValue, confCodeTtl); err != nil {
			return nil, err
		}

		return nil, nil
	}

	var ttl int
	if link.Deadline.IsZero() {
		ttl = confCodeTtl
	} else {
		if remainSeconds := int(link.Deadline.Sub(time.Now().UTC()).Seconds()); remainSeconds < confCodeTtl {
			ttl = remainSeconds
		} else {
			ttl = confCodeTtl
		}
	}
	if err := s.Cache.Set(ctx, code, link.LongUrl, ttl); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *RedirectSvc) notFound(code string) (RedirectRes, error) {
	return RedirectRes{
		StatusCode: http.StatusNotFound,
		Msg:        fmt.Sprintf("短链接(%s)无对应的长链接地址", code),
	}, nil
}

func (s *RedirectSvc) redirect(longUrl string) (RedirectRes, error) {
	return RedirectRes{
		StatusCode: confRedirectStatusCode,
		Redirect:   true,
		LongUrl:    longUrl,
	}, nil
}

func (s *RedirectSvc) internalErr(err error) (RedirectRes, error) {
	return RedirectRes{
		StatusCode: http.StatusInternalServerError,
	}, err
}
