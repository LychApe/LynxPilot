package middleware

import (
	"net/http"
	"strings"

	jwtUtil "github.com/LychApe/LynxPilot/internal/utils/jwt"
	"github.com/LychApe/LynxPilot/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// Auth JWT 认证中间件，从 Authorization: Bearer <token> 中提取并验证令牌，
// 成功后将 userID 注入 context，失败返回 401
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, 40101, "未提供认证令牌")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")
		tokenSalt := c.MustGet("tokenSalt").(string)

		userID, err := jwtUtil.ParseToken(tokenString, tokenSalt)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, 40102, "认证令牌无效或已过期")
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
