package server

import (
	apiServer "github.com/LychApe/LynxPilot/internal/api/server"
	"github.com/LychApe/LynxPilot/internal/middleware"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	// 公共接口
	serverPublicGroup := router.Group("/api/v1/public/server")
	{
		serverPublicGroup.GET("/status", apiServer.StatusHandler)
	}

	// 私有接口
	serverGroup := router.Group("/api/v1/private/server")
	serverGroup.Use(middleware.Auth())
	{
		serverGroup.GET("/reboot", apiServer.RebootHandler)
		serverGroup.GET("/shutdown", apiServer.ShutdownHandler)
	}
}
