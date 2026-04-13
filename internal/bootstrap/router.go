package bootstrap

import (
	"github.com/LychApe/LynxPilot/internal/routes"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewGinEngine(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	routes.RegisterHealthRoutes(r, db)

	return r
}
