package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/1911860538/short_link/app/component"
	"github.com/1911860538/short_link/app/server/service"
)

var getLinkSvc = service.GetLinkSvc{
	Database: component.Database,
}

type getLinkResp struct {
	Code         string `json:"code"`
	LongUrl      string `json:"long_url"`
	DeadlineUnix int64  `json:"deadline_unix" binding:"required"`
}

func GetLinkHandler(c *gin.Context) {
	jwtClaims, ok := getJwtClaims(c)
	if !ok {
		slog.Error("GetLinkHandler无法获取jwt信息")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"detail": "服务内部错误",
		})
		return
	}

	code := c.Query("code")
	longUrl := c.Query("long_url")
	if code == "" && longUrl == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"detail": "code和long_url参数不能均为空",
		})
		return
	}

	getLinkParams := service.GetLinkParams{
		UserId:  jwtClaims.Id,
		Code:    code,
		LongUrl: longUrl,
	}
	res, err := getLinkSvc.Do(c.Request.Context(), getLinkParams)
	if err != nil {
		slog.Error("GetLinkHandler错误", "err", err)
		c.AbortWithStatusJSON(res.StatusCode, gin.H{
			"detail": "服务内部错误",
		})
		return
	}

	if res.StatusCode != http.StatusOK {
		c.AbortWithStatusJSON(res.StatusCode, gin.H{
			"detail": res.Msg,
		})
		return
	}

	var deadlineUnix int64
	if res.Deadline.IsZero() {
		deadlineUnix = 0
	} else {
		deadlineUnix = res.Deadline.Unix()
	}
	c.JSON(http.StatusOK, getLinkResp{
		Code:         res.Code,
		LongUrl:      res.LongUrl,
		DeadlineUnix: deadlineUnix,
	})
}
