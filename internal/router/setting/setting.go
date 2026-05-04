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

		group.GET("/container/defaults", apiSetting.GetContainerDefaultsHandler)
		group.PUT("/container/defaults", apiSetting.SaveContainerDefaultsHandler)

		group.GET("/ui/prefs", apiSetting.GetUIPrefsHandler)
		group.PUT("/ui/prefs", apiSetting.SaveUIPrefsHandler)

		group.GET("/all", apiSetting.GetAllSettingsHandler)
	}
}
