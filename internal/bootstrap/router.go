package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	routeServer "github.com/LychApe/LynxPilot/internal/router/server"
	routeUser "github.com/LychApe/LynxPilot/internal/router/user"
	"github.com/LychApe/LynxPilot/internal/utils/logger"
	"github.com/LychApe/LynxPilot/web"
	"github.com/gin-gonic/gin"
)

func LoadRouter(config *Config) *gin.Engine {
	// 设置gin模式
	gin.SetMode(config.Server.Mode)

	// 创建gin引擎
	router := gin.Default()

	// 注入全局上下文
	router.Use(func(c *gin.Context) {
		c.Set("db", DB)
		c.Set("tokenSalt", config.Auth.TokenSalt)
		c.Set("requestId", uuid.New().String())
		c.Next()
	})

	// 注册路由
	routeServer.Register(router)
	routeUser.Register(router)

	// 前端静态文件服务
	loadRouterServeFrontend(router)

	loadRouterServer(router, config)

	return router
}

func loadRouterServer(router *gin.Engine, config *Config) {
	server := loadRouterCreateHTTPServer(router, config.Server.Port)
	loadRouterWatchShutdownSignal(server)
	loadRouterStartHTTPServer(server)
}

func loadRouterCreateHTTPServer(router *gin.Engine, port int) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
}

func loadRouterWatchShutdownSignal(server *http.Server) {
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
}

func loadRouterStartHTTPServer(server *http.Server) {
	// 启动服务
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Errorf("启动服务失败: %v", err)
		os.Exit(1)
	}
}

func loadRouterServeFrontend(router *gin.Engine) {
	distFS, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		logger.Errorf("加载前端资源失败: %v", err)
		return
	}

	fileServer := http.FileServer(http.FS(distFS))

	// 静态资源（带文件扩展名的请求）
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 请求返回 404
		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "not found"})
			return
		}

		// 尝试匹配静态文件
		if strings.Contains(path, ".") {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// SPA fallback：其他路径返回 index.html
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
