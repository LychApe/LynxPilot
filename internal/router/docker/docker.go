package docker

import (
	apiDocker "github.com/LychApe/LynxPilot/internal/api/docker"
	"github.com/LychApe/LynxPilot/internal/middleware"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	group := router.Group("/api/v1/private/docker")
	group.Use(middleware.Auth())
	{
		group.GET("/ping", apiDocker.PingHandler)

		group.GET("/containers", apiDocker.ListContainersHandler)
		group.GET("/containers/search", apiDocker.SearchContainersHandler)
		group.GET("/containers/:id", apiDocker.GetContainerDetailHandler)
		group.GET("/containers/:id/stats", apiDocker.GetContainerStatsHandler)
		group.GET("/containers/:id/logs", apiDocker.GetContainerLogsHandler)
		group.POST("/containers/:id/start", apiDocker.StartContainerHandler)
		group.POST("/containers/:id/stop", apiDocker.StopContainerHandler)
		group.POST("/containers/:id/restart", apiDocker.RestartContainerHandler)
		group.DELETE("/containers/:id", apiDocker.RemoveContainerHandler)

		group.GET("/networks", apiDocker.ListNetworksHandler)
		group.POST("/networks", apiDocker.CreateNetworkHandler)
		group.GET("/networks/:id", apiDocker.InspectNetworkHandler)
		group.DELETE("/networks/:id", apiDocker.RemoveNetworkHandler)
		group.POST("/networks/:id/connect", apiDocker.ConnectContainerHandler)
		group.POST("/networks/:id/disconnect", apiDocker.DisconnectContainerHandler)

		group.GET("/compose/available", apiDocker.ComposeAvailableHandler)
		group.GET("/compose/projects", apiDocker.ListComposeProjectsHandler)
		group.POST("/compose/up", apiDocker.ComposeUpHandler)
		group.POST("/compose/:name/down", apiDocker.ComposeDownHandler)
		group.POST("/compose/:name/restart", apiDocker.ComposeRestartHandler)
		group.POST("/compose/:name/stop", apiDocker.ComposeStopHandler)
		group.POST("/compose/:name/start", apiDocker.ComposeStartHandler)
		group.GET("/compose/:name/logs", apiDocker.ComposeLogsHandler)
		group.GET("/compose/:name/ps", apiDocker.ComposePsHandler)

		group.GET("/images", apiDocker.ListImagesHandler)
	}
}
