package docker

import (
	"net/http"

	dockerService "github.com/LychApe/LynxPilot/internal/service/docker"
	"github.com/LychApe/LynxPilot/internal/utils/response"
	"github.com/gin-gonic/gin"
)

func ListComposeProjectsHandler(c *gin.Context) {
	projects, err := dockerService.ListComposeProjects()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	if projects == nil {
		projects = []dockerService.ComposeProject{}
	}
	response.OK(c, projects)
}

func ComposeUpHandler(c *gin.Context) {
	var body struct {
		Content     string `json:"content"`
		ProjectName string `json:"project_name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}

	if body.Content == "" {
		response.Error(c, http.StatusBadRequest, 400, "compose 内容不能为空")
		return
	}

	if err := dockerService.ComposeUpFromContent(body.Content, body.ProjectName); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Compose 项目已启动"})
}

func ComposeDownHandler(c *gin.Context) {
	projectName := c.Param("name")
	var body struct {
		RemoveVolumes bool `json:"remove_volumes"`
	}
	c.ShouldBindJSON(&body)

	if err := dockerService.ComposeDown(projectName, body.RemoveVolumes); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Compose 项目已停止并移除"})
}

func ComposeRestartHandler(c *gin.Context) {
	projectName := c.Param("name")
	if err := dockerService.ComposeRestart(projectName); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Compose 项目已重启"})
}

func ComposeStopHandler(c *gin.Context) {
	projectName := c.Param("name")
	if err := dockerService.ComposeStop(projectName); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Compose 项目已停止"})
}

func ComposeStartHandler(c *gin.Context) {
	projectName := c.Param("name")
	if err := dockerService.ComposeStart(projectName); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Compose 项目已启动"})
}

func ComposeLogsHandler(c *gin.Context) {
	projectName := c.Param("name")
	tail := c.DefaultQuery("tail", "100")

	logs, err := dockerService.ComposeLogs(projectName, tail)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	response.OK(c, gin.H{"logs": logs})
}

func ComposePsHandler(c *gin.Context) {
	projectName := c.Param("name")

	containers, err := dockerService.ComposePs(projectName)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	if containers == nil {
		containers = []dockerService.ContainerInfo{}
	}
	response.OK(c, containers)
}

func ComposeAvailableHandler(c *gin.Context) {
	available := dockerService.CheckComposeAvailable()
	response.OK(c, gin.H{"available": available})
}
