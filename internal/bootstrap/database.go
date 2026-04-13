package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/LychApe/LynxPilot/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ResolveSQLitePath() string {
	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		return "config/lynxpilot.db"
	}
	return dbPath
}

func NewGorm(dbPath string) (*gorm.DB, error) {
	dbDir := filepath.Dir(dbPath)
	if dbDir != "." {
		if err := os.MkdirAll(dbDir, 0o755); err != nil {
			return nil, err
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&model.HealthRecord{}); err != nil {
		sqlDB, sqlErr := db.DB()
		if sqlErr == nil {
			if closeErr := sqlDB.Close(); closeErr != nil {
				return nil, fmt.Errorf("auto migrate failed: %v; close db failed: %w", err, closeErr)
			}
		}
		return nil, err
	}

	return db, nil
}
