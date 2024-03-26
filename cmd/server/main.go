package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/1911860538/short_link/app/component"
	"github.com/1911860538/short_link/app/server"
	"github.com/1911860538/short_link/config"
)

func main() {
	// 设置真正可用的cup核数
	_, _ = maxprocs.Set(maxprocs.Logger(nil))

	// 组件初始化（数据库、缓存等连接）
	if err := component.Startup(); err != nil {
		log.Fatal(err)
	}

	errChan := make(chan error)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// http server
	engine := gin.Default()
	if config.Conf.Debug {
		gin.SetMode(gin.DebugMode)
		pprof.Register(engine)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	server.Route(engine)
	srv := &http.Server{
		Addr:        fmt.Sprintf(":%d", config.Conf.Server.Port),
		Handler:     engine,
		IdleTimeout: time.Duration(config.Conf.Server.IdleTimeoutSeconds) * time.Second,
	}
	srv.RegisterOnShutdown(component.Shutdown)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	shutdownFunc := func(httpSrv *http.Server) {
		log.Println("程序退出")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		httpSrv.Shutdown(ctx)
	}

	// 监控启动错误、程序退出
	select {
	case err := <-errChan:
		shutdownFunc(srv)
		log.Fatalf("启动服务失败：%v\n", err)
	case <-stopChan:
		shutdownFunc(srv)
	}
}
