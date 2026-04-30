package docker

import (
	"net/http"

	dockerService "github.com/LychApe/LynxPilot/internal/service/docker"
	"github.com/LychApe/LynxPilot/internal/utils/response"
	"github.com/gin-gonic/gin"
)

func ListVolumesHandler(c *gin.Context) {
	volumes, err := dockerService.ListVolumes()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	if volumes == nil {
		volumes = []dockerService.VolumeInfo{}
	}
	response.OK(c, volumes)
}

func CreateVolumeHandler(c *gin.Context) {
	var req dockerService.CreateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}
	if req.Name == "" {
		response.Error(c, http.StatusBadRequest, 400, "存储卷名称不能为空")
		return
	}
	vol, err := dockerService.CreateVolume(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, vol)
}

func RemoveVolumeHandler(c *gin.Context) {
	name := c.Param("name")
	force := c.Query("force") == "true"
	if err := dockerService.RemoveVolume(name, force); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "存储卷已删除"})
}

func PruneVolumesHandler(c *gin.Context) {
	result, err := dockerService.PruneVolumes()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, result)
}
