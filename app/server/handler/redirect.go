package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/1911860538/short_link/app/component"
	"github.com/1911860538/short_link/app/server/service"
)

var redirectSvc = service.RedirectSvc{
	Cache:    component.Cache,
	Database: component.Database,
}

func RedirectHandler(c *gin.Context) {
	code := c.Param("code")

	res, err := redirectSvc.Do(c.Request.Context(), code)
	if err != nil {
		slog.Error("GetLinkHandler错误", "err", err)
		c.AbortWithStatusJSON(res.StatusCode, gin.H{
			"detail": "服务内部错误",
		})
		return
	}

	if !res.Redirect {
		c.AbortWithStatusJSON(res.StatusCode, gin.H{
			"detail": res.Msg,
		})
		return
	}

	c.Redirect(res.StatusCode, res.LongUrl)
}
