package setting

import (
	apiSetting "github.com/LychApe/LynxPilot/internal/api/setting"
	"github.com/LychApe/LynxPilot/internal/middleware"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	group := router.Group("/api/v1/private/setting")
	group.Use(middleware.Auth())
	{
		group.GET("/docker/connection", apiSetting.GetDockerConnectionHandler)
		group.PUT("/docker/connection", apiSetting.SaveDockerConnectionHandler)
		group.POST("/docker/connection/test", apiSetting.TestDockerConnectionHandler)
	}
}
