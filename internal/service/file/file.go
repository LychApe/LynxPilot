package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	IsDir   bool      `json:"is_dir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	Mode    string    `json:"mode"`
}

type ListResult struct {
	Path    string     `json:"path"`
	Parent  string     `json:"parent"`
	Entries []FileInfo `json:"entries"`
}

func safePath(basePath, reqPath string) (string, error) {
	clean := filepath.Clean(reqPath)
	if clean == "." {
		clean = ""
	}
	if strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("非法路径")
	}
	full := filepath.Join(basePath, clean)
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("解析基础路径失败: %w", err)
	}
	absFull, err := filepath.Abs(full)
	if err != nil {
		return "", fmt.Errorf("解析路径失败: %w", err)
	}
	if !strings.HasPrefix(absFull, absBase) {
		return "", fmt.Errorf("路径越界")
	}
	return absFull, nil
}

func ListFiles(basePath, reqPath string) (*ListResult, error) {
	absPath, err := safePath(basePath, reqPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("路径不存在或不可访问: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("不是目录")
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	relPath, _ := filepath.Rel(basePath, absPath)
	if relPath == "." {
		relPath = ""
	}

	parentRel := ""
	if relPath != "" {
		parentRel = filepath.Dir(relPath)
		if parentRel == "." {
			parentRel = ""
		}
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(relPath, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Mode:    info.Mode().String(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	return &ListResult{
		Path:    relPath,
		Parent:  parentRel,
		Entries: files,
	}, nil
}

func GetFileInfo(basePath, reqPath string) (*FileInfo, error) {
	absPath, err := safePath(basePath, reqPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}

	relPath, _ := filepath.Rel(basePath, absPath)
	if relPath == "." {
		relPath = ""
	}

	return &FileInfo{
		Name:    info.Name(),
		Path:    relPath,
		IsDir:   info.IsDir(),
		Size:    info.Size(),
		ModTime: info.ModTime(),
		Mode:    info.Mode().String(),
	}, nil
}

func ReadFileContent(basePath, reqPath string) (string, error) {
	absPath, err := safePath(basePath, reqPath)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("文件不存在: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("不能读取目录")
	}
	if info.Size() > 10*1024*1024 {
		return "", fmt.Errorf("文件过大（超过10MB），不支持在线预览")
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	return string(data), nil
}

func SaveFileContent(basePath, reqPath, content string) error {
	absPath, err := safePath(basePath, reqPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("文件不存在: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("不能写入目录")
	}

	return os.WriteFile(absPath, []byte(content), info.Mode())
}

func CreateDir(basePath, reqPath string) error {
	absPath, err := safePath(basePath, reqPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absPath); err == nil {
		return fmt.Errorf("路径已存在")
	}

	return os.MkdirAll(absPath, 0755)
}

func CreateFile(basePath, reqPath string) error {
	absPath, err := safePath(basePath, reqPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absPath); err == nil {
		return fmt.Errorf("文件已存在")
	}

	parent := filepath.Dir(absPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("创建父目录失败: %w", err)
	}

	f, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	return f.Close()
}

func Delete(basePath, reqPath string) error {
	absPath, err := safePath(basePath, reqPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("路径不存在: %w", err)
	}

	return os.RemoveAll(absPath)
}

func Rename(basePath, oldPath, newName string) error {
	absOld, err := safePath(basePath, oldPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absOld); err != nil {
		return fmt.Errorf("原路径不存在: %w", err)
	}

	if strings.Contains(newName, "/") || strings.Contains(newName, "\\") || strings.Contains(newName, "..") {
		return fmt.Errorf("新名称不能包含路径分隔符")
	}

	absNew := filepath.Join(filepath.Dir(absOld), newName)
	absBase, _ := filepath.Abs(basePath)
	absNewResolved, err := filepath.Abs(absNew)
	if err != nil {
		return fmt.Errorf("解析新路径失败: %w", err)
	}
	if !strings.HasPrefix(absNewResolved, absBase) {
		return fmt.Errorf("路径越界")
	}

	if _, err := os.Stat(absNewResolved); err == nil {
		return fmt.Errorf("目标名称已存在")
	}

	return os.Rename(absOld, absNewResolved)
}

func Upload(basePath, reqPath string, fileName string, reader io.Reader) error {
	absDir, err := safePath(basePath, reqPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return fmt.Errorf("目录不存在: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("目标不是目录")
	}

	if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") || strings.Contains(fileName, "..") {
		return fmt.Errorf("文件名非法")
	}

	absFile := filepath.Join(absDir, fileName)
	f, err := os.OpenFile(absFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

func Download(basePath, reqPath string) (string, error) {
	absPath, err := safePath(basePath, reqPath)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("文件不存在: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("不支持下载目录")
	}

	return absPath, nil
}
