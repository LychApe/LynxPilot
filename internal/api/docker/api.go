package docker

import (
	"net/http"

	dockerService "github.com/LychApe/LynxPilot/internal/service/docker"
	"github.com/LychApe/LynxPilot/internal/utils/response"
	"github.com/gin-gonic/gin"
)

func PingHandler(c *gin.Context) {
	if err := dockerService.CheckDockerAvailable(); err != nil {
		response.Error(c, http.StatusServiceUnavailable, 503, "Docker 不可用: "+err.Error())
		return
	}

	response.OK(c, gin.H{
		"available":         true,
		"custom_connection": dockerService.IsCustomConnection(),
		"host":              dockerService.GetActiveHost(),
	})
}

func ListContainersHandler(c *gin.Context) {
	all := c.Query("all") == "true"

	containers, err := dockerService.ListContainers(all)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	if containers == nil {
		containers = []dockerService.ContainerInfo{}
	}

	response.OK(c, containers)
}

func GetContainerDetailHandler(c *gin.Context) {
	containerID := c.Param("id")

	detail, err := dockerService.GetContainerDetail(containerID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	response.OK(c, detail)
}

func GetContainerStatsHandler(c *gin.Context) {
	containerID := c.Param("id")

	stats, err := dockerService.GetContainerStats(containerID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	response.OK(c, stats)
}

func StartContainerHandler(c *gin.Context) {
	containerID := c.Param("id")

	if err := dockerService.StartContainer(containerID); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "启动容器失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "容器已启动"})
}

func StopContainerHandler(c *gin.Context) {
	containerID := c.Param("id")

	if err := dockerService.StopContainer(containerID); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "停止容器失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "容器已停止"})
}

func RestartContainerHandler(c *gin.Context) {
	containerID := c.Param("id")

	if err := dockerService.RestartContainer(containerID); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "重启容器失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "容器已重启"})
}

func RemoveContainerHandler(c *gin.Context) {
	containerID := c.Param("id")
	force := c.Query("force") == "true"

	if err := dockerService.RemoveContainer(containerID, force); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "删除容器失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "容器已删除"})
}

func GetContainerLogsHandler(c *gin.Context) {
	containerID := c.Param("id")
	tail := c.DefaultQuery("tail", "100")

	logs, err := dockerService.GetContainerLogs(containerID, tail)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	response.OK(c, gin.H{"logs": logs})
}

func ListImagesHandler(c *gin.Context) {
	images, err := dockerService.ListImages()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	response.OK(c, images)
}

func SearchContainersHandler(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		ListContainersHandler(c)
		return
	}

	containers, err := dockerService.SearchContainersByName(name)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	if containers == nil {
		containers = []dockerService.ContainerInfo{}
	}

	response.OK(c, containers)
}
