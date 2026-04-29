package docker

import (
	"net/http"

	dockerService "github.com/LychApe/LynxPilot/internal/service/docker"
	"github.com/LychApe/LynxPilot/internal/utils/response"
	"github.com/gin-gonic/gin"
)

func ListNetworksHandler(c *gin.Context) {
	networks, err := dockerService.ListNetworks()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	if networks == nil {
		networks = []dockerService.NetworkInfo{}
	}
	response.OK(c, networks)
}

func CreateNetworkHandler(c *gin.Context) {
	var req dockerService.CreateNetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}

	if req.Name == "" {
		response.Error(c, http.StatusBadRequest, 400, "网络名称不能为空")
		return
	}

	net, err := dockerService.CreateNetwork(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	response.OK(c, net)
}

func RemoveNetworkHandler(c *gin.Context) {
	networkID := c.Param("id")
	if err := dockerService.RemoveNetwork(networkID); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "删除网络失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"message": "网络已删除"})
}

func InspectNetworkHandler(c *gin.Context) {
	networkID := c.Param("id")
	net, err := dockerService.InspectNetwork(networkID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, net)
}

func ConnectContainerHandler(c *gin.Context) {
	networkID := c.Param("id")
	var body struct {
		ContainerID string `json:"container_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.ContainerID == "" {
		response.Error(c, http.StatusBadRequest, 400, "container_id 不能为空")
		return
	}

	if err := dockerService.ConnectContainer(networkID, body.ContainerID); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "连接容器到网络失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"message": "容器已连接到网络"})
}

func DisconnectContainerHandler(c *gin.Context) {
	networkID := c.Param("id")
	var body struct {
		ContainerID string `json:"container_id"`
		Force       bool   `json:"force"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.ContainerID == "" {
		response.Error(c, http.StatusBadRequest, 400, "container_id 不能为空")
		return
	}

	if err := dockerService.DisconnectContainer(networkID, body.ContainerID, body.Force); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "断开容器网络失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"message": "容器已断开网络"})
}
