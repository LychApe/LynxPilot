package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	routeServer "github.com/LychApe/LynxPilot/internal/router/server"
	"github.com/LychApe/LynxPilot/internal/utils/logger"
	"github.com/gin-gonic/gin"
)

func LoadRouter(config *Config) *gin.Engine {
	// 设置gin模式
	gin.SetMode(config.Server.Mode)

	// 创建gin引擎
	router := gin.Default()

	// 注册路由
	routeServer.Register(router)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Server.Port),
		Handler: router,
	}

	// 接收系统信号，执行优雅停止
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		logger.Infof("收到停止信号，开始优雅关闭服务")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Errorf("服务优雅关闭失败: %v", err)
		}
	}()

	// 启动服务
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Errorf("启动服务失败: %v", err)
		os.Exit(1)
	}

	return router
}
