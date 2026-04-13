package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterHealthRoutes(r gin.IRoutes, db *gorm.DB) {
	r.GET("/healthz", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "down", "error": err.Error()})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "down", "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "sqlite",
		})
	})
}
