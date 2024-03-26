package service

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"testing"
	"time"

	"gopkg.in/h2non/gock.v1"

	"github.com/1911860538/short_link/app/server/service"
	"github.com/1911860538/short_link/config"
)

// 测试长链接url非法
func TestAddLinkInvalidUrl(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	// 使用mockDb作为DatabaseItf
	addLinkSvc := &service.AddLinkSvc{
		Database: db,
	}

	invalidUrls := []string{
		"Hello, world!",
		"httq://www.qq.com",
		"httqs://www.google.cn",
		"ftp://not-http.com", // 是合法的url，但不是http或者https协议
		"https:1233",
	}

	for _, url := range invalidUrls {
		params := service.AddLinkParams{
			UserId:   "user_id_1",
			LongUrl:  url,
			Deadline: time.Time{},
		}
		addRes, err := addLinkSvc.Do(context.Background(), params)
		if err != nil {
			t.Error(err)
			return
		}
		if addRes.StatusCode != http.StatusBadRequest {
			t.Errorf("添加url（%s）期望返回http状态码400，实际返回%d\n", url, addRes.StatusCode)
			return
		}
	}
}

// 测试长链接请求结果响应码不是200
func TestAddLinkUrlRespErr(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	// 使用mockDb作为DatabaseItf
	addLinkSvc := &service.AddLinkSvc{
		Database: db,
	}

	params := service.AddLinkParams{
		UserId:   "user_id_1",
		LongUrl:  "https://www.longurl0.com",
		Deadline: time.Time{},
	}
	// mock http
	gock.New(params.LongUrl).Reply(http.StatusNotFound)
	addRes, err := addLinkSvc.Do(context.Background(), params)
	if err != nil {
		t.Error(err)
		return
	}
	if addRes.StatusCode != http.StatusBadRequest {
		t.Errorf("期望返回http状态码400，实际返回%d\n", addRes.StatusCode)
		return
	}
}

// 测试添加短链接数据，但是数据已存在
func TestAddLinkExisted(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	// 使用mockDb作为DatabaseItf
	addLinkSvc := &service.AddLinkSvc{
		Database: db,
	}
	params := service.AddLinkParams{
		UserId:   "user_id_1",
		LongUrl:  "https://www.longurl2.com",
		Deadline: time.Time{},
	}
	gock.New(params.LongUrl).Reply(http.StatusOK)
	addRes, err := addLinkSvc.Do(context.Background(), params)
	if err != nil {
		t.Error(err)
		return
	}
	if addRes.StatusCode != http.StatusConflict {
		t.Errorf("期望返回http状态码409，实际返回%d\n", addRes.StatusCode)
		return
	}
}

// 测试code冲突，生成code加盐
func TestAddLinkGenCodeSalt(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	addLinkSvc := &service.AddLinkSvc{
		Database: db,
	}
	// 数据userId和longUrl与link1一样，所以期望的生成code也一样，唯一索引冲突，期望触发GenCode加盐操作
	params := service.AddLinkParams{
		UserId:   "user_id_1",
		LongUrl:  "https://www.longurl1.com",
		Deadline: time.Time{},
	}
	gock.New(params.LongUrl).Reply(http.StatusOK)
	addRes, err := addLinkSvc.Do(context.Background(), params)
	if err != nil {
		t.Error(err)
		return
	}
	if addRes.StatusCode != http.StatusCreated {
		t.Errorf("期望返回http状态码201，实际返回%d\n", addRes.StatusCode)
		return
	}

	addedLink := db.latestLink()
	if addedLink.Salt == "" {
		t.Error("同一个user_id和long_url，数据库存在已过期但未删除，code冲突，期望生成新的code触发加盐，实际未加盐")
		return
	}
}

// 测试正常保存
func TestAddLink(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	addLinkSvc := &service.AddLinkSvc{
		Database: db,
	}

	// 正常保存，有过期时间
	{
		deadline := time.Now().Add(time.Duration(20) * time.Hour)
		params := service.AddLinkParams{
			UserId:   "user_id_1",
			LongUrl:  "https://www.longurl10000.com",
			Deadline: deadline,
		}
		gock.New(params.LongUrl).Reply(http.StatusOK)
		addRes, err := addLinkSvc.Do(context.Background(), params)
		if err != nil {
			t.Error(err)
			return
		}
		if addRes.StatusCode != http.StatusCreated {
			t.Errorf("期望返回http状态码201，实际返回%d\n", addRes.StatusCode)
			return
		}

		addedLink := db.latestLink()
		if addedLink.Expired() {
			t.Error("保存的有期限的link已过期")
			return
		}

		if addedLink.Deadline != deadline {
			t.Error("保存的有期限的link，数据库数据与输入不一致")
			return
		}

		expectedTtl := deadline.Add(time.Duration(config.Conf.Core.ExpiredKeepDays*24) * time.Hour)
		if addedLink.TtlTime != expectedTtl {
			t.Errorf("保存的有期限的link，期望ttlTime为%v实际为%v", expectedTtl, addedLink.TtlTime)
			return
		}
	}

	// 正常保存，无过期时间
	{
		params := service.AddLinkParams{
			UserId:   "user_id_1",
			LongUrl:  "https://www.longurl10001.com",
			Deadline: time.Time{},
		}
		gock.New(params.LongUrl).Reply(http.StatusOK)
		addRes, err := addLinkSvc.Do(context.Background(), params)
		if err != nil {
			t.Error(err)
			return
		}
		if addRes.StatusCode != http.StatusCreated {
			t.Errorf("期望返回http状态码201，实际返回%d\n", addRes.StatusCode)
			return
		}

		addedLink := db.latestLink()
		if addedLink.Expired() {
			t.Error("保存的无期限的link已过期")
			return
		}

		if !addedLink.Deadline.IsZero() {
			t.Error("保存的无期限的link，数据库数据deadline不是零值")
			return
		}

		if !addedLink.TtlTime.IsZero() {
			t.Errorf("保存的无期限的link，期望ttlTime为时间零值实际为%v", addedLink.TtlTime)
			return
		}
	}
}

// 测试生成code长度为配置长度，且为字符或数字
func FuzzGenCode(f *testing.F) {
	params := []struct {
		userId  string
		longUrl string
		salt    string
	}{
		{
			userId:  "fake_user_id_0",
			longUrl: "http://www.abc.com",
			salt:    "",
		},
		{
			userId:  "fake_user_id_2",
			longUrl: "https://www.liwenzhou.com/posts/Go/unit-test-1/",
			salt:    strconv.FormatInt(time.Now().UnixMilli(), 10),
		},
	}

	// 正则表达式，匹配大小写字母和数字
	pattern := `^[a-zA-Z0-9]+$`
	re := regexp.MustCompile(pattern)

	for _, param := range params {
		f.Add(param.userId, param.longUrl, param.salt)
	}

	f.Fuzz(func(t *testing.T, userId string, longUrl string, salt string) {
		code, err := service.GenCode(userId, longUrl, salt)
		if err != nil {
			t.Errorf("userId(%s),url(%s),salt(%s)生成code失败：%v", userId, longUrl, salt, err)
			return
		}
		if len(code) != config.Conf.Core.CodeLen {
			t.Errorf("userId(%s),url(%s),salt(%s)生成code为%s，长度不是%d", userId, longUrl, salt, code, config.Conf.Core.CodeLen)
			return
		}
		if !re.MatchString(code) {
			t.Errorf("生成code为%s，不是大小写字母或数字组合", code)
			return
		}
		return
	})

}

// 测试多次生成的code一样
func FuzzGenCodeSame(f *testing.F) {
	userIds := []string{
		"fake_user_id_0",
		"fake_user_id_1",
		"fake_user_id_2",
	}
	urls := []string{
		"http://www.abc.com",
		"https://www.liwenzhou.com/posts/Go/unit-test-1/",
		"https://www.baidu.com",
	}
	salts := []string{
		"",
		strconv.FormatInt(time.Now().UnixMilli(), 10),
	}

	for _, userId := range userIds {
		for _, url := range urls {
			for _, salt := range salts {
				f.Add(userId, url, salt)
			}
		}
	}

	f.Fuzz(func(t *testing.T, userId string, longUrl string, salt string) {
		code0, err := service.GenCode(userId, longUrl, salt)
		if err != nil {
			t.Errorf("userId(%s),url(%s),salt(%s)生成code失败：%v", userId, longUrl, salt, err)
			return
		}

		code1, err := service.GenCode(userId, longUrl, salt)
		if err != nil {
			t.Errorf("userId(%s),url(%s),salt(%s)生成code失败：%v", userId, longUrl, salt, err)
			return
		}

		code2, err := service.GenCode(userId, longUrl, salt)
		if err != nil {
			t.Errorf("userId(%s),url(%s),salt(%s)生成code失败：%v", userId, longUrl, salt, err)
			return
		}

		if !(code0 == code1 && code1 == code2) {
			t.Errorf("三次生成的code不一致，code0=%s，code1=%s，code2=%s", code0, code1, code2)
			return
		}

		return
	})

}
