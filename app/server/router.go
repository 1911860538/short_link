package server

import (
	"github.com/gin-gonic/gin"

	"github.com/1911860538/short_link/app/server/handler"
	"github.com/1911860538/short_link/app/server/middleware"
)

// Route 路由注册
func Route(e *gin.Engine) {
	// 核心，跳转服务
	e.GET("/:code", handler.RedirectHandler)

	// 短链接管理
	linKGroupV1 := e.Group("/api/v1/links")
	linKGroupV1.Use(middleware.JwtMiddleware)
	{
		linKGroupV1.POST("", handler.AddLinkHandler)
		linKGroupV1.GET("", handler.GetLinkHandler)
	}
}
