package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	userModel "github.com/LychApe/LynxPilot/internal/model/user"
	"github.com/LychApe/LynxPilot/internal/utils/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func LoadDatabase(config *Config) (*gorm.DB, error) {
	dbPath, err := loadDatabaseResolvePath(config.Database.Path)
	if err != nil {
		return nil, logger.Errorf("解析数据库路径失败: %v", err)
	}
	config.Database.Path = dbPath

	if err := loadDatabaseEnsureDir(dbPath); err != nil {
		return nil, logger.Errorf("创建数据库目录失败: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, logger.Errorf("初始化 sqlite 数据库失败: %v", err)
	}

	if err := db.AutoMigrate(&userModel.User{}); err != nil {
		return nil, logger.Errorf("自动迁移失败: %v", err)
	}

	DB = db
	logger.Infof("数据库初始化成功: %s", dbPath)
	return db, nil
}

func loadDatabaseEnsureDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func loadDatabaseResolvePath(dbPath string) (string, error) {
	if filepath.IsAbs(dbPath) {
		return filepath.Clean(dbPath), nil
	}

	repoRoot, err := loadDatabaseFindRepoRootFromCWD()
	if err != nil {
		return "", err
	}

	return filepath.Clean(filepath.Join(repoRoot, dbPath)), nil
}

func loadDatabaseFindRepoRootFromCWD() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("未找到项目根目录(go.mod)")
		}
		dir = parent
	}
}
