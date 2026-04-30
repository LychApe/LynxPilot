package file

import (
	apiFile "github.com/LychApe/LynxPilot/internal/api/file"
	"github.com/LychApe/LynxPilot/internal/middleware"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	group := router.Group("/api/v1/private/file")
	group.Use(middleware.Auth())
	{
		group.GET("/list", apiFile.ListHandler)
		group.GET("/info", apiFile.GetFileInfoHandler)
		group.GET("/read", apiFile.ReadFileHandler)
		group.POST("/save", apiFile.SaveFileHandler)
		group.POST("/mkdir", apiFile.CreateDirHandler)
		group.POST("/touch", apiFile.CreateFileHandler)
		group.POST("/delete", apiFile.DeleteHandler)
		group.POST("/rename", apiFile.RenameHandler)
		group.POST("/upload", apiFile.UploadHandler)
		group.GET("/download", apiFile.DownloadHandler)
		group.GET("/base-path", apiFile.GetBasePathHandler)
		group.PUT("/base-path", apiFile.SetBasePathHandler)
	}
}
