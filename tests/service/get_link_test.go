package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/1911860538/short_link/app/server/service"
)

// 测试没有找到数据
func TestGetLink404(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	getLinkSvc := &service.GetLinkSvc{
		Database: db,
	}

	notExistedLinkParamsList := []service.GetLinkParams{
		{
			UserId: "user_id_1",
			Code:   "not_exist_code_1",
		},
		{
			UserId:  "user_id_1",
			LongUrl: "https://www.notexistsite.com",
		},
		{
			UserId:  "user_id_1",
			Code:    "not_exist_code_1",
			LongUrl: "https://www.notexistsite.com",
		},
	}

	for _, notExistedLinkParams := range notExistedLinkParamsList {
		getRes, err := getLinkSvc.Do(context.Background(), notExistedLinkParams)
		if err != nil {
			t.Error(err)
			return
		}
		if getRes.StatusCode != http.StatusNotFound {
			t.Errorf("期望返回http code 404，实际为%d", getRes.StatusCode)
			return
		}
	}
}

// 测试，数据有数据
func TestGetLinkExpired(t *testing.T) {
	db, err := getTestDb()
	if err != nil {
		t.Error(err)
		return
	}

	getLinkSvc := &service.GetLinkSvc{
		Database: db,
	}

	expiredLinkParamsList := []service.GetLinkParams{
		{
			UserId: "user_id_1",
			Code:   "BIkBDO",
		},
		{
			UserId:  "user_id_1",
			LongUrl: "https://www.longurl1.com",
		},
		{
			UserId:  "user_id_1",
			Code:    "BIkBDO",
			LongUrl: "https://www.longurl1.com",
		},
	}

	for _, expiredLinkParams := range expiredLinkParamsList {
		getRes, err := getLinkSvc.Do(context.Background(), expiredLinkParams)
		if err != nil {
			t.Error(err)
			return
		}
		if getRes.StatusCode != http.StatusOK {
			t.Errorf("期望返回http code 200，实际为%d", getRes.StatusCode)
			return
		}
	}
}
