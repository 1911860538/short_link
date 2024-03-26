package service

import (
	"context"
	"net/http"
	"time"

	"github.com/1911860538/short_link/app/component"
)

type GetLinkSvc struct {
	Database component.DatabaseItf
}

type GetLinkParams struct {
	UserId  string
	Code    string
	LongUrl string
}

type GetLinkRes struct {
	StatusCode int
	Msg        string

	Code     string
	LongUrl  string
	Deadline time.Time
}

func (s *GetLinkSvc) Do(ctx context.Context, params GetLinkParams) (GetLinkRes, error) {
	filter := make(map[string]any)
	if params.UserId != "" {
		filter["user_id"] = params.UserId
	}
	if params.Code != "" {
		filter["code"] = params.Code
	}
	if params.LongUrl != "" {
		filter["long_url"] = params.LongUrl
	}

	link, err := s.Database.Get(ctx, filter)
	if err != nil {
		return GetLinkRes{
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	if link == nil {
		return GetLinkRes{
			StatusCode: http.StatusNotFound,
			Msg:        "数据不存在",
		}, nil
	}

	return GetLinkRes{
		StatusCode: http.StatusOK,
		Code:       link.Code,
		LongUrl:    link.LongUrl,
		Deadline:   link.Deadline,
	}, nil
}
