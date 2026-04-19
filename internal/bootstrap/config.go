package bootstrap

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/LychApe/LynxPilot/internal/utils/logger"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Auth     AuthConfig     `yaml:"auth"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Port int    `yaml:"port"` // 服务端口
	Mode string `yaml:"mode"` // 服务模式
}

type AuthConfig struct {
	TokenSalt string `yaml:"token_salt"` // 认证密钥
}

type DatabaseConfig struct {
	Path string `yaml:"path"` // sqlite 数据库文件路径
}

// 加载配置
func LoadConfig(path string) (*Config, error) {
	candidatePaths := loadConfigBuildCandidatePaths(path)

	content, loadedPath, err := loadConfigReadByCandidates(candidatePaths)
	if err != nil {
		return nil, logger.Errorf("读取配置文件失败，已尝试路径%v: %v", candidatePaths, err)
	}

	var cfg Config
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, logger.Errorf("解析配置文件失败: %v", err)
	}

	if err := loadConfigValidate(&cfg); err != nil {
		return nil, err
	}

	logger.Infof("配置加载成功: %s", loadedPath)
	return &cfg, nil
}

// 构建候选路径
func loadConfigBuildCandidatePaths(path string) []string {
	if path == "" {
		path = "config/config.yaml"
	}

	if filepath.IsAbs(path) {
		return []string{filepath.Clean(path)}
	}

	paths := []string{
		filepath.Clean(path),
		filepath.Clean(filepath.Join("..", path)),
		filepath.Clean(filepath.Join("..", "..", path)),
		filepath.Clean(filepath.Join("..", "..", "..", path)),
	}

	unique := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		unique = append(unique, p)
	}

	return unique
}

// 读取配置文件
func loadConfigReadByCandidates(candidatePaths []string) ([]byte, string, error) {
	var lastErr error
	for _, candidate := range candidatePaths {
		content, err := os.ReadFile(candidate)
		if err == nil {
			return content, candidate, nil
		}
		lastErr = err
	}
	return nil, "", lastErr
}

// 验证配置
func loadConfigValidate(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return logger.Errorf("配置无效: server.port 必须在 1-65535 之间，当前值: %d", cfg.Server.Port)
	}

	mode := strings.TrimSpace(cfg.Server.Mode)
	if mode == "" {
		cfg.Server.Mode = "release"
	} else {
		cfg.Server.Mode = mode
	}

	if strings.TrimSpace(cfg.Auth.TokenSalt) == "" {
		return logger.Errorf("配置无效: auth.token_salt 不能为空")
	}

	dbPath := strings.TrimSpace(cfg.Database.Path)
	if dbPath == "" {
		cfg.Database.Path = "config/lynxpilot.db"
	} else {
		cfg.Database.Path = dbPath
	}

	return nil
}
