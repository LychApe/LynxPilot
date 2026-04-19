package server

import (
	apiServer "github.com/LychApe/LynxPilot/internal/api/server"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	serverPublicGroup := router.Group("/api/v1/public/server")
	{
		serverPublicGroup.GET("/reboot", apiServer.RebootHandler)
		serverPublicGroup.GET("/shutdown", apiServer.ShutdownHandler)
	}
}
