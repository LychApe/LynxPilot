package docker

import (
	"encoding/json"
	"net/http"

	"github.com/LychApe/LynxPilot/internal/model/setting"
	dockerService "github.com/LychApe/LynxPilot/internal/service/docker"
	"github.com/LychApe/LynxPilot/internal/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	if images == nil {
		images = []dockerService.ImageInfo{}
	}

	response.OK(c, images)
}

func PullImageHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var req struct {
		Image    string `json:"image"`
		Registry string `json:"registry"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}
	if req.Image == "" {
		response.Error(c, http.StatusBadRequest, 400, "镜像名称不能为空")
		return
	}

	var registryConfig *dockerService.RegistryConfig
	if req.Registry != "" {
		configs, err := getRegistries(db)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "读取仓库配置失败: "+err.Error())
			return
		}
		for _, cfg := range configs {
			if cfg.Name == req.Registry {
				registryConfig = &cfg
				break
			}
		}
	}

	if err := dockerService.PullImage(req.Image, registryConfig); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "镜像已拉取"})
}

func RemoveImageHandler(c *gin.Context) {
	imageID := c.Param("id")
	force := c.Query("force") == "true"
	if err := dockerService.RemoveImage(imageID, force); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "镜像已删除"})
}

func TagImageHandler(c *gin.Context) {
	var req dockerService.TagImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}
	if req.Source == "" || req.Target == "" {
		response.Error(c, http.StatusBadRequest, 400, "源镜像和目标标签不能为空")
		return
	}
	if err := dockerService.TagImage(req.Source, req.Target); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "镜像标签已创建"})
}

func PruneImagesHandler(c *gin.Context) {
	result, err := dockerService.PruneImages()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, result)
}

const dockerRegistriesKey = "docker_registries"

func getRegistries(db *gorm.DB) ([]dockerService.RegistryConfig, error) {
	var s setting.Setting
	if err := db.Where("`key` = ?", dockerRegistriesKey).Limit(1).Find(&s).Error; err != nil {
		return nil, err
	}
	if s.Value == "" {
		return []dockerService.RegistryConfig{}, nil
	}
	var configs []dockerService.RegistryConfig
	if err := json.Unmarshal([]byte(s.Value), &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

func ListRegistriesHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	configs, err := getRegistries(db)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "读取仓库配置失败: "+err.Error())
		return
	}
	response.OK(c, configs)
}

func SaveRegistriesHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var configs []dockerService.RegistryConfig
	if err := c.ShouldBindJSON(&configs); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}
	content, err := json.Marshal(configs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "序列化仓库配置失败: "+err.Error())
		return
	}
	if err := setting.Set(db, dockerRegistriesKey, string(content)); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "保存仓库配置失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"message": "仓库配置已保存"})
}

func TestRegistryHandler(c *gin.Context) {
	var config dockerService.RegistryConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}
	if config.ServerAddress == "" {
		response.Error(c, http.StatusBadRequest, 400, "仓库地址不能为空")
		return
	}
	if err := dockerService.TestRegistry(config); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "仓库登录成功"})
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
