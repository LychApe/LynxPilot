package user

import (
	apiUser "github.com/LychApe/LynxPilot/internal/api/user"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	userPublicGroup := router.Group("/api/v1/public/user")
	{
		//[注册]
		userPublicGroup.POST("/register", apiUser.RegisterHandler)
		//[登录]
		userPublicGroup.POST("/login", apiUser.LoginHandler)
	}
}
