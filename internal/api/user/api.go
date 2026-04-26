package user

import (
	"net/http"

	userService "github.com/LychApe/LynxPilot/internal/service/user"
	"github.com/LychApe/LynxPilot/internal/utils/response"
	jwtUtil "github.com/LychApe/LynxPilot/internal/utils/jwt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterHandler 用户注册，单用户系统仅允许创建一个用户
func RegisterHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"` // 前端传来的 MD5 值
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "用户名和密码不能为空")
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	user, err := userService.CreateUser(db, req.Username, req.Password, req.Email)
	if err != nil {
		response.Error(c, http.StatusConflict, 40901, err.Error())
		return
	}

	response.Created(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// LoginHandler 用户登录，验证账密后签发 JWT 令牌
func LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"` // 前端传来的 MD5 值
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "用户名和密码不能为空")
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	tokenSalt := c.MustGet("tokenSalt").(string)

	user, err := userService.Login(db, req.Username, req.Password)
	if err != nil {
		// 用户不存在和密码错误返回相同提示，防止枚举攻击
		response.Error(c, http.StatusUnauthorized, 40103, "用户名或密码错误")
		return
	}

	tokenString, expiresAt, err := jwtUtil.GenerateToken(user.ID, tokenSalt)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "生成令牌失败")
		return
	}

	response.OK(c, gin.H{
		"token":      tokenString,
		"expires_at": expiresAt,
	})
}
