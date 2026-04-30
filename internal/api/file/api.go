package file

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	fileService "github.com/LychApe/LynxPilot/internal/service/file"
	"github.com/LychApe/LynxPilot/internal/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/LychApe/LynxPilot/internal/model/setting"
)

const fileBasePathKey = "file_base_path"

func getBasePath(db *gorm.DB) string {
	var s setting.Setting
	db.Where("`key` = ?", fileBasePathKey).Limit(1).Find(&s)
	if s.Value == "" {
		return "/"
	}
	return s.Value
}

type listRequest struct {
	Path string `form:"path"`
}

type createDirRequest struct {
	Path string `json:"path"`
}

type createFileRequest struct {
	Path string `json:"path"`
}

type saveContentRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type renameRequest struct {
	Path    string `json:"path"`
	NewName string `json:"new_name"`
}

type deleteRequest struct {
	Path string `json:"path"`
}

func ListHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	basePath := getBasePath(db)

	var req listRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}

	result, err := fileService.ListFiles(basePath, req.Path)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "列出文件失败: "+err.Error())
		return
	}

	response.OK(c, result)
}

func GetFileInfoHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	basePath := getBasePath(db)

	reqPath := c.Query("path")
	info, err := fileService.GetFileInfo(basePath, reqPath)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "获取文件信息失败: "+err.Error())
		return
	}

	response.OK(c, info)
}

func ReadFileHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	basePath := getBasePath(db)

	reqPath := c.Query("path")
	content, err := fileService.ReadFileContent(basePath, reqPath)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "读取文件失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"content": content, "path": reqPath})
}

func SaveFileHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	basePath := getBasePath(db)

	var req saveContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}

	if req.Path == "" {
		response.Error(c, http.StatusBadRequest, 400, "路径不能为空")
		return
	}

	if err := fileService.SaveFileContent(basePath, req.Path, req.Content); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "保存文件失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "文件已保存"})
}

func CreateDirHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	basePath := getBasePath(db)

	var req createDirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}

	if req.Path == "" {
		response.Error(c, http.StatusBadRequest, 400, "路径不能为空")
		return
	}

	if err := fileService.CreateDir(basePath, req.Path); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "创建目录失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "目录已创建"})
}

func CreateFileHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	basePath := getBasePath(db)

	var req createFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}

	if req.Path == "" {
		response.Error(c, http.StatusBadRequest, 400, "路径不能为空")
		return
	}

	if err := fileService.CreateFile(basePath, req.Path); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "创建文件失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "文件已创建"})
}

func DeleteHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	basePath := getBasePath(db)

	var req deleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}

	if req.Path == "" {
		response.Error(c, http.StatusBadRequest, 400, "路径不能为空")
		return
	}

	if err := fileService.Delete(basePath, req.Path); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "删除失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "已删除"})
}

func RenameHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	basePath := getBasePath(db)

	var req renameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}

	if req.Path == "" || req.NewName == "" {
		response.Error(c, http.StatusBadRequest, 400, "路径和新名称不能为空")
		return
	}

	if err := fileService.Rename(basePath, req.Path, req.NewName); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "重命名失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "重命名成功"})
}

func UploadHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	basePath := getBasePath(db)

	reqPath := c.PostForm("path")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 400, "上传文件无效: "+err.Error())
		return
	}
	defer file.Close()

	if err := fileService.Upload(basePath, reqPath, header.Filename, file); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "上传失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "上传成功", "filename": header.Filename})
}

func DownloadHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	basePath := getBasePath(db)

	reqPath := c.Query("path")
	absPath, err := fileService.Download(basePath, reqPath)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "下载失败: "+err.Error())
		return
	}

	fileName := filepath.Base(absPath)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.File(absPath)
}

func GetBasePathHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	basePath := getBasePath(db)
	response.OK(c, gin.H{"base_path": basePath})
}

type setBasePathRequest struct {
	BasePath string `json:"base_path"`
}

func SetBasePathHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req setBasePathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "请求参数无效: "+err.Error())
		return
	}

	path := strings.TrimSpace(req.BasePath)
	if path == "" {
		response.Error(c, http.StatusBadRequest, 400, "基础路径不能为空")
		return
	}

	if err := setting.Set(db, fileBasePathKey, path); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, "保存基础路径失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "基础路径已保存"})
}
