package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/1911860538/short_link/app/component"
	"github.com/1911860538/short_link/app/server/service"
)

var addLinkSvc = service.AddLinkSvc{
	Database: component.Database,
}

type addLinkForm struct {
	LongUrl      string `json:"long_url" binding:"required,url"`
	DeadlineUnix int64  `json:"deadline_unix" binding:"required"`
}

type addLinkResp struct {
	Code string `json:"code"`
}

func AddLinkHandler(c *gin.Context) {
	jwtClaims, ok := getJwtClaims(c)
	if !ok {
		slog.Error("GetLinkHandler无法获取jwt信息")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"detail": "服务内部错误",
		})
		return
	}

	var addLinkForm addLinkForm
	if err := c.ShouldBindJSON(&addLinkForm); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"detail": err.Error(),
		})
		return
	}

	// 当过期时间戳<=0，约定为不过期
	var deadline time.Time
	if addLinkForm.DeadlineUnix <= 0 {
		deadline = time.Time{}
	} else {
		deadline = time.Unix(addLinkForm.DeadlineUnix, 0)
	}
	addLinkParams := service.AddLinkParams{
		UserId:   jwtClaims.Id,
		LongUrl:  addLinkForm.LongUrl,
		Deadline: deadline,
	}
	res, err := addLinkSvc.Do(c.Request.Context(), addLinkParams)
	if err != nil {
		slog.Error("AddLinkHandler错误", "err", err)
		c.AbortWithStatusJSON(res.StatusCode, gin.H{
			"detail": "服务内部错误",
		})
		return
	}

	if res.StatusCode != http.StatusCreated {
		c.AbortWithStatusJSON(res.StatusCode, gin.H{
			"detail": res.Msg,
		})
		return
	}

	c.JSON(http.StatusCreated, addLinkResp{
		Code: res.Code,
	})
}
