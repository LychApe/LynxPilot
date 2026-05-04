package setting

import (
	"net/http"

	dockerService "github.com/LychApe/LynxPilot/internal/service/docker"
	"github.com/LychApe/LynxPilot/internal/service/setting"
	"github.com/LychApe/LynxPilot/internal/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetDockerConnectionHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	conn, err := setting.GetDockerConnection(db)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "获取 Docker 配置失败: "+err.Error())
		return
	}

	response.OK(c, conn)
}

func SaveDockerConnectionHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var conn setting.DockerConnection
	if err := c.ShouldBindJSON(&conn); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}

	if err := setting.SaveDockerConnection(db, &conn); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "保存 Docker 配置失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Docker 配置已保存"})
}

func TestDockerConnectionHandler(c *gin.Context) {
	var conn setting.DockerConnection
	if err := c.ShouldBindJSON(&conn); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}

	cfg := &dockerService.ConnectionConfig{
		Host:      conn.Host,
		TLSVerify: conn.TLSVerify,
		CertPath:  conn.CertPath,
	}

	if cfg.Host == "" {
		cfg = nil
	}

	if err := dockerService.TestConnection(cfg); err != nil {
		response.Error(c, http.StatusServiceUnavailable, 503, "连接失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "连接成功"})
}

func GetContainerDefaultsHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	defaults, err := setting.GetContainerDefaults(db)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "获取容器默认配置失败: "+err.Error())
		return
	}

	response.OK(c, defaults)
}

func SaveContainerDefaultsHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var defaults setting.ContainerDefaults
	if err := c.ShouldBindJSON(&defaults); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}

	if err := setting.SaveContainerDefaults(db, &defaults); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "保存容器默认配置失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "容器默认配置已保存"})
}

func GetUIPrefsHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	prefs, err := setting.GetContainerUIPrefs(db)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "获取界面偏好失败: "+err.Error())
		return
	}

	response.OK(c, prefs)
}

func SaveUIPrefsHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var prefs setting.ContainerUIPrefs
	if err := c.ShouldBindJSON(&prefs); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}

	if err := setting.SaveContainerUIPrefs(db, &prefs); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "保存界面偏好失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "界面偏好已保存"})
}

func GetAllSettingsHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	all, err := setting.GetAllSettings(db)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "获取设置失败: "+err.Error())
		return
	}

	response.OK(c, all)
}
